

#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/wait.h>

#include <libubox/uloop.h>
#include <libubox/usock.h>
#include <curl/curl.h>

#include "http.h"
#include "config.h"
#include "cwmp.h"
#include "ispappcwmp.h"
#include "basicauth.h"
#include "digestauth.h"
#include "log.h"
#include "command.h"

static struct http_client http_c;
static struct http_server http_s;
CURL *curl;
char *http_redirect_url = NULL;

int
http_client_init(void)
{
	if (http_redirect_url) {
		if ((http_c.url = strdup(http_redirect_url)) == NULL)
			return -1;
	}
	else {
		if ((http_c.url = strdup(config->acs->url)) == NULL)
			return -1;
	}

	log_message(NAME, L_DEBUG, "+++ HTTP CLIENT CONFIGURATION +++\n");
	log_message(NAME, L_DEBUG, "url: %s\n", http_c.url);
	if (config->acs->ssl_cert)
		log_message(NAME, L_DEBUG, "ssl_cert: %s\n", config->acs->ssl_cert);
	if (config->acs->ssl_cacert)
		log_message(NAME, L_DEBUG, "ssl_cacert: %s\n", config->acs->ssl_cacert);
	if (!config->acs->ssl_verify)
		log_message(NAME, L_DEBUG, "ssl_verify: SSL certificate validation disabled.\n");
	log_message(NAME, L_DEBUG, "--- HTTP CLIENT CONFIGURATION ---\n");

	curl = curl_easy_init();
	if (!curl) return -1;
	curl_easy_setopt(curl, CURLOPT_URL, http_c.url);
	curl_easy_setopt(curl, CURLOPT_USERNAME, config->acs->username ? config->acs->username : "");
	curl_easy_setopt(curl, CURLOPT_PASSWORD, config->acs->password ? config->acs->password : "");
	curl_easy_setopt(curl, CURLOPT_HTTPAUTH, CURLAUTH_BASIC|CURLAUTH_DIGEST);
	curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, http_get_response);
	curl_easy_setopt(curl, CURLOPT_TIMEOUT, 30);
# ifdef DEVEL
	curl_easy_setopt(curl, CURLOPT_VERBOSE, 1L);
# endif /* DEVEL */
	curl_easy_setopt(curl, CURLOPT_COOKIEFILE, fc_cookies);
	curl_easy_setopt(curl, CURLOPT_COOKIEJAR, fc_cookies);
	if (config->acs->ssl_cert)
		curl_easy_setopt(curl, CURLOPT_SSLCERT, config->acs->ssl_cert);
	if (config->acs->ssl_cacert)
		curl_easy_setopt(curl, CURLOPT_CAINFO, config->acs->ssl_cacert);
	if (!config->acs->ssl_verify)
		curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0);

	log_message(NAME, L_NOTICE, "configured acs url %s\n", http_c.url);
	return 0;
}

void
http_client_exit(void)
{
	FREE(http_c.url);

	if(curl) {
	curl_easy_cleanup(curl);
		curl = NULL;
	}
	curl_global_cleanup();

	if(remove(fc_cookies) < 0)
		log_message(NAME, L_NOTICE, "can't remove file %s\n", fc_cookies);
}

static size_t
http_get_response(char *buffer, size_t size, size_t rxed, char **msg_in)
{
	char *c;

	if (asprintf(&c, "%s%.*s", *msg_in, size * rxed, buffer) == -1) {
		FREE(*msg_in);
		return -1;
	}

	free(*msg_in);
	*msg_in = c;

	return size * rxed;
}

