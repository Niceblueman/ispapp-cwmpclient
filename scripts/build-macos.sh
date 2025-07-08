#!/bin/bash

# ISPAppD macOS ARM Build Script
# This script builds ispappd natively on macOS with stub UCI/ubus libraries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0    # Try to install mxml via Homebrew first as it's more reliable
    if ! pkg-config --exists mxml; then
        print_status "Installing libmxml via Homebrew..."
        brew install libmxml
        export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:$PKG_CONFIG_PATH"
    fiBLUE='\033[0;34m'
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

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This script is designed for macOS only"
    exit 1
fi

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    print_error "Homebrew is required but not installed"
    print_status "Install Homebrew from https://brew.sh/"
    exit 1
fi

# Default values
BUILD_TYPE="release"
INSTALL_PREFIX="$HOME/ispappd-macos"
CLEAN_BUILD=false
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            BUILD_TYPE="debug"
            shift
            ;;
        --prefix)
            INSTALL_PREFIX="$2"
            shift 2
            ;;
        --clean)
            CLEAN_BUILD=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --debug      Build with debug flags"
            echo "  --prefix     Installation prefix (default: $HOME/ispappd-macos)"
            echo "  --clean      Clean build directory before building"
            echo "  --verbose    Enable verbose output"
            echo "  --help       Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_status "Building ispappd for macOS ARM64"
print_status "Build type: $BUILD_TYPE"
print_status "Install prefix: $INSTALL_PREFIX"

# Clean if requested
if [[ "$CLEAN_BUILD" == true ]]; then
    print_status "Cleaning build directory..."
    make clean 2>/dev/null || true
    rm -rf autom4te.cache config.log config.status Makefile bin/Makefile
    rm -rf macos-stubs
fi

# Install dependencies
print_status "Installing dependencies via Homebrew..."
brew update
brew install autoconf automake libtool curl json-c pkg-config libmicrohttpd

# Create stub libraries directory
print_status "Creating stub UCI/ubus libraries..."
mkdir -p macos-stubs/include macos-stubs/lib macos-stubs/src

# Create stub headers
cat > macos-stubs/include/uci.h << 'EOF'
#ifndef __UCI_H
#define __UCI_H

// Stub UCI header for macOS builds
// This provides minimal interface compatibility

struct uci_context;
struct uci_package;
struct uci_section;
struct uci_option;
struct uci_element;

// Minimal function stubs
static inline struct uci_context* uci_alloc_context(void) { return NULL; }
static inline void uci_free_context(struct uci_context *ctx) { }
static inline int uci_load(struct uci_context *ctx, const char *name, struct uci_package **package) { return -1; }
static inline void uci_unload(struct uci_context *ctx, struct uci_package *p) { }

#endif
EOF

cat > macos-stubs/include/libubox/uloop.h << 'EOF'
#ifndef __ULOOP_H
#define __ULOOP_H

// Stub uloop header for macOS builds
static inline int uloop_init(void) { return 0; }
static inline void uloop_run(void) { }
static inline void uloop_done(void) { }

#endif
EOF

cat > macos-stubs/include/libubox/usock.h << 'EOF'
#ifndef __USOCK_H
#define __USOCK_H

// Stub usock header for macOS builds
static inline int usock(int type, const char *host, const char *service) { return -1; }

#endif
EOF

cat > macos-stubs/include/libubus.h << 'EOF'
#ifndef __LIBUBUS_H
#define __LIBUBUS_H

// Stub libubus header for macOS builds
struct ubus_context;
struct ubus_request_data;

static inline struct ubus_context* ubus_connect(const char *path) { return NULL; }
static inline void ubus_free(struct ubus_context *ctx) { }

#endif
EOF

# Create stub source files
cat > macos-stubs/src/uci_stub.c << 'EOF'
// Stub UCI implementation for macOS builds
void uci_stub_function(void) { }
EOF

cat > macos-stubs/src/ubox_stub.c << 'EOF'
// Stub libubox implementation for macOS builds
void ubox_stub_function(void) { }
EOF

cat > macos-stubs/src/ubus_stub.c << 'EOF'
// Stub libubus implementation for macOS builds
void ubus_stub_function(void) { }
EOF

# Compile stub object files and create libraries
print_status "Creating stub libraries..."
cd macos-stubs/src
gcc -c uci_stub.c -o uci_stub.o
gcc -c ubox_stub.c -o ubox_stub.o
gcc -c ubus_stub.c -o ubus_stub.o

# Create stub libraries with object files
ar rcs ../lib/libuci.a uci_stub.o
ar rcs ../lib/libubox.a ubox_stub.o
ar rcs ../lib/libubus.a ubus_stub.o

cd ../..

# Install XML library - prefer libmxml from Homebrew
print_status "Setting up XML library..."

# First, ensure libmxml is installed
if ! brew list libmxml &>/dev/null; then
    print_status "Installing libmxml via Homebrew..."
    brew install libmxml
fi

# Set up pkg-config path
export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:$PKG_CONFIG_PATH"

