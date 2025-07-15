include $(TOPDIR)/rules.mk

PKG_NAME:=goispappd
PKG_VERSION:=1.0.24
PKG_RELEASE:=24

GO_PKG:=github.com/Niceblueman/goispappd

PKG_SOURCE:=$(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_HASH:=e53e97ac70d5433d0380a810c19f3b454c2a0b7259946bdfdfcc8f87fc4b0ff8
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-v$(PKG_VERSION)

PKG_LICENSE:=GPL-3.0
PKG_LICENSE_FILES:=LICENSE.md
PKG_MAINTAINER:=Karim O. <support@ispapp.co>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/../feeds/packages/lang/golang/golang-package.mk

CGO_CFLAGS += \
	-I$(STAGING_DIR)/usr/include/libubox \
	-I$(STAGING_DIR)/usr/include/libubus \
	-I$(STAGING_DIR)/usr/include/libuci

define Package/goispappd/Default
  TITLE:=tr181/tr098/TR-069 CWMP client for OpenWrt.
  URL:=https://ispapp.co
  DEPENDS:=$(GO_ARCH_DEPENDS) +libubox +libubus +libuci
endef

define Package/goispappd
$(call Package/goispappd/Default)
  SECTION:=utils
  CATEGORY:=Network
endef

define Package/goispappd/description
    tr181/tr098/TR-069 CWMP client for OpenWrt.
    It is a Go port of the original ispappd written in C.
endef

$(eval $(call GoBinPackage,goispappd))
$(eval $(call BuildPackage,goispappd))