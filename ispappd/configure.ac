# Configure script for ispappd
AC_INIT([ispappd], [1.8.6], [support@ispapp.co])
AC_CONFIG_SRCDIR([src/ispappcwmp.c])
AC_CONFIG_AUX_DIR([build-aux])

# Initialize automake
AM_INIT_AUTOMAKE([1.11 -Wall -Werror subdir-objects foreign])
AM_SILENT_RULES([yes])

AC_ARG_ENABLE(jsonc, [AS_HELP_STRING([--enable-jsonc], [build with jsonc])], [
 AC_DEFINE(JSONC)
 LIBJSON_LIBS='-ljson-c'
 AC_SUBST([LIBJSON_LIBS])
], [
 LIBJSON_LIBS='-ljson'
 AC_SUBST([LIBJSON_LIBS])
])

AC_ARG_ENABLE(debug, [AS_HELP_STRING([--enable-debug], [enable debugging messages])], AC_DEFINE(DEBUG),)
AC_ARG_ENABLE(devel, [AS_HELP_STRING([--enable-devel], [enable development messages])], AC_DEFINE(DEVEL),)
AC_ARG_ENABLE(backupdatainconfig, [AS_HELP_STRING([--enable-backupdatainconfig], [enable backup data saved in the config])], AC_DEFINE(BACKUP_DATA_IN_CONFIG),)

# checks for programs
AC_PROG_CC
AM_PROG_CC_C_O

# checks for libraries

AC_ARG_WITH([uci-include-path],
  [AS_HELP_STRING([--with-uci-include-path],
    [location of the uci library headers])],
  [LIBUCI_CFLAGS="-I$withval"])
AC_SUBST([LIBUCI_CFLAGS])

AC_ARG_WITH([uci-lib-path],
  [AS_HELP_STRING([--with-uci-lib-path], [location of the uci library])], [LIBUCI_LDFLAGS="-L$withval"])
AC_SUBST([LIBUCI_LDFLAGS])

LIBUCI_LIBS='-luci'
AC_SUBST([LIBUCI_LIBS])

AC_ARG_WITH([libubox-include-path],
  [AS_HELP_STRING([--with-libubox-include-path],
    [location of the libubox library headers])],
  [LIBUBOX_CFLAGS="-I$withval"])
AC_SUBST([LIBUBOX_CFLAGS])

AC_ARG_WITH([libubox-lib-path],
  [AS_HELP_STRING([--with-libubox-lib-path], [location of the libubox library])], [LIBUBOX_LDFLAGS="-L$withval"])
AC_SUBST([LIBUBOX_LDFLAGS])

LIBUBOX_LIBS='-lubox'
AC_SUBST([LIBUBOX_LIBS])

AC_ARG_WITH([libubus-include-path],
  [AS_HELP_STRING([--with-libubus-include-path],
    [location of the libubus library headers])],
  [LIBUBUS_CFLAGS="-I$withval"])
AC_SUBST([LIBUBUS_CFLAGS])

AC_ARG_WITH([libubus-lib-path],
  [AS_HELP_STRING([--with-libubus-lib-path], [location of the libubus library])], [LIBUBUS_LDFLAGS="-L$withval"])
AC_SUBST([LIBUBUS_LDFLAGS])

LIBUBUS_LIBS='-lubus'
AC_SUBST([LIBUBUS_LIBS])

PKG_CHECK_MODULES(LIBCURL, [libcurl])
AC_SUBST(LIBCURL_CFLAGS)
AC_SUBST(LIBCURL_LDFLAGS)
AC_SUBST(LIBCURL_LIBS)

PKG_CHECK_MODULES(LIBXML2, [lxml2])
AC_SUBST(LIBXML2_CFLAGS)
AC_SUBST(LIBXML2_LDFLAGS)
AC_SUBST(LIBXML2_LIBS)

# checks for header files
AC_CHECK_HEADERS([stdlib.h string.h])

# checks for typedefs, structures, and compiler characteristics
AC_TYPE_UINT8_T

# Makefiles
AC_CONFIG_FILES([
Makefile
bin/Makefile
])

AC_OUTPUT

# Display configuration summary
echo ""
echo "Configuration Summary:"
echo "======================"
echo "Package:                $PACKAGE_NAME"
echo "Version:                $PACKAGE_VERSION"
echo "Compiler:               $CC"
echo "CFLAGS:                 $CFLAGS"
echo ""
echo "Features:"
echo "  JSON-C library:       $jsonc"
echo "  Debug messages:       $debug" 
echo "  Development messages: $devel"
echo "  Backup data in config: $backupdatainconfig"
echo ""
echo "Libraries:"
echo "  libcurl:              $LIBCURL_LIBS"
echo "  libroxml:             $LIBROXML_LIBS"
echo "  UCI:                  $LIBUCI_LIBS"
echo "  libubox:              $LIBUBOX_LIBS"
echo "  libubus:              $LIBUBUS_LIBS"
echo "  JSON:                 $LIBJSON_LIBS"
echo ""