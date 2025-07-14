include $(TOPDIR)/rules.mk

PKG_NAME:=ispappd
PKG_VERSION:=1.0.20
PKG_SOURCE:=$(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_HASH:=0b8edfc267d36e99e5b92ba793bb3d618fb0e28920cae57dcef5a1333e4f134c

PKG_FIXUP:=autoreconf

PKG_CONFIG_DEPENDS:= \
	CONFIG_ISPAPPCWMP_DEBUG \
	CONFIG_ISPAPPCWMP_DEVEL

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)/$(PKG_NAME)-v$(PKG_VERSION)

include $(INCLUDE_DIR)/package.mk

define Package/ispappd
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=CWMP client (using libcurl)
  DEPENDS:=+libubus +libuci +libubox +libxml2 +libjson-c +libcurl +curl
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
	-D_GNU_SOURCE \
	-I$(STAGING_DIR)/usr/include/libxml2

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
	--enable-jsonc

define Package/ispappd/conffiles
/etc/config/ispappd
/usr/share/ispappd/defaults
endef

define Package/ispappd/install
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/bin/ispappcwmp $(1)/usr/sbin
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

define Build/Configure
	# Create necessary symlinks for libxml2 include paths
	mkdir -p $(PKG_BUILD_DIR)/include
	ln -sf $(STAGING_DIR)/usr/include/libxml2/libxml $(PKG_BUILD_DIR)/include/libxml
	
	(cd $(PKG_BUILD_DIR); \
	autoreconf -vif; \
	./configure \
		--target=$(GNU_TARGET_NAME) \
		--host=$(GNU_TARGET_NAME) \
		--build=$(GNU_HOST_NAME) \
		--prefix=/usr \
		--exec-prefix=/usr \
		--sysconfdir=/etc \
		--localstatedir=/var \
		$(CONFIGURE_ARGS) \
		CFLAGS="$(TARGET_CFLAGS) -I$(PKG_BUILD_DIR)/include" \
		LDFLAGS="$(TARGET_LDFLAGS)" \
		LIBXML2_CFLAGS="-I$(STAGING_DIR)/usr/include/libxml2" \
		LIBXML2_LIBS="-lxml2" )
endef

$(eval $(call BuildPackage,ispappd))
