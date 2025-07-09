#!/bin/bash

# ISPAppD Build Test Script
# Tests the build system to ensure everything works correctly

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[TEST]${NC} $1"; }
print_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
print_error() { echo -e "${RED}[FAIL]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }

FAILED_TESTS=0
TOTAL_TESTS=0

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    print_status "Running: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        print_success "$test_name"
    else
        print_error "$test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

print_status "Starting ISPAppD build system tests..."
echo ""

# Test 1: Check if scripts exist and are executable
print_status "Checking build scripts..."
run_test "IPK build script exists" "test -f scripts/build-ipk.sh"
run_test "IPK build script is executable" "test -x scripts/build-ipk.sh"
run_test "macOS build script exists" "test -f scripts/build-macos.sh"
run_test "macOS build script is executable" "test -x scripts/build-macos.sh"

# Test 2: Check if required source files exist
print_status "Checking source files..."
run_test "Source directory exists" "test -d src"
run_test "Main source file exists" "test -f src/ispappcwmp.c"
run_test "Configure script exists" "test -f configure.ac"
run_test "Makefile template exists" "test -f Makefile.am"

# Test 3: Check OpenWrt files
print_status "Checking OpenWrt files..."
run_test "OpenWrt extension directory exists" "test -d ext/openwrt"
run_test "OpenWrt Makefile exists" "test -f ext/openwrt/build/Makefile"
run_test "OpenWrt Config.in exists" "test -f ext/openwrt/build/Config.in"
run_test "OpenWrt config file exists" "test -f ext/openwrt/config/ispappd"

# Test 4: Check GitHub Actions workflows
print_status "Checking GitHub Actions workflows..."
run_test "GitHub workflows directory exists" "test -d .github/workflows"
run_test "Main build workflow exists" "test -f .github/workflows/build-ipk.yml"
run_test "Dev build workflow exists" "test -f .github/workflows/dev-build.yml"

# Test 5: Test script help functions
print_status "Testing script help functions..."
run_test "IPK script shows help" "./scripts/build-ipk.sh --help"
run_test "IPK script lists architectures" "./scripts/build-ipk.sh --list-archs"
run_test "macOS script shows help" "./scripts/build-macos.sh --help"

# Test 6: Test Makefile wrapper
print_status "Testing Makefile wrapper..."
run_test "Makefile wrapper exists" "test -f Makefile.build"
run_test "Makefile shows help" "make -f Makefile.build help"
run_test "Makefile checks tools" "make -f Makefile.build check-tools"

# Test 7: Test build prerequisites (platform-specific)
print_status "Checking build prerequisites..."
UNAME_S=$(uname -s)

if [[ "$UNAME_S" == "Darwin" ]]; then
    print_status "Testing macOS prerequisites..."
    run_test "Homebrew is available" "command -v brew"
    run_test "macOS has required tools" "command -v gcc && command -v make"
    
    # Test if we can create stub libraries
    print_status "Testing macOS stub library creation..."
    run_test "Can create stub directory" "mkdir -p test-stubs/lib && rmdir test-stubs/lib test-stubs"
    
elif [[ "$UNAME_S" == "Linux" ]]; then
    print_status "Testing Linux prerequisites..."
    run_test "Linux has wget" "command -v wget"
    run_test "Linux has tar" "command -v tar"
    run_test "Linux has make" "command -v make"
    run_test "Linux has gcc" "command -v gcc"
    
    # Test if we can download (without actually downloading)
    print_status "Testing Linux network connectivity..."
    run_test "Can reach OpenWrt downloads" "curl -I https://downloads.openwrt.org/ --connect-timeout 5"
fi

# Test 8: Validate workflow files
print_status "Validating workflow files..."
if command -v yq >/dev/null 2>&1; then
    run_test "Main workflow is valid YAML" "yq eval '.name' .github/workflows/build-ipk.yml"
    run_test "Dev workflow is valid YAML" "yq eval '.name' .github/workflows/dev-build.yml"
else
    print_warning "yq not available, skipping YAML validation"
fi

# Test 9: Test documentation
print_status "Checking documentation..."
run_test "Build documentation exists" "test -f BUILD.md"
run_test "Build status documentation exists" "test -f BUILD_STATUS.md"

# Test 10: Configuration validation
print_status "Testing configuration files..."
run_test "configure.ac is readable" "grep -q 'AC_INIT' configure.ac"
run_test "OpenWrt config is readable" "grep -q 'config local' ext/openwrt/config/ispappd"

echo ""
print_status "Test Summary"
echo "============="
echo "Total tests: $TOTAL_TESTS"
echo "Passed: $((TOTAL_TESTS - FAILED_TESTS))"
echo "Failed: $FAILED_TESTS"

if [[ $FAILED_TESTS -eq 0 ]]; then
    print_success "All tests passed! ✅"
    echo ""
    print_status "Build system is ready to use."
    echo ""
    echo "Next steps:"
    if [[ "$UNAME_S" == "Darwin" ]]; then
        echo "  make macos          # Build for macOS"
        echo "  make macos DEBUG=1  # Debug build"
    else
        echo "  make ipk ARCH=x86_64     # Build IPK"
        echo "  make ipk-all             # Build all architectures"
    fi
    echo "  make setup-dev           # Setup development environment"
    echo ""
    exit 0
else
    print_error "Some tests failed! ❌"
    echo ""
    print_status "Please fix the issues before proceeding:"
    echo "  - Check that all required files are present"
    echo "  - Ensure scripts have correct permissions"
    echo "  - Install missing dependencies with 'make setup-dev'"
    echo ""
    exit 1
fi
