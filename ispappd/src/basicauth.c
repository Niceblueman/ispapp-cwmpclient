

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <limits.h>
#include <errno.h>
#include <time.h>
#include <fcntl.h>
#include <unistd.h>
#include "ispappcwmp.h"
#include "basicauth.h"
#include "base64.h"
#include "log.h"

/**
 * Get the username and password from the basic authorization header sent by the client
 *
 * @param fp
 * @param password a pointer for the password
 * @param user a pointer for the user
 * @return MHD_NO if no username could be found
 * MHD_YES if username is found
 */

int http_basic_auth_get_username_password (char buffer[BUFSIZ], char **user, char **password)
{
	char header[BUFSIZ] = "";
	char *decode;
	const char *separator;
	if (strstr(buffer, "Authorization: Basic ") != NULL) {
		sscanf(buffer, "Authorization: Basic %s", header);
	}
	if ('\0' == *header)
		return MHD_NO;

	decode = BASE64Decode(header);
	if (NULL == decode) {
		return MHD_NO;
	}
  /* Find user:password pattern */
	if (NULL == (separator = strchr(decode, ':'))) {
		free(decode);
		return MHD_NO;
	}
	if (NULL == (*user = strdup(decode))) {
		free(decode);
		return MHD_NO;
	}
	(*user)[separator - decode] = '\0'; /* cut off at ':' */
	if (NULL != password) {
		*password = strdup(separator + 1);
		if (NULL == *password) {
			free(decode);
			free(*user);
			*user = NULL;
			return MHD_NO;
		}
	}
	free(decode);
	return MHD_YES;
}

int http_basic_auth_check(char buffer[BUFSIZ], char *username, char *password) {
	char *user = NULL, *pass = NULL;
	int ret;
	ret = http_basic_auth_get_username_password(buffer, &user, &pass);
	if ((ret == MHD_NO || strcmp(user, username) != 0)
			|| (strcmp(pass, password) != 0)) {
		log_message(NAME, L_DEBUG, "Authentication failed: username or password invalid \n");
		free(user);
		free(pass);
		return MHD_NO;
	}
	free(user);
	free(pass);
	return MHD_YES;
}

/**
 * make response to request authentication from the client.
 * @param fp
 * @param realm the realm presented to the client
 * @return #MHD_YES on success, #MHD_NO otherwise
 */

int http_basic_auth_fail_response(FILE *fp, const char *realm) {
	int ret;
	int res;
	int hlen = strlen(realm) + strlen("Basic realm=\"\"") + 1;
	char header[hlen];

	res = snprintf(header, hlen, "Basic realm=\"%s\"", realm);
	if (res > 0 && res < hlen) {
		log_message(NAME, L_DEBUG, "%s: header: %s", __FUNCTION__, header);
		fputs("WWW-Authenticate: ", fp);
		fputs(header, fp);
		return MHD_YES;
	}
	else {
		return MHD_NO;
	}
}

