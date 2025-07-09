#!/bin/bash

# ISPAppD OpenWrt IPK Build Script
# This script builds ispappd IPK packages for various OpenWrt architectures using Docker

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
USE_DOCKER=true
DOCKER_REGISTRY="openwrt/sdk"
PARALLEL_JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")

# Function to convert SDK path to filename format (slash to hyphen)
sdk_path_to_filename() {
    echo "$1" | sed 's#/#-#g'
}

# Function to get Docker SDK tag for architecture
get_docker_sdk_tag() {
    local arch="$1"
    local version="$OPENWRT_VERSION"
    
    case "$arch" in
        "mips_24kc") echo "ath79-generic-${version}" ;;
        "mipsel_24kc") echo "ramips-mt7621-${version}" ;;
        "mipsel_74kc") echo "ramips-mt7620-${version}" ;;
        "arm_cortex-a7_neon-vfpv4") echo "bcm27xx-bcm2710-${version}" ;;
        "arm_cortex-a53") echo "bcm27xx-bcm2711-${version}" ;;
        "arm_cortex-a15_neon-vfpv4") echo "ipq806x-generic-${version}" ;;
        "aarch64_cortex-a53") echo "bcm27xx-bcm2711-${version}" ;;
        "aarch64_cortex-a72") echo "bcm27xx-bcm2711-${version}" ;;
        "x86_64") echo "x86-64-${version}" ;;
        "i386") echo "x86-generic-${version}" ;;
        *) echo "" ;;
    esac
}

# Function to get SDK for architecture (compatible with older bash) - kept for reference
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
    echo "  --no-docker          Use traditional SDK download instead of Docker (not recommended on macOS ARM)"
    echo "  --list-archs         List available architectures"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --arch x86_64"
    echo "  $0 --arch mips_24kc --debug"
    echo "  $0 --arch arm_cortex-a53 --clean --verbose"
    echo ""
    echo "Docker Usage (default):"
    echo "  Uses OpenWrt SDK Docker containers to avoid x86_64 binary compatibility issues on ARM macOS"
    echo ""
    echo "Production Builds:"
    echo "  For production builds, use the GitHub Actions workflow (.github/workflows/build-ipk.yml)"
    echo "  which uses the official openwrt/gh-action-sdk@main action for reliable builds"
}

list_architectures() {
    echo "Available architectures:"
    echo "  mips_24kc                      -> ath79-generic (Docker: ath79-generic-${OPENWRT_VERSION})"
    echo "  mipsel_24kc                    -> ramips-mt7621 (Docker: ramips-mt7621-${OPENWRT_VERSION})"
    echo "  mipsel_74kc                    -> ramips-mt7620 (Docker: ramips-mt7620-${OPENWRT_VERSION})"
    echo "  arm_cortex-a7_neon-vfpv4       -> bcm27xx-bcm2710 (Docker: bcm27xx-bcm2710-${OPENWRT_VERSION})"
    echo "  arm_cortex-a53                 -> bcm27xx-bcm2711 (Docker: bcm27xx-bcm2711-${OPENWRT_VERSION})"
    echo "  arm_cortex-a15_neon-vfpv4      -> ipq806x-generic (Docker: ipq806x-generic-${OPENWRT_VERSION})"
    echo "  aarch64_cortex-a53             -> bcm27xx-bcm2711 (Docker: bcm27xx-bcm2711-${OPENWRT_VERSION})"
    echo "  aarch64_cortex-a72             -> bcm27xx-bcm2711 (Docker: bcm27xx-bcm2711-${OPENWRT_VERSION})"
    echo "  x86_64                         -> x86-64 (Docker: x86-64-${OPENWRT_VERSION})"
    echo "  i386                           -> x86-generic (Docker: x86-generic-${OPENWRT_VERSION})"
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
        --no-docker)
            USE_DOCKER=false
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
if [[ "$USE_DOCKER" == true ]]; then
    DOCKER_TAG=$(get_docker_sdk_tag "$TARGET_ARCH")
    if [[ -z "$DOCKER_TAG" ]]; then
        print_error "Unknown architecture for Docker: $TARGET_ARCH"
        echo ""
        list_architectures
        exit 1
    fi
