# Modified for ispappd

include $(TOPDIR)/rules.mk

PKG_NAME:=ispappd
PKG_VERSION:=1.0.0
PKG_RELEASE:=1

# Source is local (no download needed)
PKG_SOURCE:=
PKG_SOURCE_URL:=
PKG_HASH:=skip

PKG_FIXUP:=autoreconf

PKG_CONFIG_DEPENDS:= \
	CONFIG_ISPAPPD_DEBUG \
	CONFIG_ISPAPPD_DEVEL

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/ispappd
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=ISP App Daemon (CWMP client using libcurl)
  DEPENDS:=+libubus +libuci +libubox +libroxml +libjson-c +libcurl +curl
endef

define Package/ispappd/description
 An ISP application daemon based on CWMP (TR-069) protocol
endef

define Package/ispappd/config
	source "$(SOURCE)/Config.in"
endef

# Use local source files for feeds (files are in feed directory)  
define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) src $(PKG_BUILD_DIR)/
	$(CP) ext $(PKG_BUILD_DIR)/
	$(CP) configure.ac $(PKG_BUILD_DIR)/
	$(CP) Makefile.am $(PKG_BUILD_DIR)/
	$(CP) bin $(PKG_BUILD_DIR)/
endef

TARGET_CFLAGS += \
	-D_GNU_SOURCE

TARGET_LDFLAGS += \
	-Wl,-rpath-link=$(STAGING_DIR)/usr/lib

CONFIGURE_ARGS += \
	--with-uci-include-path=$(STAGING_DIR)/usr/include \
	--with-libubox-include-path=$(STAGING_DIR)/usr/include \
	--with-libubus-include-path=$(STAGING_DIR)/usr/include \
	--enable-jsonc=1

ifeq ($(CONFIG_ISPAPPD_DEBUG),y)
CONFIGURE_ARGS += \
	--enable-debug
endif

ifeq ($(CONFIG_ISPAPPD_DEVEL),y)
CONFIGURE_ARGS += \
	--enable-devel
endif

ifeq ($(CONFIG_ISPAPPD_BACKUP_DATA_CONFIG),y)
CONFIGURE_ARGS += \
	--enable-backupdatainconfig
endif

CONFIGURE_ARGS += \
	--enable-jsonc=1

define Package/ispappd/conffiles
/etc/config/ispappd
/usr/share/ispappd/defaults
endef

define Package/ispappd/install
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/bin/ispappd $(1)/usr/sbin
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) $(PKG_BUILD_DIR)/ext/openwrt/config/ispappd $(1)/etc/config
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/init.d/ispappd $(1)/etc/init.d
ifeq ($(ISPAPPD_BACKUP_DATA_FILE),y)
	$(INSTALL_DIR) $(1)/etc/ispappd
endif
ifeq ($(CONFIG_ISPAPPD_SCRIPTS_FULL),y)
	$(INSTALL_DIR) $(1)/usr/share/ispappd/functions/
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/defaults $(1)/usr/share/ispappd
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/common/* $(1)/usr/share/ispappd/functions/
ifeq ($(CONFIG_ISPAPPD_DATA_MODEL_TR181),y)
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/tr181/* $(1)/usr/share/ispappd/functions/
else
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/tr098/* $(1)/usr/share/ispappd/functions/
endif
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/scripts/ispappd.sh $(1)/usr/sbin/ispappd
	chmod +x $(1)/usr/share/ispappd/functions/*
else
	$(INSTALL_DIR) $(1)/usr/share/ispappd/functions/
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/common/ipping_launch $(1)/usr/share/ispappd/functions/ipping_launch
endif
endef

$(eval $(call BuildPackage,ispappd))
