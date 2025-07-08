#!/bin/bash
# Test script to verify SDK download fixes

set -e

echo "Testing SDK download fixes..."

# Test the problematic ipq806x/generic target
SDK_TARGET="ipq806x/generic"
SDK_FILENAME_TARGET=$(echo "$SDK_TARGET" | sed 's#/#-#g')

echo "Testing SDK filename pattern for $SDK_TARGET..."
echo "SDK_FILENAME_TARGET: $SDK_FILENAME_TARGET"

# Test the correct filename pattern
if [ "$SDK_TARGET" = "ipq806x/generic" ]; then
    SDK_FILENAME="openwrt-sdk-23.05.4-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl_eabi.Linux-x86_64.tar.xz"
else
    SDK_FILENAME="openwrt-sdk-23.05.4-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
fi

SDK_URL="https://downloads.openwrt.org/releases/23.05.4/targets/${SDK_TARGET}/${SDK_FILENAME}"

echo "Testing URL: $SDK_URL"

# Test if URL exists
if curl -I "$SDK_URL" 2>/dev/null | head -1 | grep -q "200"; then
    echo "✅ SDK URL is accessible: $SDK_URL"
else
    echo "❌ SDK URL is not accessible: $SDK_URL"
    exit 1
fi

# Test a few other targets
echo ""
echo "Testing other targets..."

for target in "x86/64" "ath79/generic" "bcm27xx/bcm2711"; do
    SDK_FILENAME_TARGET=$(echo "$target" | sed 's#/#-#g')
    SDK_FILENAME="openwrt-sdk-23.05.4-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
    SDK_URL="https://downloads.openwrt.org/releases/23.05.4/targets/${target}/${SDK_FILENAME}"
    
    if curl -I "$SDK_URL" 2>/dev/null | head -1 | grep -q "200"; then
        echo "✅ $target: $SDK_URL"
    else
        echo "❌ $target: $SDK_URL"
    fi
done

echo ""
echo "SDK download test completed!"
