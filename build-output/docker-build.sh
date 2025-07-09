#!/bin/bash
set -e

print_status() {
    echo -e "\033[0;34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Get build parameters from environment
DEBUG_BUILD=${DEBUG_BUILD:-false}
VERBOSE=${VERBOSE:-false}
PARALLEL_JOBS=${PARALLEL_JOBS:-4}
KEEP_RUNNING=${KEEP_RUNNING:-false}

print_status "Setting up ispappd package in SDK..."

# Copy package to SDK
rm -rf package/ispappd
mkdir -p package/ispappd

# Copy all source files with error checking
if [[ -d /source/src ]]; then
    cp -r /source/src package/ispappd/
else
    print_error "Source directory /source/src not found"
    exit 1
fi

if [[ -d /source/ext ]]; then
    cp -r /source/ext package/ispappd/
else
    print_error "Extension directory /source/ext not found"
    exit 1
fi

if [[ -d /source/bin ]]; then
    cp -r /source/bin package/ispappd/
else
    print_error "Binary directory /source/bin not found"
    exit 1
fi

if [[ -f /source/configure.ac ]]; then
    cp /source/configure.ac package/ispappd/
else
    print_error "File /source/configure.ac not found"
    exit 1
fi

if [[ -f /source/Makefile.am ]]; then
    cp /source/Makefile.am package/ispappd/
else
    print_error "File /source/Makefile.am not found"
    exit 1
fi

# Use the OpenWrt-specific Makefile and Config.in
if [[ -f /source/ext/openwrt/build/Makefile ]]; then
    cp /source/ext/openwrt/build/Makefile package/ispappd/Makefile
else
    print_error "OpenWrt Makefile not found at /source/ext/openwrt/build/Makefile"
    exit 1
fi

if [[ -f /source/ext/openwrt/build/Config.in ]]; then
    cp /source/ext/openwrt/build/Config.in package/ispappd/Config.in
else
    print_error "OpenWrt Config.in not found at /source/ext/openwrt/build/Config.in"
    exit 1
fi

# Configure feeds
print_status "Configuring feeds..."

# Update and install feeds
./scripts/feeds update -a
./scripts/feeds install -a

# Configure package
print_status "Configuring package build..."
make defconfig

# Enable ispappd package
echo "CONFIG_PACKAGE_ispappd=m" >> .config
echo "CONFIG_ISPAPPD_SCRIPTS_FULL=y" >> .config
echo "CONFIG_ISPAPPD_DATA_MODEL_TR181=y" >> .config

if [[ "$DEBUG_BUILD" == true ]]; then
    echo "CONFIG_ISPAPPD_DEBUG=y" >> .config
    echo "CONFIG_ISPAPPD_DEVEL=y" >> .config
fi

# Apply configuration
make defconfig

# Build package
print_status "Building package..."
BUILD_ARGS="-j$PARALLEL_JOBS"
if [[ "$VERBOSE" == true ]]; then
    BUILD_ARGS="$BUILD_ARGS V=s"
fi

if ! make package/ispappd/compile $BUILD_ARGS; then
    print_error "Build failed"
    exit 1
fi

# Find and copy built packages
print_status "Locating built packages..."
IPK_FILES=$(find bin/ -name "ispappd*.ipk" -type f)
if [[ -z "$IPK_FILES" ]]; then
    print_error "No IPK files found"
    exit 1
fi

# Copy IPK files to output directory
for ipk in $IPK_FILES; do
    cp "$ipk" /output/
    print_success "Created: $(basename "$ipk")"
done

print_success "Docker build completed successfully!"

# Keep container running if requested
if [[ "$KEEP_RUNNING" == true ]]; then
    print_status "Keeping container running for inspection..."
    print_status "Container will remain active for dependency analysis"
    print_status "You can connect to it using: docker exec -it <container_id> /bin/bash"
    print_status "Press Ctrl+C to exit and stop the container"
    
    # Keep the container running
    while true; do
        sleep 30
        print_status "Container still running... (Ctrl+C to stop)"
    done
fi
