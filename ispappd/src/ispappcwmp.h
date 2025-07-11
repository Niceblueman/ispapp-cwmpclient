
#ifndef _ISPAPPCWMP_ISPAPPCWMP_H__
#define _ISPAPPCWMP_ISPAPPCWMP_H__
#ifndef CLOCK_MONOTONIC
#define CLOCK_MONOTONIC		1
#endif
#include <stdlib.h>
#include <stdlib.h>
#include <string.h>
#include <syslog.h>
#include <getopt.h>
#include <limits.h>
#include <locale.h>
#include <unistd.h>
#include <net/if.h>
#include <arpa/inet.h>
#include <linux/netlink.h>
#include <linux/rtnetlink.h>
#include <libubox/uloop.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/file.h>
#ifdef JSONC
 #include <json-c/json.h>
#else
 #include <json/json.h>
#endif
#include "ispappcwmp.h"
#include "config.h"
#include "cwmp.h"
#include "ubus.h"
#include "command.h"
#include "log.h"
#include "external.h"
#include "backup.h"
#include "http.h"
#include "xml.h"

#define NAME "ispappcwmp"
#define ISPAPPCWMP_VERSION "1.0.0"
#ifndef ARRAY_SIZE
#define ARRAY_SIZE(x) (sizeof(x) / sizeof(x[0]))
#define ARRAY_AND_SIZE(x) (x), ARRAY_SIZE(x)
#endif

#define FREE(x) do { free(x); x = NULL; } while (0);

#ifdef DEBUG
#define D(format, ...) fprintf(stderr, "%s(%d): " format, __func__, __LINE__, ## __VA_ARGS__)
#else
#define D(format, ...) no_debug(0, format, ## __VA_ARGS__)
#endif

#ifdef DEVEL
#define DD(format, ...) fprintf(stderr, "%s(%d):: " format, __func__, __LINE__, ## __VA_ARGS__)
#define DDF(format, ...) fprintf(stderr, format, ## __VA_ARGS__)
#else
#define DD(format, ...) no_debug(0, format, ## __VA_ARGS__)
#define DDF(format, ...) no_debug(0, format, ## __VA_ARGS__)
#endif

static inline void no_debug(int _, const char *fmt, ...)
{
}

enum start_event_enum {
	START_BOOT = 0x1,
	START_GET_RPC_METHOD = 0x2
};

void ISPAPPCWMP_reload(void);
void ISPAPPCWMP_notify(void);

#define TRACE(MESSAGE,args...) { \
  const char *A[] = {MESSAGE}; \
  printf("(TRACE: %s %s %d)  ",__FUNCTION__,__FILE__,__LINE__); \
  if(sizeof(A) > 0) \
	printf(*A,##args); \
  printf("%s\n", " "); \
  fflush(stdout); \
}

#endif

