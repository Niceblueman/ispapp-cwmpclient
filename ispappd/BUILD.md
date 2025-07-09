# ISPAppD Build System

This repository includes comprehensive build workflows for creating OpenWrt IPK packages and macOS ARM64 binaries for the ISPAppD CWMP client.

## Overview
ISPAppD is a CWMP (TR-069) client daemon designed for OpenWrt routers and other embedded systems. This build system provides:

- **GitHub Actions workflows** for automated CI/CD
- **Local build scripts** for development and testing
- **Multi-architecture support** for various OpenWrt targets
- **macOS ARM64 native builds** for development

## Quick Start

### Prerequisites

#### For OpenWrt IPK builds (Linux):
**Automatic setup (recommended for local development):**
```bash
./scripts/setup-build-env.sh
```

**For GitHub Actions/CI (guaranteed compatibility):**
```bash
./scripts/ci-install-deps.sh
```

**Manual setup:**
```bash
# Ubuntu/Debian 22.04+ (current GitHub Actions runners)
sudo apt-get update
sudo apt-get install -y build-essential wget curl tar xz-utils \
    gcc g++ libc6-dev make git unzip libncurses5-dev \
    libssl-dev zlib1g-dev python3-setuptools python3-dev \
    clang flex bison gawk gettext rsync file time

# Ubuntu/Debian 20.04 and older
sudo apt-get install -y python3-distutils instead of python3-setuptools python3-dev
```

#### For macOS ARM64 builds:
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install dependencies
brew install autoconf automake libtool curl json-c pkg-config libmicrohttpd
```

### Building OpenWrt IPK Packages

#### Using the build script (recommended):
```bash
# List available architectures
./scripts/build-ipk.sh --list-archs

# Build for a specific architecture
./scripts/build-ipk.sh --arch x86_64

# Build with debug flags
./scripts/build-ipk.sh --arch mips_24kc --debug

# Clean build with verbose output
./scripts/build-ipk.sh --arch arm_cortex-a53 --clean --verbose

# Keep container running for dependency inspection
./scripts/build-ipk.sh --arch x86_64 --keep-running
```

#### Manual build:
```bash
# Download OpenWrt SDK
wget https://downloads.openwrt.org/releases/23.05.4/targets/x86/64/openwrt-sdk-23.05.4-x86-64_gcc-12.3.0_musl.Linux-x86_64.tar.xz
tar xf openwrt-sdk-*.tar.xz
cd openwrt-sdk-*