int8_t
http_send_message(char *msg_out, char **msg_in)
{
	CURLcode res;
	char error_buf[CURL_ERROR_SIZE] = "";

	curl_easy_setopt(curl, CURLOPT_POSTFIELDS, msg_out);
	http_c.header_list = NULL;
	http_c.header_list = curl_slist_append(http_c.header_list, "Accept:");
	if (!http_c.header_list) return -1;
	http_c.header_list = curl_slist_append(http_c.header_list, "User-Agent: ispappcwmp");
	if (!http_c.header_list) return -1;
	http_c.header_list = curl_slist_append(http_c.header_list, "Content-Type: text/xml; charset=\"utf-8\"");
	if (!http_c.header_list) return -1;
	if (config->acs->http100continue_disable) {
		http_c.header_list = curl_slist_append(http_c.header_list, "Expect:");
		if (!http_c.header_list) return -1;
	}
	if (msg_out) {
		log_message(NAME, L_DEBUG, "+++ SEND HTTP REQUEST +++\n%s\n", msg_out);
		log_message(NAME, L_DEBUG, "--- SEND HTTP REQUEST ---\n");
		curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, (long) strlen(msg_out));
		http_c.header_list = curl_slist_append(http_c.header_list, "SOAPAction;");
		if (!http_c.header_list) return -1;
	}
	else {
		log_message(NAME, L_DEBUG, "+++ SEND EMPTY HTTP REQUEST +++\n");
		curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, 0);
	}
	curl_easy_setopt(curl, CURLOPT_FAILONERROR, true);
	curl_easy_setopt(curl, CURLOPT_ERRORBUFFER, error_buf);

	curl_easy_setopt(curl, CURLOPT_HTTPHEADER, http_c.header_list);

	curl_easy_setopt(curl, CURLOPT_WRITEDATA, msg_in);

	*msg_in = (char *) calloc (1, sizeof(char));

	res = curl_easy_perform(curl);

	if (http_c.header_list) {
		curl_slist_free_all(http_c.header_list);
		http_c.header_list = NULL;
	}

	if (error_buf[0] != '\0')
		log_message(NAME, L_NOTICE, "LibCurl Error: %s\n", error_buf);

	if (!strlen(*msg_in)) {
		FREE(*msg_in);
	}
	
	long httpCode = 0;
	curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpCode);

	if (httpCode == 302 || httpCode == 307) {
		curl_easy_getinfo(curl, CURLINFO_REDIRECT_URL, &http_redirect_url);
		if ((http_redirect_url = strdup(http_redirect_url)) == NULL)
			return -1;
		http_client_exit();
		if (http_client_init()) {
			log_message(NAME, L_DEBUG, "receiving http redirect: re-initializing http client failed\n");
			FREE(http_redirect_url);
			return -1;
		}
		FREE(http_redirect_url);
		FREE(*msg_in);
		int redirect = http_send_message(msg_out, msg_in);
		return redirect;
	}

	if (res || (httpCode != 200 && httpCode != 204)) {
		log_message(NAME, L_NOTICE, "sending http message failed\n");
		return -1;
	}

	if (*msg_in) {
		log_message(NAME, L_DEBUG, "+++ RECEIVED HTTP RESPONSE +++\n%s\n", *msg_in);
		log_message(NAME, L_DEBUG, "--- RECEIVED HTTP RESPONSE ---\n");
	} else {
		log_message(NAME, L_DEBUG, "+++ RECEIVED EMPTY HTTP RESPONSE +++\n");
	}

	return 0;
}

void
http_server_init(void)
{
	http_digest_init_nonce_priv_key();

	http_s.http_event.cb = http_new_client;

	http_s.http_event.fd = usock(USOCK_TCP | USOCK_SERVER | USOCK_NOCLOEXEC | USOCK_NONBLOCK, "0.0.0.0", config->local->port);
	uloop_fd_add(&http_s.http_event, ULOOP_READ | ULOOP_EDGE_TRIGGER);

	log_message(NAME, L_DEBUG, "+++ HTTP SERVER CONFIGURATION +++\n");
	if (config->local->ip)
		log_message(NAME, L_DEBUG, "ip: '%s'\n", config->local->ip);
	else
		log_message(NAME, L_DEBUG, "NOT BOUND TO IP\n");
	log_message(NAME, L_DEBUG, "port: '%s'\n", config->local->port);
	log_message(NAME, L_DEBUG, "--- HTTP SERVER CONFIGURATION ---\n");

	log_message(NAME, L_NOTICE, "http server initialized\n");
}

