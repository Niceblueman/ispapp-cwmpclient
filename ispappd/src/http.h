

#ifndef _ispappcwmp_HTTP_H__
#define _ispappcwmp_HTTP_H__

#include <stdint.h>

#include <libubox/uloop.h>
#include <curl/curl.h>

static char *fc_cookies = "/tmp/ispappcwmp_cookies";
struct http_client
{
	struct curl_slist *header_list;
	char *url;
};

struct http_server
{
	struct uloop_fd http_event;
};

static size_t http_get_response(char *buffer, size_t size, size_t rxed, char **msg_in);

int http_client_init(void);
void http_client_exit(void);
int8_t http_send_message(char *msg_out, char **msg_in);

void http_server_init(void);
static void http_new_client(struct uloop_fd *ufd, unsigned events);
static void http_del_client(struct uloop_process *uproc, int ret);

#endif