# Copy package
mkdir -p package/ispappd
cp -r /path/to/ispapp-cwmpclient/* package/ispappd/
cp package/ispappd/ext/openwrt/build/Makefile package/ispappd/Makefile

# Configure and build
echo "src-link ispappd_local $(pwd)/package" >> feeds.conf.default
./scripts/feeds update -a
./scripts/feeds install -a
make defconfig
echo "CONFIG_PACKAGE_ispappd=m" >> .config
make defconfig
make package/ispappd/compile V=s
```

### Building for macOS ARM64

#### Using the build script (recommended):
```bash
# Standard build
./scripts/build-macos.sh

# Debug build
./scripts/build-macos.sh --debug

# Custom prefix and verbose output
./scripts/build-macos.sh --prefix /usr/local --verbose

# Clean build
./scripts/build-macos.sh --clean --debug
```

#### Manual build:
```bash
# Install dependencies
brew install autoconf automake libtool curl json-c pkg-config libmicrohttpd

# Generate configure script
autoreconf -fiv

# Configure (with stub UCI/ubus libraries)
./configure --prefix=$HOME/ispappd-macos --enable-jsonc \
  --with-uci-include-path=./macos-stubs/include \
  --with-uci-lib-path=./macos-stubs/lib \
  --with-libubox-include-path=./macos-stubs/include \
  --with-libubox-lib-path=./macos-stubs/lib \
  --with-libubus-include-path=./macos-stubs/include \
  --with-libubus-lib-path=./macos-stubs/lib

# Build and install
make -j$(sysctl -n hw.ncpu)
make install
```

## Supported Architectures

### OpenWrt Targets
| Architecture | SDK Target | Description |
|--------------|------------|-------------|
| `mips_24kc` | `ath79-generic` | MIPS 24Kc (TP-Link, etc.) |
| `mipsel_24kc` | `ramips-mt7621` | MIPS 24Kc LE (MediaTek MT7621) |
| `mipsel_74kc` | `ramips-mt7620` | MIPS 74Kc LE (MediaTek MT7620) |
| `arm_cortex-a7_neon-vfpv4` | `bcm27xx-bcm2710` | ARM Cortex-A7 (RPi 2/3) |
| `arm_cortex-a53` | `bcm27xx-bcm2711` | ARM Cortex-A53 (RPi 4) |
| `arm_cortex-a15_neon-vfpv4` | `ipq806x-generic` | ARM Cortex-A15 (Qualcomm IPQ806x) |
| `aarch64_cortex-a53` | `bcm27xx-bcm2711` | ARM64 Cortex-A53 |
| `aarch64_cortex-a72` | `bcm27xx-bcm2711` | ARM64 Cortex-A72 |
| `x86_64` | `x86-64` | x86_64 (Generic PC) |
| `i386` | `x86-generic` | i386 (32-bit PC) |

### Native Targets
- **macOS ARM64**: Native macOS build for Apple Silicon

## GitHub Actions Workflows

### Main Build Workflow (`.github/workflows/build-ipk.yml`)
- **Trigger**: Push, PR, releases, manual dispatch
- **Targets**: All supported OpenWrt architectures + macOS ARM64
- **Features**:
  - Parallel builds for all architectures
  - Automatic artifact upload
  - Release asset publishing
  - Build summaries

### Development Workflow (`.github/workflows/dev-build.yml`)
- **Trigger**: Manual dispatch only
- **Features**:
  - Single architecture builds
  - Debug build option
  - Quick iteration for testing

### Running Workflows
```bash
# Trigger main build (push to main branch)
git push origin main

# Manual trigger with GitHub CLI
gh workflow run build-ipk.yml

# Manual dev build
gh workflow run dev-build.yml -f target_arch=x86_64 -f debug_build=true
```

## Configuration Options

### Build-time Options
- `CONFIG_ISPAPPD_DEBUG=y`: Enable debug output
- `CONFIG_ISPAPPD_DEVEL=y`: Enable development features
- `CONFIG_ISPAPPD_SCRIPTS_FULL=y`: Install all scripts
- `CONFIG_ISPAPPD_DATA_MODEL_TR181=y`: Use TR-181 data model
- `CONFIG_ISPAPPD_DATA_MODEL_TR98=y`: Use TR-098 data model

### Configure Flags
- `--enable-jsonc`: Use json-c library
- `--enable-debug`: Enable debugging messages
- `--enable-devel`: Enable development messages
- `--enable-backupdatainconfig`: Save backup data in config

## Dependencies

### Core Dependencies
- **libcurl**: HTTP client library
- **json-c**: JSON parsing library
- **microxml**: Lightweight XML library

### OpenWrt Dependencies
- **libuci**: UCI configuration library
- **libubox**: OpenWrt utility library
- **libubus**: OpenWrt message bus

### macOS Dependencies
- **Homebrew**: Package manager
- **autotools**: Build system (autoconf, automake, libtool)

## File Structure

```
.
├── .github/workflows/          # GitHub Actions workflows
│   ├── build-ipk.yml          # Main build workflow
│   └── dev-build.yml          # Development workflow
├── scripts/                   # Build scripts
│   ├── build-ipk.sh          # OpenWrt IPK builder
│   └── build-macos.sh        # macOS native builder
├── src/                       # Source code
├── ext/openwrt/              # OpenWrt-specific files
│   ├── build/                # Build configuration
│   ├── config/               # Default configuration
│   ├── init.d/               # Init scripts
│   └── scripts/              # Runtime scripts
├── bin/                      # Binary build configuration
├── configure.ac              # Autoconf configuration
├── Makefile.am              # Automake configuration
└── README.md                # This file
```

## Development

### Setting up Development Environment

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd ispapp-cwmpclient
   ```

2. **For OpenWrt development**:
   ```bash
   # Install dependencies (Ubuntu/Debian)
   sudo apt-get install build-essential wget curl

   # Test build
   ./scripts/build-ipk.sh --arch x86_64 --debug
   ```

3. **For macOS development**:
   ```bash
   # Install Homebrew and dependencies
   brew install autoconf automake libtool curl json-c

   # Test build
   ./scripts/build-macos.sh --debug
   ```

### Testing Builds

### Inspecting Build Dependencies

For analyzing build dependencies and debugging build issues, you can keep the Docker container running after the build completes:

```bash
# Keep container running for inspection
./scripts/build-ipk.sh --arch x86_64 --keep-running

# In another terminal, connect to the running container
docker ps  # Find the container name
docker exec -it ispappd-build-<timestamp> /bin/bash

# Inside the container, you can:
# - Inspect installed packages: opkg list-installed
# - Check build logs: find . -name "*.log" -type f
# - Examine the SDK structure: ls -la /
# - Review package configuration: cat .config
# - Check feed sources: cat feeds.conf.default
```

This is particularly useful for:
- Debugging build failures
- Understanding dependency requirements
- Verifying package configurations
- Inspecting the OpenWrt SDK environment

#### OpenWrt Testing
```bash
# Test in OpenWrt buildroot
cp ispappd_*.ipk /path/to/openwrt/bin/packages/
opkg install ispappd_*.ipk

# Or test in QEMU
qemu-system-x86_64 -kernel openwrt-kernel -initrd openwrt-rootfs
```

#### macOS Testing
```bash
# Test native binary
./bin/ispappd --help

# Test with configuration
./bin/ispappd -c ispappd.conf
```

### Contributing

1. **Create feature branch**:
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Test changes**:
   ```bash
   # Test OpenWrt build
   ./scripts/build-ipk.sh --arch x86_64 --debug

   # Test macOS build
   ./scripts/build-macos.sh --debug
   ```

3. **Submit pull request**:
   - Ensure all builds pass
   - Update documentation if needed
   - Add tests for new features

## Troubleshooting

### Common Issues

#### python3-distutils Package Not Available
```bash
# This error occurs on Ubuntu 22.04+ and Debian 12+ because python3-distutils
# has been removed and replaced with python3-setuptools

# Solutions (in order of preference):

# 1. Use the CI-compatible script (most reliable):
./scripts/ci-install-deps.sh

# 2. Use the smart setup script (detects your system):
./scripts/setup-build-env.sh

# 3. Install manually for modern systems:
sudo apt-get update
sudo apt-get install -y python3-setuptools python3-dev

# 4. For older systems (Ubuntu 20.04 and earlier):
sudo apt-get install -y python3-distutils
```

#### GitHub Actions Specific Issues
```bash
# If you see python3-distutils errors in GitHub Actions:
# 1. The workflows now use scripts/ci-install-deps.sh which avoids this issue
# 2. GitHub Actions uses Ubuntu 22.04+ runners which don't have python3-distutils
# 3. The ci-install-deps.sh script only uses python3-setuptools
```

#### OpenWrt SDK Download Fails
```bash
# Check OpenWrt version and target
curl -I https://downloads.openwrt.org/releases/23.05.4/targets/x86/64/

# Use alternative mirror
export OPENWRT_MIRROR="https://archive.openwrt.org"
```

#### Docker Container Issues
```bash
# If Docker build fails or you need to inspect the build environment:

# Keep container running for debugging
./scripts/build-ipk.sh --arch x86_64 --keep-running

# In another terminal, connect to the running container
docker ps  # Find the container name (ispappd-build-<timestamp>)
docker exec -it ispappd-build-<timestamp> /bin/bash

# Inside the container, useful commands:
opkg list-installed    # See installed packages
cat .config            # View build configuration
find . -name "*.log"   # Find build logs
ls -la bin/packages/   # Check built packages
./scripts/feeds update -a  # Update feeds manually

# Stop the container when done
docker stop ispappd-build-<timestamp>
```

#### OpenWrt SDK Filename Pattern Issues
```bash
# Some OpenWrt targets use different filename patterns
# For example, ipq806x/generic uses musl_eabi instead of musl

# Check actual available SDK files:
curl -s "https://downloads.openwrt.org/releases/23.05.4/targets/ipq806x/generic/" | grep -E 'openwrt-sdk.*\.tar\.xz'

# Common patterns:
# Most targets: openwrt-sdk-23.05.4-target_gcc-12.3.0_musl.Linux-x86_64.tar.xz
# ipq806x/generic: openwrt-sdk-23.05.4-ipq806x-generic_gcc-12.3.0_musl_eabi.Linux-x86_64.tar.xz

# The build scripts automatically handle this, but for manual builds:
SDK_TARGET="ipq806x/generic"
SDK_FILENAME_TARGET=$(echo "$SDK_TARGET" | sed 's#/#-#g')
if [ "$SDK_TARGET" = "ipq806x/generic" ]; then
    SDK_FILENAME="openwrt-sdk-23.05.4-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl_eabi.Linux-x86_64.tar.xz"
else
    SDK_FILENAME="openwrt-sdk-23.05.4-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
fi
```

#### Package Compilation Fails
```bash
# Check if required dependencies are available
cd openwrt-sdk-*
./scripts/feeds search mxml
./scripts/feeds install mxml

# Enable required packages in config
echo "CONFIG_PACKAGE_mxml=y" >> .config
make defconfig

# Build with verbose output to see errors
make package/ispappd/compile V=s
```

#### macOS Build Dependencies Missing
```bash
# Reinstall dependencies
brew update
brew reinstall autoconf automake libtool

# Check pkg-config paths
export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:$PKG_CONFIG_PATH"
```

#### Build Fails with Missing Headers
```bash
# For OpenWrt: Check SDK integrity
tar -tf openwrt-sdk-*.tar.xz | grep -E "(gcc|include)"

# For macOS: Check Xcode tools
xcode-select --install
```

### Debug Mode

Enable verbose logging in builds:
```bash
# OpenWrt
./scripts/build-ipk.sh --arch x86_64 --debug --verbose

# macOS  
./scripts/build-macos.sh --debug --verbose
```

## License

This build system follows the same license as the ISPAppD project. See the main project documentation for license details.

## Support

- **Issues**: Report bugs and feature requests in the GitHub issue tracker
- **Documentation**: See the main project README for ISPAppD-specific documentation
- **Development**: Use the development workflow for testing changes
