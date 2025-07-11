#!/bin/bash

# Build script for ispappd with proper object file destinations
set -e

# Configuration
BUILD_DIR="build-output"
OBJ_DIR="${BUILD_DIR}/obj"
BIN_DIR="${BUILD_DIR}/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create build directories
print_status "Creating build directories..."
mkdir -p "${OBJ_DIR}"
mkdir -p "${BIN_DIR}"

# Check if autotools is available
if ! command -v autoreconf &> /dev/null; then
    print_error "autoreconf not found. Please install autotools (autoconf, automake, libtool)"
    exit 1
fi

# Generate configure script if it doesn't exist
if [ ! -f "configure" ]; then
    print_status "Generating configure script..."
    autoreconf -fiv
fi

# Configure with build directory
print_status "Configuring build..."
if [ ! -f "Makefile" ]; then
    ./configure \
        --prefix=/usr \
        --with-build-dir="${BUILD_DIR}" \
        --enable-jsonc \
        --enable-debug
fi

# Build the project
print_status "Building project..."
make -j$(nproc) V=1

# Copy binaries to destination
if [ -f "bin/ispappd" ]; then
    print_status "Copying binary to ${BIN_DIR}..."
    cp bin/ispappd "${BIN_DIR}/"
fi

print_status "Build completed successfully!"
print_status "Binary location: ${BIN_DIR}/ispappd"
print_status "Object files location: ${OBJ_DIR}/"