static void
http_new_client(struct uloop_fd *ufd, unsigned events)
{
	struct timeval t;
	int cr_auth_type = config->local->cr_auth_type;
	char buffer[BUFSIZ];
	char *auth_digest, *auth_basic;
	char *ispapp_command_header = NULL;
	int8_t auth_status = 0;
	int8_t command_request = 0;
	FILE *fp;
	int cnt = 0;

	t.tv_sec = 60;
	t.tv_usec = 0;

	for (;;) {
		int  client = -1, last_client = -1;
		while ((last_client = accept(ufd->fd, NULL, NULL)) >= 0) {
			if (client >= 0)
				close(client);
			client = last_client;
		}
		/* set one minute timeout */
		if (setsockopt(ufd->fd, SOL_SOCKET, SO_RCVTIMEO, (char *)&t, sizeof t)) {
			log_message(NAME, L_DEBUG, "setsockopt() failed\n");
		}
		if (client < 0) {
			break;
		}
		fp = fdopen(client, "r+");
		if (fp == NULL) {
			close(client);
			continue;
		}

		log_message(NAME, L_DEBUG, "+++ RECEIVED HTTP REQUEST +++\n");
		*buffer = '\0';
		while (fgets(buffer, sizeof(buffer), fp)) {
			char *username = config->local->username;
			char *password = config->local->password;
			
			// Check for X-ISPAPP-Command header
			if (strncasecmp(buffer, "X-ISPAPP-Command:", 17) == 0) {
				command_request = 1;
				char *header_value = buffer + 17;
				// Skip whitespace
				while (*header_value == ' ' || *header_value == '\t') {
					header_value++;
				}
				// Remove trailing newline/carriage return
				char *end = header_value + strlen(header_value) - 1;
				while (end > header_value && (*end == '\n' || *end == '\r')) {
					*end = '\0';
					end--;
				}
				ispapp_command_header = strdup(header_value);
				log_message(NAME, L_DEBUG, "X-ISPAPP-Command header found: %s\n", ispapp_command_header);
			}
			
			if (!username || !password) {
				// if we dont have username or password configured proceed with connecting to ACS
				auth_status = 1;
			}
			else if ((cr_auth_type == AUTH_DIGEST) && (auth_digest = strstr(buffer, "Authorization: Digest "))) {
				if (http_digest_auth_check("GET", "/", auth_digest + strlen("Authorization: Digest "), REALM, username, password, 300) == MHD_YES)
					auth_status = 1;
				else {
					auth_status = 0;
					log_message(NAME, L_NOTICE, "Connection Request authorization failed\n");
				}
			}
			else if ((cr_auth_type == AUTH_BASIC) && (auth_basic = strstr(buffer, "Authorization: Basic "))) {
				if (http_basic_auth_check(buffer ,username, password) == MHD_YES)
					auth_status = 1;
				else {
					auth_status = 0;
					log_message(NAME, L_NOTICE, "Connection Request authorization failed\n");
				}
			}
			if (buffer[0] == '\r' || buffer[0] == '\n') {
				/* end of http request (empty line) */
				goto http_end;
			}
		}

http_end:
		if (*buffer) {
			fflush(fp);
			if (auth_status) {
				if (command_request && ispapp_command_header) {
					// Handle command execution request
					log_message(NAME, L_NOTICE, "Processing ISPAPP command request\n");
					
					struct command_message *cmd_msg = command_parse_header(ispapp_command_header);
					if (cmd_msg) {
						struct command_result *result = command_execute_safe(cmd_msg);
						if (result) {
							char *json_response = command_result_to_json(result);
							if (json_response) {
								// Send HTTP response with JSON result
								fprintf(fp, "HTTP/1.1 200 OK\r\n");
								fprintf(fp, "Content-Type: application/json\r\n");
								fprintf(fp, "Content-Length: %zu\r\n", strlen(json_response));
								fprintf(fp, "Connection: close\r\n");
								fprintf(fp, "\r\n");
								fprintf(fp, "%s", json_response);
								
								log_message(NAME, L_DEBUG, "+++ ISPAPP COMMAND RESPONSE SENT +++\n");
								free(json_response);
							} else {
								fputs("HTTP/1.1 500 Internal Server Error\r\n", fp);
								fputs("Content-Length: 0\r\n", fp);
								fputs("Connection: close\r\n", fp);
								fputs("\r\n", fp);
							}
							command_result_free(result);
						} else {
							// Command execution failed
							const char *error_response = "{\"status\":\"error\",\"message\":\"Command execution failed\"}";
							fprintf(fp, "HTTP/1.1 400 Bad Request\r\n");
							fprintf(fp, "Content-Type: application/json\r\n");
							fprintf(fp, "Content-Length: %zu\r\n", strlen(error_response));
							fprintf(fp, "Connection: close\r\n");
							fprintf(fp, "\r\n");
							fprintf(fp, "%s", error_response);
						}
						command_message_free(cmd_msg);
					} else {
						// Invalid command format
						const char *error_response = "{\"status\":\"error\",\"message\":\"Invalid command format\"}";
						fprintf(fp, "HTTP/1.1 400 Bad Request\r\n");
						fprintf(fp, "Content-Type: application/json\r\n");
						fprintf(fp, "Content-Length: %zu\r\n", strlen(error_response));
						fprintf(fp, "Connection: close\r\n");
						fprintf(fp, "\r\n");
						fprintf(fp, "%s", error_response);
					}
				} else {
					// Standard CWMP connection request
					fputs("HTTP/1.1 200 OK\r\n", fp);
					fputs("Content-Length: 0\r\n", fp);
					fputs("Connection: close\r\n", fp);
					log_message(NAME, L_DEBUG, "+++ HTTP SERVER CONNECTION SUCCESS +++\n");
					log_message(NAME, L_NOTICE, "ACS initiated connection\n");
					cwmp_connection_request(EVENT_CONNECTION_REQUEST);
				}
			}
			else {
				fputs("HTTP/1.1 401 Unauthorized\r\n", fp);
				fputs("Content-Length: 0\r\n", fp);
				fputs("Connection: close\r\n", fp);
				if (cr_auth_type == AUTH_BASIC) {
					http_basic_auth_fail_response(fp, REALM);
				}
				else {
					http_digest_auth_fail_response(fp, "GET", "/", REALM, OPAQUE);
				}
				fputs("\r\n", fp);
			}
			fputs("\r\n", fp);
		}
		else {
			fputs("HTTP/1.1 409 Conflict\r\nConnection: close\r\n\r\n", fp);
		}
		
		// Cleanup
		if (ispapp_command_header) {
			free(ispapp_command_header);
		}
		
		fflush(fp);
		fclose(fp);
		close(client);
		log_message(NAME, L_DEBUG, "--- RECEIVED HTTP REQUEST ---\n");
		break;
	}
}
