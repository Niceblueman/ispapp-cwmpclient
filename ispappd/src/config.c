#include "cwmp.h"
#include "backup.h" 
#include "config.h"
#include "log.h"
#include "ispappcwmp.h"

typedef xmlNodePtr xml_node_t;
static bool first_run = true;
static struct uci_context *uci_ctx;
static struct uci_package *uci_ispappcwmp;
#ifdef BACKUP_DATA_IN_CONFIG
static struct uci_context *ISPAPPCWMP_uci_ctx = NULL;
#endif

struct core_config *config;


static void config_free_local(void) {
	if (config->local) {
		FREE(config->local->ip);
		FREE(config->local->interface);
		FREE(config->local->port);
		FREE(config->local->username);
		FREE(config->local->password);
		FREE(config->local->ubus_socket);
	}
}

static int config_init_local(void)
{
	struct uci_section *s;
	struct uci_element *e1;

	uci_foreach_element(&uci_ispappcwmp->sections, e1) {
		s = uci_to_section(e1);
		if (strcmp(s->type, "local") == 0) {
			config_free_local();

			config->local->logging_level = DEFAULT_LOGGING_LEVEL;
			config->local->cr_auth_type = DEFAULT_CR_AUTH_TYPE;
			uci_foreach_element(&s->options, e1) {
				if (!strcmp((uci_to_option(e1))->e.name, "interface")) {
					config->local->interface = strdup(uci_to_option(e1)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@local[0].interface=%s\n", config->local->interface);
					continue;
				}

				if (!strcmp((uci_to_option(e1))->e.name, "port")) {
					if (!atoi((uci_to_option(e1))->v.string)) {
						log_message(NAME, L_DEBUG, "in section local port has invalid value...\n");
						return -1;
					}
					config->local->port = strdup(uci_to_option(e1)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@local[0].port=%s\n", config->local->port);
					continue;
				}

				if (!strcmp((uci_to_option(e1))->e.name, "username")) {
					config->local->username = strdup(uci_to_option(e1)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@local[0].username=%s\n", config->local->username);
					continue;
				}

				if (!strcmp((uci_to_option(e1))->e.name, "password")) {
					config->local->password = strdup(uci_to_option(e1)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@local[0].password=%s\n", config->local->password);
					continue;
				}

				if (!strcmp((uci_to_option(e1))->e.name, "ubus_socket")) {
					config->local->ubus_socket = strdup(uci_to_option(e1)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@local[0].ubus_socket=%s\n", config->local->ubus_socket);
					continue;
				}
				
				if (!strcmp((uci_to_option(e1))->e.name, "logging_level")) {
					char *c;
					int log_level = atoi((uci_to_option(e1))->v.string);					 
					if(asprintf(&c, "%d", log_level) != -1)
					{	
						if (strcmp(c, uci_to_option(e1)->v.string) == 0) 
							config->local->logging_level = log_level;						
						free(c);
					}
					log_message(NAME, L_DEBUG, "ispappd.@local[0].logging_level=%d\n", config->local->logging_level);
					continue;
				}
				
				if (!strcmp((uci_to_option(e1))->e.name, "authentication")) {
					if (strcasecmp((uci_to_option(e1))->v.string, "Basic") == 0)
						config->local->cr_auth_type = AUTH_BASIC;
					else
						config->local->cr_auth_type = AUTH_DIGEST;				 

					log_message(NAME, L_DEBUG, "ispappd.@local[0].authentication=%s\n",
						(config->local->cr_auth_type == AUTH_BASIC) ? "Basic" : "Digest");
					continue;
				}
			}

			if (!config->local->interface) {
				log_message(NAME, L_DEBUG, "in local you must define interface\n");
				return -1;
			}

			if (!config->local->port) {
				log_message(NAME, L_DEBUG, "in local you must define port\n");
				return -1;
			}

			if (!config->local->ubus_socket) {
				log_message(NAME, L_DEBUG, "in local you must define ubus_socket\n");
				return -1;
			}

			return 0;
		}
	}
	log_message(NAME, L_DEBUG, "uci section local not found...\n");
	return -1;
}

static void config_free_acs(void) {
	if (config->acs) {
		FREE(config->acs->url);
		FREE(config->acs->username);
		FREE(config->acs->password);
		FREE(config->acs->ssl_cert);
		FREE(config->acs->ssl_cacert);
	}
}

static int config_init_acs(void)
{
	struct uci_section *s;
	struct uci_element *e;
	struct tm tm;

	uci_foreach_element(&uci_ispappcwmp->sections, e) {
		s = uci_to_section(e);
		if (strcmp(s->type, "acs") == 0) {
			config_free_acs();
			config->acs->periodic_time = -1;

			uci_foreach_element(&s->options, e) {
				if (!strcmp((uci_to_option(e))->e.name, "url")) {
					bool valid = false;

					if (!(strncmp((uci_to_option(e))->v.string, "http:", 5)))
						valid = true;

					if (!(strncmp((uci_to_option(e))->v.string, "https:", 6)))
						valid = true;

					if (!valid) {
						log_message(NAME, L_DEBUG, "in section acs scheme must be either http or https...\n");
						return -1;
					}

					config->acs->url = strdup(uci_to_option(e)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].url=%s\n", config->acs->url);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "username")) {
					config->acs->username = strdup(uci_to_option(e)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].username=%s\n", config->acs->username);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "password")) {
					config->acs->password = strdup(uci_to_option(e)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].password=%s\n", config->acs->password);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "periodic_enable")) {
					config->acs->periodic_enable = (atoi((uci_to_option(e))->v.string) == 1) ? true : false;
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].periodic_enable=%d\n", config->acs->periodic_enable);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "periodic_interval")) {
					config->acs->periodic_interval = atoi((uci_to_option(e))->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].periodic_interval=%d\n", config->acs->periodic_interval);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "periodic_time")) {
					const char *timestr = uci_to_option(e)->v.string;
					struct tm tm = {0};
					time_t t = -1;
					size_t len = strlen(timestr);
					int parse_ok = 0;
					if (len > 0 && timestr[len-1] == 'Z') {
						char buf[32];
						strncpy(buf, timestr, sizeof(buf)-1);
						buf[sizeof(buf)-1] = 0;
						buf[len-1] = 0; // Remove 'Z'
						if (strptime(buf, "%Y-%m-%dT%H:%M:%S", &tm)) {
#if defined(_GNU_SOURCE) || defined(__USE_BSD)
							t = timegm(&tm); // Use UTC
#else
							t = mktime(&tm); // Fallback: localtime
							log_message(NAME, L_WARNING, "timegm() not available, falling back to localtime for periodic_time\n");
#endif
							parse_ok = 1;
						}
					} else {
						if (strptime(timestr, "%Y-%m-%dT%H:%M:%S", &tm)) {
							t = mktime(&tm); // Use local time
							parse_ok = 1;
						}
					}
					if (!parse_ok || t < 0) {
						log_message(NAME, L_WARNING, "Failed to parse periodic_time '%s', using current time\n", timestr);
						t = time(NULL);
					}
					config->acs->periodic_time = t;
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].periodic_time=%s (epoch=%ld)\n", timestr, (long)t);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "http100continue_disable")) {
					config->acs->http100continue_disable = (atoi(uci_to_option(e)->v.string)) ? true : false;
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].http100continue_disable=%d\n", config->acs->http100continue_disable);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "ssl_cert")) {
					config->acs->ssl_cert = strdup(uci_to_option(e)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].ssl_cert=%s\n", config->acs->ssl_cert);
					continue;
				}
				if (!strcmp((uci_to_option(e))->e.name, "ssl_cacert")) {
					config->acs->ssl_cacert = strdup(uci_to_option(e)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].ssl_cacert=%s\n", config->acs->ssl_cacert);
					continue;
				}

				if (!strcmp((uci_to_option(e))->e.name, "ssl_verify")) {
					if (!strcmp((uci_to_option(e))->v.string, "enabled")) {
						config->acs->ssl_verify = true;
					} else {
						config->acs->ssl_verify = false;
					}
					log_message(NAME, L_DEBUG, "ispappd.@acs[0].ssl_verify=%d\n", config->acs->ssl_verify);
					continue;
				}
			}

			if (!config->acs->url) {
				log_message(NAME, L_DEBUG, "acs url must be defined in the config\n");
				return -1;
			}

			return 0;
		}
	}
	log_message(NAME, L_DEBUG, "uci section acs not found...\n");
	return -1;
}

static void config_free_device(void) {
	if (config->device) {
		FREE(config->device->software_version);
	}
}

static int config_init_device(void)
{
	struct uci_section *s;
	struct uci_element *e2;

	uci_foreach_element(&uci_ispappcwmp->sections, e2) {
		s = uci_to_section(e2);
		if (strcmp(s->type, "device") == 0) {
			config_free_device();
			uci_foreach_element(&s->options, e2) {
				if (!strcmp((uci_to_option(e2))->e.name, "software_version")) {
					config->device->software_version = strdup(uci_to_option(e2)->v.string);
					log_message(NAME, L_DEBUG, "ispappd.@device[0].software_version=%s\n", config->device->software_version);
					continue;
				}
			}
			return 0;
		}
	}
	log_message(NAME, L_DEBUG, "uci section device not found...\n");
	return -1;
}
static struct uci_package *
config_init_package(const char *c)
{
	if (first_run) {
		config = calloc(1, sizeof(struct core_config));
		if (!config) goto error;

		config->acs = calloc(1, sizeof(struct acs));
		if (!config->acs) goto error;

		config->local = calloc(1, sizeof(struct local));
		if (!config->local) goto error;
		config->device = calloc(1, sizeof(struct device));
		if (!config->device) goto error;
	}
	if (!uci_ctx) {
		uci_ctx = uci_alloc_context();
		if (!uci_ctx) goto error;
	} else {
		if (uci_ispappcwmp) {
			uci_unload(uci_ctx, uci_ispappcwmp);
			uci_ispappcwmp = NULL;
		}
	}
	if (uci_load(uci_ctx, c, &uci_ispappcwmp)) {
		uci_free_context(uci_ctx);
		uci_ctx = NULL;
		return NULL;
	}
	return uci_ispappcwmp;

error:
	config_exit();
	return NULL;
}

static inline void config_free_ctx(void)
{
	if (uci_ctx) {
		if (uci_ispappcwmp) {
			uci_unload(uci_ctx, uci_ispappcwmp);
			uci_ispappcwmp = NULL;
		}
		uci_free_context(uci_ctx);
		uci_ctx = NULL;
	}
}

void config_exit(void)
{
	if (config) {
		config_free_acs();
		FREE(config->acs);
		config_free_local();
		FREE(config->local);
		config_free_device();
		FREE(config->device);
		FREE(config);
	}
	config_free_ctx();
}

void config_load(void)
{

	uci_ispappcwmp = config_init_package("ispappd");

	if (!uci_ispappcwmp) goto error;
	if (config_init_device()) goto error;
	if (config_init_local()) goto error;
	if (config_init_acs()) goto error;

	backup_check_acs_url();
	backup_check_software_version();
	cwmp_periodic_inform_init();

	first_run = false;
	config_free_ctx();

	cwmp_update_value_change();
	return;

error:
	log_message(NAME, L_CRIT, "configuration (re)loading failed, exit daemon\n");
	exit(EXIT_FAILURE);
}

#ifdef BACKUP_DATA_IN_CONFIG
int ISPAPPCWMP_uci_init(void)
{
	ISPAPPCWMP_uci_ctx = uci_alloc_context();
	if (!ISPAPPCWMP_uci_ctx) {
		return -1;
	}
	return 0;
}

int ISPAPPCWMP_uci_fini(void)
{
	if (ISPAPPCWMP_uci_ctx) {
		uci_free_context(ISPAPPCWMP_uci_ctx);
	}
	return 0;
}

static bool ISPAPPCWMP_uci_validate_section(const char *str)
{
	if (!*str)
		return false;

	for (; *str; str++) {
		unsigned char c = *str;

		if (isalnum(c) || c == '_')
			continue;

		return false;
	}
	return true;
}

int ISPAPPCWMP_uci_init_ptr(struct uci_context *ctx, struct uci_ptr *ptr, char *package, char *section, char *option, char *value)
{
	char *last = NULL;
	char *tmp;

	memset(ptr, 0, sizeof(struct uci_ptr));

	/* value */
	if (value) {
		ptr->value = value;
	}
	ptr->package = package;
	if (!ptr->package)
		goto error;

	ptr->section = section;
	if (!ptr->section) {
		ptr->target = UCI_TYPE_PACKAGE;
		goto lastval;
	}

	ptr->option = option;
	if (!ptr->option) {
		ptr->target = UCI_TYPE_SECTION;
		goto lastval;
	} else {
		ptr->target = UCI_TYPE_OPTION;
	}

lastval:
	if (ptr->section && !ISPAPPCWMP_uci_validate_section(ptr->section))
		ptr->flags |= UCI_LOOKUP_EXTENDED;

	return 0;

error:
	return -1;
}

char *ISPAPPCWMP_uci_get_value(char *package, char *section, char *option)
{
	struct uci_ptr ptr;
	char *val = "";

	if (!section || !option)
		return val;

	if (ISPAPPCWMP_uci_init_ptr(ISPAPPCWMP_uci_ctx, &ptr, package, section, option, NULL)) {
		return val;
	}
	if (uci_lookup_ptr(ISPAPPCWMP_uci_ctx, &ptr, NULL, true) != UCI_OK) {
		return val;
	}

	if (!ptr.o)
		return val;

	if (ptr.o->v.string)
		return ptr.o->v.string;
	else
		return val;
}

char *ISPAPPCWMP_uci_set_value(char *package, char *section, char *option, char *value)
{
	struct uci_ptr ptr;
	int ret = UCI_OK;

	if (!section)
		return "";

	if (ISPAPPCWMP_uci_init_ptr(ISPAPPCWMP_uci_ctx, &ptr, package, section, option, value)) {
		return "";
	}
	if (uci_lookup_ptr(ISPAPPCWMP_uci_ctx, &ptr, NULL, true) != UCI_OK) {
		return "";
	}

	uci_set(ISPAPPCWMP_uci_ctx, &ptr);

	if (ret == UCI_OK)
		ret = uci_save(ISPAPPCWMP_uci_ctx, ptr.p);

	if (ptr.o && ptr.o->v.string)
		return ptr.o->v.string;

	return "";
}

int ISPAPPCWMP_uci_commit(void)
{
	struct uci_element *e;
	struct uci_context *ctx;
	struct uci_ptr ptr;

	ctx = uci_alloc_context();
	if (!ctx) {
		return -1;
	}

	uci_foreach_element(&ISPAPPCWMP_uci_ctx->root, e) {
		if (ISPAPPCWMP_uci_init_ptr(ctx, &ptr, e->name, NULL, NULL, NULL)) {
			return -1;
		}
		if (uci_lookup_ptr(ctx, &ptr, NULL, true) != UCI_OK) {
			return -1;
		}
		uci_commit(ctx, &ptr.p, false);
	}

	uci_free_context(ctx);

	return 0;
}
#endif
