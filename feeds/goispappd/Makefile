include $(TOPDIR)/rules.mk

PKG_NAME:=goispappd
PKG_VERSION:=0.0.1
PKG_RELEASE:=1

PKG_MAINTAINER:=Karim O. <support@ispapp.co>
PKG_LICENSE:=CC0-1.0
PKG_LICENSE_FILES:=LICENSE

PKG_BUILD_DEPENDS:=golang/host
PKG_BUILD_PARALLEL:=1
PKG_BUILD_FLAGS:=no-mips16

GO_PKG:=github.com/Niceblueman/goispappd

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/goispappd
	SECTION:=utils
	CATEGORY:=Utilities
	TITLE:=Go Version of ispappd
	URL:= ispapp.co
	MAINTAINER:=Karim O. <
	DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/goispappd/description
	An open source implementation of the client CWMP (TR-069) protocol written in Go.
	It is a lightweight and efficient alternative to the original C version, designed for easy integration and deployment.
endef

define Build/Prepare	
	$(CP) ./src/* $(PKG_BUILD_DIR)
endef

define Package/goispappd/install
	$(INSTALL_DIR) $(1)/usr/sbin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/goispappd $(1)/usr/sbin
endef

$(eval $(call BuildPackage,goispappd))