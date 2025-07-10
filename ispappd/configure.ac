# Configure script for ispappd
AC_INIT([ispappd], [1.8.6], [support@ispapp.co])
AC_CONFIG_SRCDIR([src/ispappcwmp.c])
AC_CONFIG_AUX_DIR([build-aux])

# Initialize automake
AM_INIT_AUTOMAKE([1.11 -Wall -Werror subdir-objects foreign])
AM_SILENT_RULES([yes])

# Configure options
AC_ARG_ENABLE([jsonc],
    [AS_HELP_STRING([--enable-jsonc], [build with json-c library instead of json])],
    [case "${enableval}" in
        yes) jsonc=true ;;
        no)  jsonc=false ;;
        *) AC_MSG_ERROR([bad value ${enableval} for --enable-jsonc]) ;;
    esac],
    [jsonc=false])

AC_ARG_ENABLE([debug],
    [AS_HELP_STRING([--enable-debug], [enable debugging messages])],
    [case "${enableval}" in
        yes) debug=true ;;
        no)  debug=false ;;
        *) AC_MSG_ERROR([bad value ${enableval} for --enable-debug]) ;;
    esac],
    [debug=false])

AC_ARG_ENABLE([devel],
    [AS_HELP_STRING([--enable-devel], [enable development messages])],
    [case "${enableval}" in
        yes) devel=true ;;
        no)  devel=false ;;
        *) AC_MSG_ERROR([bad value ${enableval} for --enable-devel]) ;;
    esac],
    [devel=false])

AC_ARG_ENABLE([backupdatainconfig],
    [AS_HELP_STRING([--enable-backupdatainconfig], [enable backup data saved in the config])],
    [case "${enableval}" in
        yes) backupdatainconfig=true ;;
        no)  backupdatainconfig=false ;;
        *) AC_MSG_ERROR([bad value ${enableval} for --enable-backupdatainconfig]) ;;
    esac],
    [backupdatainconfig=false])

# Set up conditional compilation flags
AS_IF([test "x$jsonc" = "xtrue"], [
    AC_DEFINE([JSONC], [1], [Define to 1 if using json-c library])
    LIBJSON_LIBS='-ljson-c'
], [
    LIBJSON_LIBS='-ljson'
])
AC_SUBST([LIBJSON_LIBS])

AS_IF([test "x$debug" = "xtrue"], [
    AC_DEFINE([DEBUG], [1], [Define to 1 to enable debug messages])
])

AS_IF([test "x$devel" = "xtrue"], [
    AC_DEFINE([DEVEL], [1], [Define to 1 to enable development messages])
])

AS_IF([test "x$backupdatainconfig" = "xtrue"], [
    AC_DEFINE([BACKUP_DATA_IN_CONFIG], [1], [Define to 1 to enable backup data in config])
])

# Checks for programs
AC_PROG_CC
AC_PROG_INSTALL
AC_PROG_MAKE_SET
AM_PROG_CC_C_O

# Check for pkg-config
PKG_PROG_PKG_CONFIG

# Library path arguments for UCI
AC_ARG_WITH([uci-include-path],
    [AS_HELP_STRING([--with-uci-include-path=PATH],
        [location of the UCI library headers])],
    [LIBUCI_CFLAGS="-I$withval"],
    [LIBUCI_CFLAGS=""])
AC_SUBST([LIBUCI_CFLAGS])

AC_ARG_WITH([uci-lib-path],
    [AS_HELP_STRING([--with-uci-lib-path=PATH],
        [location of the UCI library])],
    [LIBUCI_LDFLAGS="-L$withval"],
    [LIBUCI_LDFLAGS=""])
AC_SUBST([LIBUCI_LDFLAGS])

LIBUCI_LIBS='-luci'
AC_SUBST([LIBUCI_LIBS])

# Library path arguments for libubox
AC_ARG_WITH([libubox-include-path],
    [AS_HELP_STRING([--with-libubox-include-path=PATH],
        [location of the libubox library headers])],
    [LIBUBOX_CFLAGS="-I$withval"],
    [LIBUBOX_CFLAGS=""])
AC_SUBST([LIBUBOX_CFLAGS])

AC_ARG_WITH([libubox-lib-path],
    [AS_HELP_STRING([--with-libubox-lib-path=PATH],
        [location of the libubox library])],
    [LIBUBOX_LDFLAGS="-L$withval"],
    [LIBUBOX_LDFLAGS=""])
AC_SUBST([LIBUBOX_LDFLAGS])

LIBUBOX_LIBS='-lubox'
AC_SUBST([LIBUBOX_LIBS])

# Library path arguments for libubus
AC_ARG_WITH([libubus-include-path],
    [AS_HELP_STRING([--with-libubus-include-path=PATH],
        [location of the libubus library headers])],
    [LIBUBUS_CFLAGS="-I$withval"],
    [LIBUBUS_CFLAGS=""])
AC_SUBST([LIBUBUS_CFLAGS])

AC_ARG_WITH([libubus-lib-path],
    [AS_HELP_STRING([--with-libubus-lib-path=PATH],
        [location of the libubus library])],
    [LIBUBUS_LDFLAGS="-L$withval"],
    [LIBUBUS_LDFLAGS=""])
AC_SUBST([LIBUBUS_LDFLAGS])

LIBUBUS_LIBS='-lubus'
AC_SUBST([LIBUBUS_LIBS])

# Check for required libraries using pkg-config
PKG_CHECK_MODULES([LIBCURL], [libcurl >= 7.20.0])

# Check for microxml/roxml library
PKG_CHECK_MODULES([MICROXML], [microxml],
    [have_microxml=yes],
    [have_microxml=no])

AS_IF([test "x$have_microxml" = "xno"], [
    AC_MSG_WARN([microxml not found via pkg-config, trying manual detection])
    # Fallback to manual library detection
    AC_CHECK_LIB([roxml], [roxml_load_doc],
        [MICROXML_LIBS="-lroxml"],
        [AC_MSG_ERROR([microxml/roxml library not found])])
    AC_CHECK_HEADER([roxml.h],
        [MICROXML_CFLAGS=""],
        [AC_MSG_ERROR([microxml/roxml headers not found])])
])

# Checks for header files
AC_CHECK_HEADERS([stdlib.h string.h unistd.h sys/types.h sys/socket.h netinet/in.h arpa/inet.h])

# Checks for typedefs, structures, and compiler characteristics
AC_TYPE_UINT8_T
AC_TYPE_SIZE_T
AC_TYPE_SSIZE_T
AC_C_CONST

# Checks for library functions
AC_FUNC_MALLOC
AC_FUNC_REALLOC
AC_CHECK_FUNCS([memset strchr strdup strerror])

# Output files
AC_CONFIG_FILES([
    Makefile
    bin/Makefile
])

# Generate the configuration
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
echo "  microxml:             $MICROXML_LIBS"
echo "  UCI:                  $LIBUCI_LIBS"
echo "  libubox:              $LIBUBOX_LIBS"
echo "  libubus:              $LIBUBUS_LIBS"
echo "  JSON:                 $LIBJSON_LIBS"
echo ""