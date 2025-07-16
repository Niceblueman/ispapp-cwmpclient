include $(TOPDIR)/rules.mk

PKG_NAME:=goispappd
PKG_VERSION:=1.0.26
PKG_RELEASE:=26

GO_PKG:=github.com/Niceblueman/goispappd

PKG_SOURCE:=$(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_HASH:=2fb0a11e327a76f7244ec5cd1b0cbf37c1ca5cf6c9a72d4ec75de7acba152d35
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-v$(PKG_VERSION)

PKG_LICENSE:=GPL-3.0
PKG_LICENSE_FILES:=LICENSE.md
PKG_MAINTAINER:=Karim O. <support@ispapp.co>

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0

include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/../feeds/packages/lang/golang/golang-package.mk

GO_PKG_BUILD_VARS += GO111MODULE=on GOPROXY=direct GOSUMDB=off CGO_ENABLED=1
CGO_CFLAGS += \
	-I$(STAGING_DIR)/usr/include/libubox \
	-I$(STAGING_DIR)/usr/include/libubus \
	-I$(STAGING_DIR)/usr/include/libuci \
	-I$(STAGING_DIR)/usr/include/libjson-c \
	-I$(STAGING_DIR)/usr/include

CGO_LDFLAGS += \
	-L$(STAGING_DIR)/usr/lib \
	-luci \
	-lubus \
	-lubox \
	-lblobmsg_json \
	-ljson-c

GO_PKG_LDFLAGS:=-s -w

define Package/goispappd/Default
  TITLE:=tr181/tr098/TR-069 CWMP client for OpenWrt.
  URL:=https://ispapp.co
  DEPENDS:=$(GO_ARCH_DEPENDS) \
    +libubox \
    +libubus \
    +libuci \
    +libblobmsg_json \
    +libjson-c
endef

define Package/goispappd
$(call Package/goispappd/Default)
  SECTION:=utils
  CATEGORY:=Network
endef

define Build/Compile
    @echo "Go version being used:"
    $(GO_PKG_VARS) go version
    $(call GoBinPackage/Build/Compile)
endef

define Package/goispappd/description
tr181/tr098/TR-069 CWMP client for OpenWrt.
It is a Go port of the original ispappd written in C.
endef

$(eval $(call GoBinPackage,goispappd))
$(eval $(call BuildPackage,goispappd))