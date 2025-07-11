# ispappd Build System

This document describes the build system for ispappd and how object files are organized.

## Build Directories

The build system is configured to place object files and binaries in organized directories:

- `build-output/` - Root build directory
  - `obj/` - Object files (.o)
  - `bin/` - Final binaries
  - `deps/` - Dependency files (.d)

## Build Methods

### Method 1: Using Autotools (Recommended)

1. **Generate configure script** (if not present):
   ```bash
   autoreconf -fiv
   ```

2. **Configure the build**:
   ```bash
   ./configure --prefix=/usr --enable-jsonc --enable-debug
   ```

3. **Build the project**:
   ```bash
   make -j$(nproc)
   ```

4. **Install** (optional):
   ```bash
   make install
   ```

### Method 2: Using the Build Script

A convenient build script is provided:

```bash
./build.sh
```

This script will:
- Create necessary directories
- Generate configure script if needed
- Configure the build
- Compile the project
- Place binaries in the correct location

### Method 3: Using Standalone Makefile

If autotools is not available, use the standalone Makefile:

```bash
make -f Makefile.standalone
```

## Object File Destinations

The build system ensures that:

1. **Object files** are placed in `build-output/obj/`
2. **Binaries** are placed in `build-output/bin/`
3. **Dependency files** are placed in `build-output/deps/`

This keeps the source directory clean and organizes build artifacts.

## Configuration Options

The build system supports several configuration options:

- `--enable-jsonc` - Use json-c library instead of json
- `--enable-debug` - Enable debug messages
- `--enable-devel` - Enable development messages
- `--enable-backupdatainconfig` - Enable backup data in config
- `--with-build-dir=DIR` - Specify custom build directory

## Cleaning

To clean build artifacts:

```bash
# Using autotools
make clean

# Using the clean script
./clean.sh

# Using standalone Makefile
make -f Makefile.standalone clean
```

## Dependencies

The project requires the following libraries:

- libuci
- libubox
- libubus
- libxml2
- libcurl
- json-c (or json)

## Build Targets

Available make targets:

- `all` - Build everything (default)
- `clean` - Clean build artifacts
- `install` - Install binaries
- `uninstall` - Uninstall binaries
- `debug` - Build with debug symbols
- `devel` - Build with development flags
- `info` - Show build information

## Troubleshooting

1. **Missing dependencies**: Install required development packages
2. **Configure script missing**: Run `autoreconf -fiv`
3. **Build directory issues**: Run `./clean.sh` and rebuild
4. **Permission issues**: Check file permissions and ownership

## Examples

### Debug Build
```bash
./configure --enable-debug --enable-devel
make
```

### Custom Build Directory
```bash
./configure --with-build-dir=/tmp/ispappd-build
make
```

### Verbose Build
```bash
make V=1
```
