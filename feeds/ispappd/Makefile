# Copyright (C) 2012-2016 PIVA Software <www.pivasoftware.com>
# 	Author: MOHAMED Kallel <mohamed.kallel@pivasoftware.com>
# 	Author: ANIS ELLOUZE <anis.ellouze@pivasoftware.com>

include $(TOPDIR)/rules.mk

PKG_NAME:=ispapp
PKG_VERSION:=1.0.0
PKG_SOURCE:=$(PKG_NAME)-$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/

PKG_FIXUP:=autoreconf

PKG_CONFIG_DEPENDS:= \
	CONFIG_ISPAPPCWMP_DEBUG \
	CONFIG_ISPAPPCWMP_DEVEL

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/ispapp
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=CWMP client (using libcurl)
  DEPENDS:=+libubus +libuci +libubox +libmicroxml +libjson-c +libcurl +curl
endef

define Package/ispapp/description
 An open source implementation of the client CWMP (TR-069) protocol
endef

define Package/ispapp/config
	source "$(SOURCE)/Config.in"
endef

USE_LOCAL=$(shell ls ./src/ 2>/dev/null >/dev/null && echo 1)
ifneq ($(USE_LOCAL),)
define Build/Prepare
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef
endif

TARGET_CFLAGS += \
	-D_GNU_SOURCE

TARGET_LDFLAGS += \
	-Wl,-rpath-link=$(STAGING_DIR)/usr/lib

CONFIGURE_ARGS += \
	--with-uci-include-path=$(STAGING_DIR)/usr/include \
	--with-libubox-include-path=$(STAGING_DIR)/usr/include \
	--with-libubus-include-path=$(STAGING_DIR)/usr/include

ifeq ($(CONFIG_ISPAPPCWMP_DEBUG),y)
CONFIGURE_ARGS += \
	--enable-debug
endif

ifeq ($(CONFIG_ISPAPPCWMP_DEVEL),y)
CONFIGURE_ARGS += \
	--enable-devel
endif

ifeq ($(CONFIG_ISPAPPCWMP_BACKUP_DATA_CONFIG),y)
CONFIGURE_ARGS += \
	--enable-backupdatainconfig
endif

CONFIGURE_ARGS += \
	--enable-jsonc=1

define Package/ispapp/conffiles
/etc/config/ispapp
/usr/share/ispapp/defaults
endef

define Package/ispapp/install
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/bin/ispappd $(1)/usr/sbin
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) $(PKG_BUILD_DIR)/ext/openwrt/config/ispapp $(1)/etc/config
	$(INSTALL_DIR) $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/init.d/ispappd $(1)/etc/init.d
ifeq ($(ISPAPPCWMP_BACKUP_DATA_FILE),y)
	$(INSTALL_DIR) $(1)/etc/ispapp
endif
ifeq ($(CONFIG_ISPAPPCWMP_SCRIPTS_FULL),y)
	$(INSTALL_DIR) $(1)/usr/share/ispapp/functions/
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/defaults $(1)/usr/share/ispapp
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/common/* $(1)/usr/share/ispapp/functions/
ifeq ($(CONFIG_ISPAPPCWMP_DATA_MODEL_TR181),y)
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/tr181/* $(1)/usr/share/ispapp/functions/
else
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/tr098/* $(1)/usr/share/ispapp/functions/
endif
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/scripts/ispapp.sh $(1)/usr/sbin/ispapp
	chmod +x $(1)/usr/share/ispapp/functions/*
else
	$(INSTALL_DIR) $(1)/usr/share/ispapp/functions/
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/common/ipping_launch $(1)/usr/share/ispapp/functions/ipping_launch
endif
endef

$(eval $(call BuildPackage,ispapp))
