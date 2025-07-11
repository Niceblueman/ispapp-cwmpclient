


#ifndef DIGESTAUTH_H_
#define DIGESTAUTH_H_

#define REALM "realm@ispappcwmp"
#define OPAQUE "ispappcwmp" // opaque@ispappcwmp

/**
 * MHD-internal return code for "YES".
 */
#define MHD_YES 1

/**
 * MHD-internal return code for "NO".
 */
#define MHD_NO 0

/**
 * MHD digest auth internal code for an invalid nonce.
 */
#define MHD_INVALID_NONCE -1

int http_digest_auth_fail_response(FILE *fp, const char *http_method,
		const char *url, const char *realm, const char *opaque);

int http_digest_auth_check(const char *http_method, const char *url,
		const char *header, const char *realm, const char *username,
		const char *password, unsigned int nonce_timeout);
void http_digest_init_nonce_priv_key(void);

#endif /* DIGESTAUTH_H_ */
