include $(TOPDIR)/rules.mk

PKG_NAME := goispappd
PKG_VERSION := 1.0.31
PKG_RELEASE := 31

GO_PKG := github.com/Niceblueman/goispappd

PKG_SOURCE := $(PKG_NAME)-v$(PKG_VERSION).tar.gz
PKG_SOURCE_URL := https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
PKG_HASH := b1e97d16549c737ce6f59443ac2ac64ef725d3a7943dccd28255d1999d8c3860
PKG_BUILD_DIR := $(BUILD_DIR)/$(PKG_NAME)-v$(PKG_VERSION)

PKG_LICENSE := GPL-3.0
PKG_LICENSE_FILES := LICENSE.md
PKG_MAINTAINER := Karim O. <support@ispapp.co>

PKG_BUILD_DEPENDS := golang/host
PKG_BUILD_PARALLEL := 1
PKG_USE_MIPS16 := 0

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

GO_PKG_BUILD_VARS += CGO_ENABLED=1 GO111MODULE=on GOPROXY=direct
GO_PKG_GCFLAGS += -luci
GO_PKG_LDFLAGS := -s -w
GO_PKG_BUILD_PKG := github.com/Niceblueman/goispappd

define Package/goispappd/Default
  TITLE:=tr181/tr098/TR-069 CWMP client for OpenWrt.
  URL:=https://ispapp.co
  DEPENDS:=$(GO_ARCH_DEPENDS) +libuci
endef

define Package/goispappd
	$(call Package/goispappd/Default)
	SECTION:=utils
	CATEGORY:=Network
	SUBMENU:=Go Applications
	USERID:=goispappd:goispappd
endef

define Package/goispappd/install
	$(call GoPackage/Package/Install/Bin,$(PKG_INSTALL_DIR))
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/goispappd $(1)/usr/bin/goispappd
endef

define Package/goispappd/description
tr181/tr098/TR-069 CWMP client for OpenWrt.
It is a Go port of the original ispappd written in C.
endef

$(eval $(call GoBinPackage,goispappd))
$(eval $(call BuildPackage,goispappd))
