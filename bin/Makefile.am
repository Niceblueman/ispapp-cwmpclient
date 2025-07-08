bin_PROGRAMS = ispappd

ispappd_SOURCES =		\
	../src/backup.c		\
	../src/backup.h		\
	../src/config.h		\
	../src/config.c		\
	../src/cwmp.h		\
	../src/cwmp.c		\
	../src/digestauth.h	\
	../src/digestauth.c	\
	../src/external.h	\
	../src/external.c	\
	../src/ispappcwmp.h	\
	../src/ispappcwmp.c	\
	../src/http.h		\
	../src/http.c		\
	../src/json.c		\
	../src/json.h		\
	../src/log.c		\
	../src/log.h		\
	../src/md5.h		\
	../src/md5.c		\
	../src/time.h		\
	../src/time.c		\
	../src/ubus.h		\
	../src/ubus.c		\
	../src/xml.h		\
	../src/xml.c       	\
	../src/base64.h		\
	../src/base64.c		\
	../src/basicauth.h  	\
	../src/basicauth.c	\
	../src/command.h	\
	../src/command.c

ispappd_CFLAGS =		\
	$(AM_CFLAGS)		\
	$(LIBUCI_CFLAGS)	\
	$(LIBUBOX_CFLAGS)	\
	$(LIBUBUS_CFLAGS)	\
	$(MICROXML_CFLAGS)	\
	$(LIBCURL_CFLAGS)

ispappd_LDFLAGS =		\
	$(AM_LDFLAGS)		\
	$(LIBUCI_LDFLAGS)	\
	$(LIBUBOX_LDFLAGS)	\
	$(LIBUBUS_LDFLAGS)	\
	$(MICROXML_LDFLAGS)	\
	$(LIBCURL_LDFLAGS)

ispappd_LDADD =		\
	$(AM_LIBS)			\
	$(LIBUCI_LIBS)		\
	$(LIBUBOX_LIBS)		\
	$(LIBUBUS_LIBS)		\
	$(MICROXML_LIBS)	\
	$(LIBCURL_LIBS)		\
	$(LIBJSON_LIBS)

