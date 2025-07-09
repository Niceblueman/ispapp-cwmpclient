#!/bin/bash

# Build IPK script using OpenWrt SDK container
docker run -it -d \
    --name openwrt-sdk-build \
    -v /Volumes/OpenWrt/ispapp-cwmpclient:/app \
    openwrt/sdk:bcm27xx-bcm2710-23.05.4 \
    /bin/bash

# Copy ispappd folder to the packages directory in the SDK
docker exec openwrt-sdk-build cp -r /app/ispappd /builder/package/
# Open bash shell in the container
docker exec -it openwrt-sdk-build /bin/bash
# Build the package
# docker exec openwrt-sdk-build make package/ispappd/compile V=s

echo "IPK build completed. Check build_dir for the generated package."
