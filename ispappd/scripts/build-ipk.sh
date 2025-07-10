#!/bin/bash

# Build IPK script using OpenWrt SDK container with privileges
# Remove existing container if it exists
docker rm -f openwrt-sdk-build 2>/dev/null || true

docker run -it -d \
    --name openwrt-sdk-build \
    --privileged \
    -v /Volumes/OpenWrt/ispapp-cwmpclient/feeds/ispappd:/builder/package/ispappd \
    openwrt/sdk:bcm27xx-bcm2710-23.05.4 \
    /bin/bash

# Open bash shell in the container
docker exec -it openwrt-sdk-build /bin/bash
# Build the package
# docker exec openwrt-sdk-build make package/ispappd/compile V=s

echo "IPK build completed. Check build_dir for the generated package."
