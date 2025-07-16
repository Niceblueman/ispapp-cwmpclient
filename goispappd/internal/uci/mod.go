package uci

// #cgo LDFLAGS: -luci
// #include <uci.h>
// #include <stdlib.h>
// #include <string.h>
//
// static char* get_option_value(struct uci_option *opt) {
//     if (opt && opt->type == UCI_TYPE_STRING) {
//         return opt->v.string;
//     }
//     return NULL;
// }
//
// static void init_uci_ptr(struct uci_ptr *ptr) {
//     memset(ptr, 0, sizeof(struct uci_ptr));
// }
import "C"
import (
	"errors"
	"unsafe"
)

// UCI represents the UCI context
type UCI struct {
	ctx *C.struct_uci_context
}

// NewUCI creates a new UCI context
func NewUCI() (*UCI, error) {
	ctx := C.uci_alloc_context()
	if ctx == nil {
		return nil, errors.New("failed to allocate UCI context")
	}
	return &UCI{ctx: ctx}, nil
}

// Free releases the UCI context
func (u *UCI) Free() {
	if u.ctx != nil {
		C.uci_free_context(u.ctx)
		u.ctx = nil
	}
}

// Get retrieves a configuration value
func (u *UCI) Get(packageName, section, option string) (string, error) {
	if u.ctx == nil {
		return "", errors.New("UCI context is nil")
	}

	// Construct the full UCI path: package.section.option
	path := packageName + "." + section + "." + option
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var ptr C.struct_uci_ptr
	// Initialize the pointer structure to zero
	C.init_uci_ptr(&ptr)

	// Lookup the complete path
	if C.uci_lookup_ptr(u.ctx, &ptr, cPath, C._Bool(true)) != C.UCI_OK {
		return "", errors.New("failed to lookup UCI path: " + path)
	}

	if ptr.o == nil {
		return "", errors.New("option not found: " + path)
	}

	// Use our C helper function to get the option value
	optValue := C.get_option_value(ptr.o)
	if optValue == nil {
		return "", errors.New("option value is nil: " + path)
	}

	return C.GoString(optValue), nil
}

// Set sets a configuration value
func (u *UCI) Set(packageName, section, option, value string) error {
	if u.ctx == nil {
		return errors.New("UCI context is nil")
	}

	// Construct the full UCI path: package.section.option
	path := packageName + "." + section + "." + option
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	var ptr C.struct_uci_ptr
	// Lookup the complete path
	if C.uci_lookup_ptr(u.ctx, &ptr, cPath, C._Bool(true)) != C.UCI_OK {
		return errors.New("failed to lookup UCI path: " + path)
	}

	ptr.value = cValue
	if C.uci_set(u.ctx, &ptr) != C.UCI_OK {
		return errors.New("failed to set value for: " + path)
	}

	return nil
}

// Commit saves changes to the configuration
func (u *UCI) Commit(packageName string) error {
	if u.ctx == nil {
		return errors.New("UCI context is nil")
	}

	cPackage := C.CString(packageName)
	defer C.free(unsafe.Pointer(cPackage))

	var pkg *C.struct_uci_package
	if C.uci_load(u.ctx, cPackage, &pkg) != C.UCI_OK {
		return errors.New("failed to load package")
	}
	if C.uci_commit(u.ctx, &pkg, C._Bool(false)) != C.UCI_OK {
		return errors.New("failed to commit changes")
	}

	return nil
}