# Check if mxml4 is available via pkg-config (libmxml provides mxml4.pc)
if pkg-config --exists mxml4; then
    print_success "Using mxml4 from Homebrew (libmxml package)"
    
    # Create a microxml.pc that points to mxml4 for compatibility
    MXML_VERSION=$(pkg-config --modversion mxml4)
    MXML_CFLAGS=$(pkg-config --cflags mxml4)
    MXML_LIBS=$(pkg-config --libs mxml4)
    
    # Create compatibility pkg-config file
    mkdir -p /opt/homebrew/lib/pkgconfig
    cat > /opt/homebrew/lib/pkgconfig/microxml.pc << EOF
prefix=/opt/homebrew
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: microxml
Description: Mini-XML compatibility layer using mxml4
Version: $MXML_VERSION
Libs: $MXML_LIBS
Cflags: $MXML_CFLAGS
EOF
    print_success "Created microxml compatibility layer for mxml4"
    
elif pkg-config --exists mxml; then
    print_success "Using mxml from Homebrew"
    
    # Create compatibility layer for mxml
    MXML_VERSION=$(pkg-config --modversion mxml)
    MXML_CFLAGS=$(pkg-config --cflags mxml)
    MXML_LIBS=$(pkg-config --libs mxml)
    
    mkdir -p /opt/homebrew/lib/pkgconfig
    cat > /opt/homebrew/lib/pkgconfig/microxml.pc << EOF
prefix=/opt/homebrew
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: microxml
Description: Mini-XML compatibility layer using mxml
Version: $MXML_VERSION
Libs: $MXML_LIBS
Cflags: $MXML_CFLAGS
EOF
    print_success "Created microxml compatibility layer for mxml"
    
else
    print_warning "No XML library found, creating stub configuration"
    # Create stub microxml.pc to satisfy configure
    mkdir -p /opt/homebrew/lib/pkgconfig
    cat > /opt/homebrew/lib/pkgconfig/microxml.pc << EOF
prefix=/opt/homebrew
exec_prefix=\${prefix}
libdir=\${exec_prefix}/lib
includedir=\${prefix}/include

Name: microxml
Description: Stub microxml for compilation
Version: 1.0.0
Libs: 
Cflags: -DNO_XML
EOF
    print_warning "XML functionality will be limited"
fi

# Generate configure script
print_status "Generating configure script..."
autoreconf -fiv

# Configure build
print_status "Configuring build..."
CONFIGURE_ARGS="--prefix=$INSTALL_PREFIX --enable-jsonc"

if [[ "$BUILD_TYPE" == "debug" ]]; then
    CONFIGURE_ARGS="$CONFIGURE_ARGS --enable-debug --enable-devel"
fi

export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:$PKG_CONFIG_PATH"
export CFLAGS="-I/opt/homebrew/include -DMACOS_BUILD"
export LDFLAGS="-L/opt/homebrew/lib"

if [[ "$VERBOSE" == true ]]; then
    print_status "Configure command:"
    echo "./configure $CONFIGURE_ARGS \\"
    echo "  --with-uci-include-path=$(pwd)/macos-stubs/include \\"
    echo "  --with-uci-lib-path=$(pwd)/macos-stubs/lib \\"
    echo "  --with-libubox-include-path=$(pwd)/macos-stubs/include \\"
    echo "  --with-libubox-lib-path=$(pwd)/macos-stubs/lib \\"
    echo "  --with-libubus-include-path=$(pwd)/macos-stubs/include \\"
    echo "  --with-libubus-lib-path=$(pwd)/macos-stubs/lib"
fi

./configure $CONFIGURE_ARGS \
  --with-uci-include-path=$(pwd)/macos-stubs/include \
  --with-uci-lib-path=$(pwd)/macos-stubs/lib \
  --with-libubox-include-path=$(pwd)/macos-stubs/include \
  --with-libubox-lib-path=$(pwd)/macos-stubs/lib \
  --with-libubus-include-path=$(pwd)/macos-stubs/include \
  --with-libubus-lib-path=$(pwd)/macos-stubs/lib

# Build
print_status "Building..."
if [[ "$VERBOSE" == true ]]; then
    make -j$(sysctl -n hw.ncpu) V=1
else
    make -j$(sysctl -n hw.ncpu)
fi

print_success "Build completed successfully!"

# Install
print_status "Installing to $INSTALL_PREFIX..."
make install

print_success "Installation completed!"
print_status "Binary location: $INSTALL_PREFIX/bin/ispappd"

# Create a distribution package
DIST_DIR="$PWD/ispappd-macos-arm64-$BUILD_TYPE"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

print_status "Creating distribution package..."
make install DESTDIR="$DIST_DIR"

cd "$DIST_DIR"
tar -czf "../ispappd-macos-arm64-$BUILD_TYPE.tar.gz" .
cd - > /dev/null

print_success "Distribution package created: ispappd-macos-arm64-$BUILD_TYPE.tar.gz"

print_status "Build summary:"
echo "  - Architecture: macOS ARM64"
echo "  - Build type: $BUILD_TYPE"
echo "  - Binary: $INSTALL_PREFIX/bin/ispappd"
echo "  - Package: ispappd-macos-arm64-$BUILD_TYPE.tar.gz"
echo ""
print_success "ispappd build for macOS ARM64 completed successfully!"
