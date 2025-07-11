bin_PROGRAMS = ispappd

# Define build directory for object files
builddir = $(top_builddir)/build-output
objdir = $(builddir)/obj
bindir = $(builddir)/bin

# Ensure build directories exist
$(builddir):
	@$(MKDIR_P) $(builddir)

$(objdir): | $(builddir)
	@$(MKDIR_P) $(objdir)

$(bindir): | $(builddir)
	@$(MKDIR_P) $(bindir)

# Object file destination configuration
AUTOMAKE_OPTIONS = subdir-objects
AM_CPPFLAGS = -I$(top_srcdir)/src

# Create object files in the build directory
ispappd_OBJECTS = \
    $(objdir)/backup.o \
    $(objdir)/base64.o \
    $(objdir)/basicauth.o \
    $(objdir)/command.o \
    $(objdir)/config.o \
    $(objdir)/cwmp.o \
    $(objdir)/digestauth.o \
    $(objdir)/ispappcwmp.o \
    $(objdir)/external.o \
    $(objdir)/http.o \
    $(objdir)/json.o \
    $(objdir)/log.o \
    $(objdir)/md5.o \
    $(objdir)/time.o \
    $(objdir)/ubus.o \
    $(objdir)/xml.o \
    $(objdir)/libxml_helpers.o

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
    ../src/time.c		\
    ../src/time.h		\
    ../src/ubus.c		\
    ../src/ubus.h		\
    ../src/xml.c		\
    ../src/xml.h		\
    ../src/libxml_helpers.h \
    ../src/libxml_helpers.c

# Custom rules for object file destinations
$(objdir)/%.o: $(top_srcdir)/src/%.c | $(objdir)
	@echo "  CC      $@"
	@$(CC) $(DEFS) $(DEFAULT_INCLUDES) $(INCLUDES) $(AM_CPPFLAGS) $(CPPFLAGS) $(ispappd_CFLAGS) $(CFLAGS) -c -o $@ $<

# Override default build rule to use our object directory
ispappd$(EXEEXT): $(ispappd_OBJECTS) | $(bindir)
	@echo "  CCLD    $@"
	@$(CC) $(ispappd_OBJECTS) -o $(bindir)/$@ $(ispappd_LDFLAGS) $(LDFLAGS) $(ispappd_LDADD) $(LIBS)
	@cp $(bindir)/$@ .

# Clean rule to remove build directory
clean-local:
	@rm -rf $(builddir)

# Install rule
install-exec-local:
	@$(MKDIR_P) $(DESTDIR)$(bindir)
	@$(INSTALL_PROGRAM) $(bindir)/ispappd$(EXEEXT) $(DESTDIR)$(bindir)/ispappd$(EXEEXT) 

ispappd_CFLAGS =		\
	$(AM_CFLAGS)		\
	$(LIBUCI_CFLAGS)	\
	$(LIBUBOX_CFLAGS)	\
	$(LIBUBUS_CFLAGS)	\
	$(LIBXML2_CFLAGS)	\
	$(LIBCURL_CFLAGS)

ispappd_LDFLAGS =		\
	$(AM_LDFLAGS)		\
	$(LIBUCI_LDFLAGS)	\
	$(LIBUBOX_LDFLAGS)	\
	$(LIBUBUS_LDFLAGS)	\
	$(LIXML2_LDFLAGS)	\
	$(LIBCURL_LDFLAGS)

ispappd_LDADD =		\
	$(AM_LIBS)			\
	$(LIBUCI_LIBS)		\
	$(LIBUBOX_LIBS)		\
	$(LIBUBUS_LIBS)		\
	$(LIXML2_LIBS)	\
	$(LIBCURL_LIBS)		\
	$(LIBJSON_LIBS)
