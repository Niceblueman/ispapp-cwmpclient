#!/bin/bash

# ISPAppD OpenWrt IPK Build Script
# This script builds ispappd IPK packages for various OpenWrt architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
OPENWRT_VERSION="23.05.4"
TARGET_ARCH=""
DEBUG_BUILD=false
CLEAN_BUILD=false
VERBOSE=false
PARALLEL_JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")

# Function to get SDK for architecture (compatible with older bash)
get_sdk_for_arch() {
    case "$1" in
        "mips_24kc") echo "ath79/generic" ;;
        "mipsel_24kc") echo "ramips/mt7621" ;;
        "mipsel_74kc") echo "ramips/mt7620" ;;
        "arm_cortex-a7_neon-vfpv4") echo "bcm27xx/bcm2710" ;;
        "arm_cortex-a53") echo "bcm27xx/bcm2711" ;;
        "arm_cortex-a15_neon-vfpv4") echo "ipq806x/generic" ;;
        "aarch64_cortex-a53") echo "bcm27xx/bcm2711" ;;
        "aarch64_cortex-a72") echo "bcm27xx/bcm2711" ;;
        "x86_64") echo "x86/64" ;;
        "i386") echo "x86/generic" ;;
        *) echo "" ;;
    esac
}

# Function to convert SDK path to filename format (slash to hyphen)
sdk_path_to_filename() {
    echo "$1" | sed 's#/#-#g'
}

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -a, --arch ARCH      Target architecture (required)"
    echo "  -v, --version VER    OpenWrt version (default: $OPENWRT_VERSION)"
    echo "  -d, --debug          Build with debug flags"
    echo "  -c, --clean          Clean build before starting"
    echo "  -j, --jobs N         Number of parallel jobs (default: $PARALLEL_JOBS)"
    echo "  --verbose            Enable verbose output"
    echo "  --list-archs         List available architectures"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --arch x86_64"
    echo "  $0 --arch mips_24kc --debug"
    echo "  $0 --arch arm_cortex-a53 --clean --verbose"
}

list_architectures() {
    echo "Available architectures:"
    echo "  mips_24kc                      -> ath79/generic"
    echo "  mipsel_24kc                    -> ramips/mt7621"
    echo "  mipsel_74kc                    -> ramips/mt7620"
    echo "  arm_cortex-a7_neon-vfpv4       -> bcm27xx/bcm2710"
    echo "  arm_cortex-a53                 -> bcm27xx/bcm2711"
    echo "  arm_cortex-a15_neon-vfpv4      -> ipq806x/generic"
    echo "  aarch64_cortex-a53             -> bcm27xx/bcm2711"
    echo "  aarch64_cortex-a72             -> bcm27xx/bcm2711"
    echo "  x86_64                         -> x86/64"
    echo "  i386                           -> x86/generic"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--arch)
            TARGET_ARCH="$2"
            shift 2
            ;;
        -v|--version)
            OPENWRT_VERSION="$2"
            shift 2
            ;;
        -d|--debug)
            DEBUG_BUILD=true
            shift
            ;;
        -c|--clean)
            CLEAN_BUILD=true
            shift
            ;;
        -j|--jobs)
            PARALLEL_JOBS="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --list-archs)
            list_architectures
            exit 0
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate required arguments
if [[ -z "$TARGET_ARCH" ]]; then
    print_error "Target architecture is required"
    echo ""
    show_help
    exit 1
fi

# Validate architecture
SDK_TARGET=$(get_sdk_for_arch "$TARGET_ARCH")
if [[ -z "$SDK_TARGET" ]]; then
    print_error "Unknown architecture: $TARGET_ARCH"
    echo ""
    list_architectures
    exit 1
fi

print_status "Building ispappd IPK package"
print_status "Target architecture: $TARGET_ARCH"
print_status "OpenWrt SDK: $SDK_TARGET"
print_status "OpenWrt version: $OPENWRT_VERSION"
print_status "Debug build: $DEBUG_BUILD"
print_status "Parallel jobs: $PARALLEL_JOBS"

# Check for required tools
REQUIRED_TOOLS=("wget" "tar" "make" "gcc" "g++" "file")
for tool in "${REQUIRED_TOOLS[@]}"; do
    if ! command -v "$tool" &> /dev/null; then
        print_error "Required tool '$tool' is not installed"
        exit 1
    fi
done

# Clean if requested
if [[ "$CLEAN_BUILD" == true ]]; then
    print_status "Cleaning previous builds..."
    rm -rf openwrt-sdk-* *.ipk
fi

