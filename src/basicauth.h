

#ifndef BASICAUTH_H_
#define BASICAUTH_H_

#define REALM "realm@ispappcwmp"

#define MHD_YES 1

#define MHD_NO 0

int http_basic_auth_fail_response(FILE *fp, const char *realm);
int http_basic_auth_check(char buffer[BUFSIZ], char *username, char *password);

#endif /* BASICAUTH_H_ */
