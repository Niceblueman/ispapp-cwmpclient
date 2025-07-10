# Copyright (C) 2012-2016 PIVA Software <www.pivasoftware.com>
# 	Author: MOHAMED Kallel <mohamed.kallel@pivasoftware.com>
# 	Author: ANIS ELLOUZE <anis.ellouze@pivasoftware.com>

include $(TOPDIR)/rules.mk

PKG_NAME:=ispappd
PKG_VERSION:=1.0.1
PKG_SOURCE:=$(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_FIXUP:=autoreconf
PKG_HASH:=b04b1ef883d05bb9f8ef29bdcc6918a70134728c9031594f6075aef358dc585b

PKG_CONFIG_DEPENDS:= \
	CONFIG_ISPAPPCWMP_DEBUG \
	CONFIG_ISPAPPCWMP_DEVEL

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)/$(PKG_NAME)-$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/ispappd
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=CWMP client (using libcurl)
  DEPENDS:=+libubus +libuci +libubox +libroxml +libjson-c +libcurl +curl
endef

define Package/ispappd/description
 An open source implementation of the client CWMP (TR-069) protocol
endef

define Package/ispappd/config
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
ifeq ($(ISPAPPCWMP_BACKUP_DATA_FILE),y)
	$(INSTALL_DIR) $(1)/etc/ispappd
endif
ifeq ($(CONFIG_ISPAPPCWMP_SCRIPTS_FULL),y)
	$(INSTALL_DIR) $(1)/usr/share/ispappd/functions/
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/defaults $(1)/usr/share/ispappd
	$(CP) $(PKG_BUILD_DIR)/ext/openwrt/scripts/functions/common/* $(1)/usr/share/ispappd/functions/
ifeq ($(CONFIG_ISPAPPCWMP_DATA_MODEL_TR181),y)
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
