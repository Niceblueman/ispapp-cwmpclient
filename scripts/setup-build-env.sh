#!/bin/bash

# OpenWrt Build Environment Setup Script
# Handles different Linux distributions and versions

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Detect OS and version
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS_ID="$ID"
        OS_VERSION="$VERSION_ID"
        OS_NAME="$NAME"
    elif [[ -f /etc/redhat-release ]]; then
        OS_ID="rhel"
        OS_NAME=$(cat /etc/redhat-release)
    elif [[ -f /etc/debian_version ]]; then
        OS_ID="debian"
        OS_NAME="Debian"
    else
        OS_ID="unknown"
        OS_NAME="Unknown"
    fi
    
    print_status "Detected OS: $OS_NAME ($OS_ID $OS_VERSION)"
}

# Install dependencies for Ubuntu/Debian
install_ubuntu_deps() {
    print_status "Installing Ubuntu/Debian dependencies..."
    
    # Update package list
    sudo apt-get update
    
    # Core build dependencies
    PACKAGES="build-essential wget curl tar xz-utils gcc g++ libc6-dev make git unzip libncurses5-dev libssl-dev zlib1g-dev file time"
    
    # Handle python3-distutils vs python3-setuptools based on OS version
    # python3-distutils was removed in Ubuntu 22.04+ and Debian 12+
    if [[ "$OS_ID" == "ubuntu" ]]; then
        # Extract major version number
        UBUNTU_VERSION=$(echo "$OS_VERSION" | cut -d. -f1)
        if [[ "$UBUNTU_VERSION" -ge 22 ]]; then
            print_status "Ubuntu $OS_VERSION detected - using python3-setuptools"
            PACKAGES="$PACKAGES python3-setuptools python3-dev"
        else
            print_status "Ubuntu $OS_VERSION detected - using python3-distutils"
            PACKAGES="$PACKAGES python3-distutils"
        fi
    elif [[ "$OS_ID" == "debian" ]]; then
        # Extract major version number
        DEBIAN_VERSION=$(echo "$OS_VERSION" | cut -d. -f1)
        if [[ "$DEBIAN_VERSION" -ge 12 ]]; then
            print_status "Debian $OS_VERSION detected - using python3-setuptools"
            PACKAGES="$PACKAGES python3-setuptools python3-dev"
        else
            print_status "Debian $OS_VERSION detected - using python3-distutils"
            PACKAGES="$PACKAGES python3-distutils"
        fi
    else
        # For other Debian-based distributions, try setuptools first
        print_status "Debian-based system - trying python3-setuptools"
        PACKAGES="$PACKAGES python3-setuptools python3-dev"
    fi
    
    # Additional OpenWrt build dependencies
    PACKAGES="$PACKAGES clang flex bison gawk gcc-multilib g++-multilib gettext rsync"
    
    print_status "Installing packages: $PACKAGES"
    sudo apt-get install -y $PACKAGES
    
    print_success "Ubuntu/Debian dependencies installed"
}

# Install dependencies for CentOS/RHEL/Fedora
install_rhel_deps() {
    print_status "Installing Red Hat family dependencies..."
    
    if command -v dnf >/dev/null 2>&1; then
        # Fedora / newer RHEL
        sudo dnf groupinstall -y "Development Tools"
        sudo dnf install -y wget curl tar xz gcc gcc-c++ make git unzip \
            ncurses-devel openssl-devel zlib-devel python3-setuptools \
            clang flex bison gawk gettext rsync file
    elif command -v yum >/dev/null 2>&1; then
        # CentOS / older RHEL
        sudo yum groupinstall -y "Development Tools"
        sudo yum install -y wget curl tar xz gcc gcc-c++ make git unzip \
            ncurses-devel openssl-devel zlib-devel python3-setuptools \
            clang flex bison gawk gettext rsync file
    else
        print_error "No package manager found (dnf/yum)"
        exit 1
    fi
    
    print_success "Red Hat family dependencies installed"
}

# Install dependencies for Arch Linux
install_arch_deps() {
    print_status "Installing Arch Linux dependencies..."
    
    sudo pacman -Sy --noconfirm
    sudo pacman -S --noconfirm base-devel wget curl tar xz gcc make git unzip \
        ncurses openssl zlib python-setuptools clang flex bison gawk gettext rsync file
    
    print_success "Arch Linux dependencies installed"
}

# Install dependencies for openSUSE
install_opensuse_deps() {
    print_status "Installing openSUSE dependencies..."
    
    sudo zypper refresh
    sudo zypper install -y patterns-devel-base-devel_basis wget curl tar xz gcc gcc-c++ \
        make git unzip ncurses-devel openssl-devel zlib-devel python3-setuptools \
        clang flex bison gawk gettext-tools rsync file
    
    print_success "openSUSE dependencies installed"
}

# Main installation function
install_dependencies() {
    case "$OS_ID" in
        "ubuntu"|"debian")
            install_ubuntu_deps
            ;;
        "rhel"|"centos"|"fedora"|"rocky"|"almalinux")
            install_rhel_deps
            ;;
        "arch"|"manjaro")
            install_arch_deps
            ;;
        "opensuse"|"opensuse-leap"|"opensuse-tumbleweed")
            install_opensuse_deps
            ;;
        *)
            print_error "Unsupported distribution: $OS_ID"
            print_status "Please install the following packages manually:"
            echo "  - build-essential / Development Tools"
            echo "  - wget, curl, tar, xz-utils"
            echo "  - gcc, g++, make, git, unzip"
            echo "  - libncurses5-dev, libssl-dev, zlib1g-dev"
            echo "  - python3-setuptools or python3-distutils"
            echo "  - clang, flex, bison, gawk, gettext, rsync"
            exit 1
            ;;
    esac
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    REQUIRED_TOOLS="wget curl tar make gcc g++ git unzip"
    MISSING=""
    
    for tool in $REQUIRED_TOOLS; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            MISSING="$MISSING $tool"
        fi
    done
    
    if [[ -n "$MISSING" ]]; then
        print_error "Missing tools:$MISSING"
        exit 1
    fi
    
    print_success "All required tools are available"
}

# Main execution
main() {
    print_status "OpenWrt Build Environment Setup"
    print_status "==============================="
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        print_error "Don't run this script as root"
        exit 1
    fi
    
    # Check for sudo
    if ! command -v sudo >/dev/null 2>&1; then
        print_error "sudo is required but not installed"
        exit 1
    fi
    
    detect_os
    install_dependencies
    verify_installation
    
    print_success "Build environment setup completed!"
    print_status "You can now build OpenWrt IPK packages:"
    echo "  ./scripts/build-ipk.sh --arch x86_64"
    echo "  make ipk ARCH=mips_24kc"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "OpenWrt Build Environment Setup"
        echo "Usage: $0 [--help]"
        echo ""
        echo "This script automatically detects your Linux distribution"
        echo "and installs the required dependencies for building OpenWrt IPK packages."
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
