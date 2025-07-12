bin_PROGRAMS = ispappcwmp

# Object file destination configuration
AUTOMAKE_OPTIONS = subdir-objects
AM_CPPFLAGS = -I$(top_srcdir)/src

ispappcwmp_SOURCES =		\
    ../src/backup.c		\
    ../src/backup.h		\
    ../src/backup_stubs.h	\
    ../src/base64.c		\
    ../src/base64.h		\
    ../src/basicauth.c	\
    ../src/basicauth.h	\
    ../src/command.c	\
    ../src/command.h	\
    ../src/config.c		\
    ../src/config.h		\
    ../src/cwmp.c		\
    ../src/cwmp.h		\
    ../src/digestauth.c	\
    ../src/digestauth.h	\
    ../src/ispappcwmp.c	\
    ../src/ispappcwmp.h	\
    ../src/external.c	\
    ../src/external.h	\
    ../src/http.c		\
    ../src/http.h		\
    ../src/json.c		\
    ../src/json.h		\
    ../src/log.c		\
    ../src/log.h		\
    ../src/md5.c		\
    ../src/md5.h		\
    ../src/messages.h	\
    ../src/dtime.c		\
    ../src/dtime.h		\
    ../src/ubus.c		\
    ../src/ubus.h		\
    ../src/libxml_helpers.h \
    ../src/libxml_helpers.c \
    ../src/xml.c		\
    ../src/xml.h 

ispappcwmp_CFLAGS =		\
	$(AM_CFLAGS)		\
	$(LIBUCI_CFLAGS)	\
	$(LIBUBOX_CFLAGS)	\
	$(LIBUBUS_CFLAGS)	\
	$(LIBXML2_CFLAGS)	\
	$(LIBCURL_CFLAGS)

ispappcwmp_LDFLAGS =		\
	$(AM_LDFLAGS)		\
	$(LIBUCI_LDFLAGS)	\
	$(LIBUBOX_LDFLAGS)	\
	$(LIBUBUS_LDFLAGS)	\
	$(LIBXML2_LDFLAGS)	\
	$(LIBCURL_LDFLAGS)

ispappcwmp_LDADD =		\
	$(AM_LIBS)			\
	$(LIBUCI_LIBS)		\
	$(LIBUBOX_LIBS)		\
	$(LIBUBUS_LIBS)		\
	$(LIBXML2_LIBS)	\
	$(LIBCURL_LIBS)		\
	$(LIBJSON_LIBS)