# Download OpenWrt SDK
SDK_FILENAME_TARGET=$(sdk_path_to_filename "$SDK_TARGET")

# Determine SDK filename pattern based on target
# Some targets use musl_eabi instead of musl
if [[ "$SDK_TARGET" == "ipq806x/generic" ]]; then
    SDK_FILENAME="openwrt-sdk-${OPENWRT_VERSION}-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl_eabi.Linux-x86_64.tar.xz"
else
    SDK_FILENAME="openwrt-sdk-${OPENWRT_VERSION}-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
fi

SDK_URL="https://downloads.openwrt.org/releases/${OPENWRT_VERSION}/targets/${SDK_TARGET}/${SDK_FILENAME}"

if [[ ! -f "$SDK_FILENAME" ]]; then
    print_status "Downloading OpenWrt SDK..."
    print_status "URL: $SDK_URL"
    
    if ! wget -q --show-progress "$SDK_URL"; then
        # Try fallback with different musl pattern
        if [[ "$SDK_TARGET" != "ipq806x/generic" ]]; then
            print_warning "Primary download failed, trying musl_eabi pattern..."
            SDK_FILENAME="openwrt-sdk-${OPENWRT_VERSION}-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl_eabi.Linux-x86_64.tar.xz"
            SDK_URL="https://downloads.openwrt.org/releases/${OPENWRT_VERSION}/targets/${SDK_TARGET}/${SDK_FILENAME}"
            print_status "Fallback URL: $SDK_URL"
            
            if ! wget -q --show-progress "$SDK_URL"; then
                print_error "Failed to download SDK with both patterns"
                print_error "Last tried URL: $SDK_URL"
                exit 1
            fi
        else
            print_error "Failed to download SDK"
            print_error "URL: $SDK_URL"
            exit 1
        fi
    fi
fi

# Extract SDK
print_status "Extracting SDK..."
if [[ ! -d "openwrt-sdk-${OPENWRT_VERSION}-${SDK_FILENAME_TARGET}" ]]; then
    tar xf "$SDK_FILENAME"
fi

SDK_DIR=$(find . -maxdepth 1 -type d -name "openwrt-sdk-*" | head -1)
if [[ -z "$SDK_DIR" ]]; then
    print_error "SDK directory not found"
    exit 1
fi

print_status "Using SDK directory: $SDK_DIR"

# Copy package to SDK
print_status "Copying package files to SDK..."
rm -rf "$SDK_DIR/package/ispappd"
mkdir -p "$SDK_DIR/package/ispappd"

# Copy all source files
cp -r src/ "$SDK_DIR/package/ispappd/"
cp -r ext/ "$SDK_DIR/package/ispappd/"
cp -r bin/ "$SDK_DIR/package/ispappd/"
cp configure.ac "$SDK_DIR/package/ispappd/"
cp Makefile.am "$SDK_DIR/package/ispappd/"

# Use the OpenWrt-specific Makefile and Config.in
cp ext/openwrt/build/Makefile "$SDK_DIR/package/ispappd/Makefile"
cp ext/openwrt/build/Config.in "$SDK_DIR/package/ispappd/Config.in"

# Configure feeds
print_status "Configuring feeds..."
cd "$SDK_DIR"

# Add local package feed
if ! grep -q "ispappd_local" feeds.conf.default; then
    echo "src-link ispappd_local $(pwd)/package" >> feeds.conf.default
fi

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

# Find built packages
print_status "Locating built packages..."
cd - > /dev/null

IPK_FILES=$(find "$SDK_DIR/bin" -name "ispappd*.ipk" -type f)
if [[ -z "$IPK_FILES" ]]; then
    print_error "No IPK files found"
    exit 1
fi

# Copy IPK files to current directory
for ipk in $IPK_FILES; do
    cp "$ipk" .
    print_success "Created: $(basename "$ipk")"
done

# Show package information
print_status "Package information:"
for ipk in *.ipk; do
    if [[ -f "$ipk" ]]; then
        SIZE=$(du -h "$ipk" | cut -f1)
        echo "  - $ipk ($SIZE)"
    fi
done

print_success "IPK build completed successfully!"
print_status "Architecture: $TARGET_ARCH"
print_status "Build type: $([ "$DEBUG_BUILD" == true ] && echo "debug" || echo "release")"
print_status "Files: $(echo *.ipk | tr ' ' ', ')"

# Verify IPK contents
if command -v ar &> /dev/null; then
    print_status "IPK package contents:"
    for ipk in ispappd*.ipk; do
        if [[ -f "$ipk" ]]; then
            echo "  $ipk:"
            ar -t "$ipk" | sed 's/^/    /'
        fi
    done
fi
