
#ifndef _LOG_H__
#define _LOG_H__

#include <stdlib.h>
#define DEFAULT_LOGGING_LEVEL 3

enum {
	L_CRIT,
	L_WARNING,
	L_NOTICE,
	L_INFO,
	L_DEBUG
};

void log_message(char *name, int priority, const char *format, ...);

#endif
