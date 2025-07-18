include $(TOPDIR)/rules.mk

PKG_NAME := goispappd
PKG_VERSION := 1.0.30
PKG_RELEASE := 30

GO_PKG := github.com/Niceblueman/goispappd

PKG_SOURCE := $(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL := https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_HASH := 1bf937e4dd10c77dc278348692adc61fe8fabbc90b6144ac82836ffeb0d6b14d
PKG_BUILD_DIR := $(BUILD_DIR)/$(PKG_NAME)-v$(PKG_VERSION)

PKG_LICENSE := GPL-3.0
PKG_LICENSE_FILES := LICENSE.md
PKG_MAINTAINER := Karim O. <support@ispapp.co>

PKG_BUILD_DEPENDS := golang/host
PKG_BUILD_PARALLEL := 1
PKG_USE_MIPS16 := 0

include $(INCLUDE_DIR)/package.mk
include $(INCLUDE_DIR)/../feeds/packages/lang/golang/golang-package.mk

GO_PKG_BUILD_VARS += GO111MODULE=on GOPROXY=direct GOSUMDB=off CGO_ENABLED=1
CGO_CFLAGS += -luci
GO_PKG_LDFLAGS := -s -w

define Package/goispappd/Default
  TITLE:=tr181/tr098/TR-069 CWMP client for OpenWrt.
  URL:=https://ispapp.co
  DEPENDS:=$(GO_ARCH_DEPENDS) +libuci

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