else
    SDK_TARGET=$(get_sdk_for_arch "$TARGET_ARCH")
    if [[ -z "$SDK_TARGET" ]]; then
        print_error "Unknown architecture: $TARGET_ARCH"
        echo ""
        list_architectures
        exit 1
    fi
fi

print_status "Building ispappd IPK package"
print_status "Target architecture: $TARGET_ARCH"
if [[ "$USE_DOCKER" == true ]]; then
    print_status "Using Docker SDK: ${DOCKER_REGISTRY}:${DOCKER_TAG}"
else
    print_status "OpenWrt SDK: $SDK_TARGET"
fi
print_status "OpenWrt version: $OPENWRT_VERSION"
print_status "Debug build: $DEBUG_BUILD"
print_status "Parallel jobs: $PARALLEL_JOBS"
print_status "Docker mode: $USE_DOCKER"

# Check for required tools
if [[ "$USE_DOCKER" == true ]]; then
    REQUIRED_TOOLS=("docker")
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
else
    REQUIRED_TOOLS=("wget" "tar" "make" "gcc" "g++" "file")
fi

for tool in "${REQUIRED_TOOLS[@]}"; do
    if ! command -v "$tool" &> /dev/null; then
        print_error "Required tool '$tool' is not installed"
        exit 1
    fi
done

# Clean if requested
if [[ "$CLEAN_BUILD" == true ]]; then
    print_status "Cleaning previous builds..."
    if [[ "$USE_DOCKER" == true ]]; then
        rm -rf *.ipk build-output/
    else
        rm -rf openwrt-sdk-* *.ipk
    fi
fi

if [[ "$USE_DOCKER" == true ]]; then
    # Docker-based build
    print_status "Building with Docker SDK container..."
    
    # Create output directory
    mkdir -p build-output
    
    # Prepare build script for container
    cat > build-output/docker-build.sh << 'EOF'
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

print_status "Setting up ispappd package in SDK..."

# Copy package to SDK
rm -rf package/ispappd
mkdir -p package/ispappd

# Copy all source files
cp -r /source/src/ package/ispappd/
cp -r /source/ext/ package/ispappd/
cp -r /source/bin/ package/ispappd/
cp /source/configure.ac package/ispappd/
cp /source/Makefile.am package/ispappd/

# Use the OpenWrt-specific Makefile and Config.in
cp /source/ext/openwrt/build/Makefile package/ispappd/Makefile
cp /source/ext/openwrt/build/Config.in package/ispappd/Config.in

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
EOF

    chmod +x build-output/docker-build.sh
    
    # Run Docker build
    print_status "Starting Docker container: ${DOCKER_REGISTRY}:${DOCKER_TAG}"
    
    DOCKER_ARGS=(
        "--rm"
        "-v" "$(pwd):/source:ro"
        "-v" "$(pwd)/build-output:/output"
        "-e" "DEBUG_BUILD=$DEBUG_BUILD"
        "-e" "VERBOSE=$VERBOSE"
        "-e" "PARALLEL_JOBS=$PARALLEL_JOBS"
        "${DOCKER_REGISTRY}:${DOCKER_TAG}"
        "/output/docker-build.sh"
    )
    
    if ! docker run "${DOCKER_ARGS[@]}"; then
        print_error "Docker build failed"
        exit 1
    fi
    
    # Move IPK files to current directory
    if ls build-output/*.ipk 1> /dev/null 2>&1; then
        mv build-output/*.ipk .
        print_success "IPK files moved to current directory"
    else
        print_error "No IPK files found in build output"
        exit 1
    fi
    
else
    # Traditional SDK download build (kept for compatibility)
    print_warning "Using traditional SDK download. This may not work on ARM macOS."
    
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
fi

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
print_status "Build method: $([ "$USE_DOCKER" == true ] && echo "Docker" || echo "Traditional SDK")"
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

# Clean up Docker build files
if [[ "$USE_DOCKER" == true ]] && [[ -d "build-output" ]]; then
    rm -rf build-output
fi
