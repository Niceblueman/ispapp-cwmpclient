

#ifndef _ISPAPPCWMP_CONFIG_H__
#define _ISPAPPCWMP_CONFIG_H__

#include <uci.h>
#include <time.h>

#include "ispappcwmp.h"

void config_exit(void);
void config_load(void);
int config_remove_event(char *event);
int config_check_acs_url(void);

#ifdef BACKUP_DATA_IN_CONFIG
int ISPAPPCWMP_uci_init(void);
int ISPAPPCWMP_uci_fini(void);
char *ISPAPPCWMP_uci_get_value(char *package, char *section, char *option);
char *ISPAPPCWMP_uci_set_value(char *package, char *section, char *option, char *value);
int ISPAPPCWMP_uci_commit(void);
#endif

struct device {
	char *software_version;
};

struct acs {
	char *url;
	char *username;
	char *password;
	bool periodic_enable;
	bool http100continue_disable;
	int  periodic_interval;
	time_t periodic_time;
	char *ssl_cert;
	char *ssl_cacert;
	bool ssl_verify;
};

struct local {
	char *ip;
	char *interface;
	char *port;
	char *username;
	char *password;
	char *ubus_socket;
	int logging_level;
	int cr_auth_type;
};

struct core_config {
	struct device *device;
	struct acs *acs;
	struct local *local;
};

enum auth_type_enum {
	AUTH_BASIC,
	AUTH_DIGEST
};

#define DEFAULT_CR_AUTH_TYPE AUTH_DIGEST;


extern struct core_config *config;

#endif

