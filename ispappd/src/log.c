

#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <syslog.h>
#include <time.h>

#include "log.h"
#include "ispappcwmp.h"
#include "config.h"

static const int log_class[] = {
	[L_CRIT] = LOG_CRIT,
	[L_WARNING] = LOG_WARNING,
	[L_NOTICE] = LOG_NOTICE,
	[L_INFO] = LOG_INFO,
	[L_DEBUG] = LOG_DEBUG
};

#ifdef DEBUG
static const char* log_str[] = {
	[L_CRIT] = "CRITICAL",
	[L_WARNING] = "WARNING",
	[L_NOTICE] = "NOTICE",
	[L_INFO] = "INFO",
	[L_DEBUG] = "DEBUG"
};
#endif

void log_message(char *name, int priority, const char *format, ...)
{
	va_list vl;

	if (!config || priority <= config->local->logging_level) {
#ifdef DEBUG
		time_t t = time(NULL);
		struct tm tm = *localtime(&t);
		va_start(vl, format);
		printf("%d-%02d-%02d %02d:%02d:%02d [ispappcwmp] %s - ", tm.tm_year + 1900, tm.tm_mon + 1, tm.tm_mday, tm.tm_hour, tm.tm_min, tm.tm_sec, log_str[priority]);
		vprintf(format, vl);
		va_end(vl);
#endif
		openlog(name, 0, LOG_DAEMON);
		va_start(vl, format);
		vsyslog(log_class[priority], format, vl);
		va_end(vl);
		closelog();
	}
}
