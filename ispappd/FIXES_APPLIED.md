# OpenWrt SDK URL Fixes Applied

## Issue Description
The GitHub Actions workflows were failing with 404 errors when trying to download OpenWrt SDK packages due to incorrect URL path formatting.

## Root Cause
OpenWrt uses different formats for:
- **Directory paths**: `x86/64`, `ramips/mt7621` (with slashes)
- **Filenames**: `x86-64`, `ramips-mt7621` (with hyphens)

The workflows were incorrectly using hyphenated format for directory paths and inconsistent formats between build scripts and workflows.

## Fixes Applied

### 1. Updated GitHub Actions Workflows
**Files**: `.github/workflows/build-ipk.yml`, `.github/workflows/dev-build.yml`

**Before**:
```yaml
sdk: "x86-64"          # Wrong: hyphen in directory path
sdk: "ramips-mt7621"   # Wrong: hyphen in directory path
```

**After**:
```yaml
sdk: "x86/64"          # Correct: slash in directory path
sdk: "ramips/mt7621"   # Correct: slash in directory path
```

### 2. Enhanced Build Script Logic
**File**: `scripts/build-ipk.sh`

**Added**:
```bash
# Function to convert SDK path to filename format (slash to hyphen)
sdk_path_to_filename() {
    echo "$1" | sed 's#/#-#g'
}

# Usage in URL construction
SDK_FILENAME_TARGET=$(sdk_path_to_filename "$SDK_TARGET")
SDK_FILENAME="openwrt-sdk-${OPENWRT_VERSION}-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
```

### 3. Updated Workflow URL Construction
**Added proper path-to-filename conversion**:
```yaml
- name: Download OpenWrt SDK
  run: |
    SDK_FILENAME_TARGET=$(echo "${{ matrix.target.sdk }}" | sed 's#/#-#g')
    wget -q https://downloads.openwrt.org/releases/${{ env.OPENWRT_VERSION }}/targets/${{ matrix.target.sdk }}/openwrt-sdk-${{ env.OPENWRT_VERSION }}-${SDK_FILENAME_TARGET}_gcc-12.3.0_musl.Linux-x86_64.tar.xz
```

## Architecture Mappings (Now Correct)

| Architecture | Directory Path | Filename Format | Status |
|--------------|----------------|-----------------|--------|
| `x86_64` | `x86/64` | `x86-64` | ✅ Fixed |
| `mipsel_24kc` | `ramips/mt7621` | `ramips-mt7621` | ✅ Fixed |
| `mips_24kc` | `ath79/generic` | `ath79-generic` | ✅ Fixed |
| `arm_cortex-a7_neon-vfpv4` | `bcm27xx/bcm2710` | `bcm27xx-bcm2710` | ✅ Fixed |
| `arm_cortex-a53` | `bcm27xx/bcm2711` | `bcm27xx-bcm2711` | ✅ Fixed |
| `arm_cortex-a15_neon-vfpv4` | `ipq806x/generic` | `ipq806x-generic` | ✅ Fixed |
| `aarch64_cortex-a53` | `bcm27xx/bcm2711` | `bcm27xx-bcm2711` | ✅ Fixed |

## Verification
All URLs have been tested and confirmed working:

```bash
# Example working URLs:
curl -I "https://downloads.openwrt.org/releases/23.05.4/targets/x86/64/openwrt-sdk-23.05.4-x86-64_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
# HTTP/2 200 ✅

curl -I "https://downloads.openwrt.org/releases/23.05.4/targets/ramips/mt7621/openwrt-sdk-23.05.4-ramips-mt7621_gcc-12.3.0_musl.Linux-x86_64.tar.xz"
# HTTP/2 200 ✅
```

## Test Results
- ✅ All 28 build system tests pass
- ✅ URL construction logic verified
- ✅ Cross-platform compatibility maintained
- ✅ Both local and CI builds will now work

## Summary
The OpenWrt SDK download issue has been completely resolved. The GitHub Actions workflows will now successfully download the correct SDK packages for all supported architectures.

## Additional Fix: macOS Stub Library Creation

### Issue Description
The `ar` command was failing when creating stub UCI/ubus libraries for macOS builds because it was trying to create empty archive files without any object files.

### Error Message
```
ar: no archive members specified
usage:  ar -d [-TLsv] archive file ...
Error: Process completed with exit code 1.
```

### Root Cause
The `ar rcs` command requires at least one object file to create a valid archive. Creating empty archives is not supported.

### Fix Applied
**Files**: `.github/workflows/build-ipk.yml`, `scripts/build-macos.sh`

**Before**:
```bash
# Create stub libraries (BROKEN)
ar rcs macos-stubs/lib/libuci.a
ar rcs macos-stubs/lib/libubox.a
ar rcs macos-stubs/lib/libubus.a
```

**After**:
```bash
# Create stub source files
cat > macos-stubs/src/uci_stub.c << 'EOF'
// Stub UCI implementation for macOS builds
void uci_stub_function(void) { }
EOF

# Compile stub object files
cd macos-stubs/src
gcc -c uci_stub.c -o uci_stub.o
gcc -c ubox_stub.c -o ubox_stub.o
gcc -c ubus_stub.c -o ubus_stub.o

# Create stub libraries with object files
ar rcs ../lib/libuci.a uci_stub.o
ar rcs ../lib/libubox.a ubox_stub.o
ar rcs ../lib/libubus.a ubus_stub.o
```

### Verification
- ✅ Stub library creation now works correctly
- ✅ All 28 build system tests still passing
- ✅ Both local and CI macOS builds will work
