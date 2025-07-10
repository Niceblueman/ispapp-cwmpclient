bin_PROGRAMS = ispappd

ispappd_SOURCES =		\
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
    ../src/easycwmp.c	\
    ../src/easycwmp.h	\
    ../src/external.c	\
    ../src/external.h	\
    ../src/http.c		\
    ../src/http.h		\
    ../src/ispappcwmp.c	\
    ../src/ispappcwmp.h	\
    ../src/json.c		\
    ../src/json.h		\
    ../src/log.c		\
    ../src/log.h		\
    ../src/md5.c		\
    ../src/md5.h		\
    ../src/messages.h	\
    ../src/time.c		\
    ../src/time.h		\
    ../src/ubus.c		\
    ../src/ubus.h		\
    ../src/xml.c		\
    ../src/xml.h		\
    ../src/xml_compat.h

ispappd_CFLAGS =		\
	$(AM_CFLAGS)		\
	$(LIBUCI_CFLAGS)	\
	$(LIBUBOX_CFLAGS)	\
	$(LIBUBUS_CFLAGS)	\
	$(LIBROXML_CFLAGS)	\
	$(LIBCURL_CFLAGS)

ispappd_LDFLAGS =		\
	$(AM_LDFLAGS)		\
	$(LIBUCI_LDFLAGS)	\
	$(LIBUBOX_LDFLAGS)	\
	$(LIBUBUS_LDFLAGS)	\
	$(LIBROXML_LDFLAGS)	\
	$(LIBCURL_LDFLAGS)

ispappd_LDADD =		\
	$(AM_LIBS)			\
	$(LIBUCI_LIBS)		\
	$(LIBUBOX_LIBS)		\
	$(LIBUBUS_LIBS)		\
	$(LIBROXML_LIBS)	\
	$(LIBCURL_LIBS)		\
	$(LIBJSON_LIBS)
