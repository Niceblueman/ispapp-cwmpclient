include $(TOPDIR)/rules.mk

PKG_NAME:=goispappd
PKG_VERSION:=1.0.23
PKG_RELEASE:=1

PKG_MAINTAINER:=Karim O. <support@ispapp.co>
PKG_LICENSE:=CC0-1.0
PKG_LICENSE_FILES:=LICENSE
# PKG_SOURCE:=$(PKG_NAME)-v$(PKG_VERSION).tar.gz
# PKG_SOURCE_URL:=https://github.com/Niceblueman/ispapp-cwmpclient/releases/download/v$(PKG_VERSION)/
# PKG_HASH:=53ff8ca80d0e7857ac034d4b54e70385cfecb44a5aafafed7da8ee5ee26c556c
PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_USE_MIPS16:=0
PKG_BUILD_FLAGS:=no-mips16

GO_PKG:=github.com/Niceblueman/goispappd
GO_PKG_BUILD_PKG:=./cmd/goispappd
GO_PKG_LDFLAGS_X:=
GO_PKG_TARGET_BIN_NAME:=goispappd

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/$(PKG_NAME)
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=Go Version of ispappd
  URL:=https://ispapp.co
  MAINTAINER:=Karim O. <support@ispapp.co>
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/$(PKG_NAME)/description
  An open source implementation of the client CWMP (TR-069) protocol written in Go.
  It is a lightweight and efficient alternative to the original C version, designed for easy integration and deployment.
endef

define Package/$(PKG_NAME)/install
	$(call GoPackage/Package/Install/Bin,$(1))
    $(INSTALL_DIR) $(1)/usr/sbin
    $(INSTALL_BIN) $(GO_PKG_TARGET_BIN_DIR)/goispappd $(1)/usr/sbin/
endef

$(eval $(call GoBinPackage,$(PKG_NAME)))
$(eval $(call BuildPackage,$(PKG_NAME)))