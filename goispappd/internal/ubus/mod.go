package ubus

// #cgo LDFLAGS: -lubus -lubox -lblobmsg_json -ljson-c
// #include <libubus.h>
// #include <libubox/blobmsg.h>
// #include <libubox/blobmsg_json.h>
// #include <stdlib.h>
// #include <string.h>
//
// void lookup_cb(struct ubus_request *req, int type, struct blob_attr *msg) {
//     char **result = (char **)req->priv;
//     if (!msg) return;
//     char *str = blobmsg_format_json(msg, true);
//     if (str) {
//         *result = str; // Store the JSON string in the priv pointer
//     }
// }
import "C"
import (
	"errors"
	"sync"
	"unsafe"
)

// Global registry for lookup callbacks
var (
	lookupCallbacks = make(map[uintptr]*lookupCallbackData)
	lookupMutex     sync.RWMutex
	nextCallbackID  uintptr = 1
)

// lookupCallbackData holds the data for a lookup callback
type lookupCallbackData struct {
	objects []UbusObjectData
	done    chan struct{}
}

// UbusObjectData represents information about a ubus object
type UbusObjectData struct {
	ID        uint32 `json:"id"`
	TypeID    uint32 `json:"type_id"`
	Path      string `json:"path"`
	Signature string `json:"signature,omitempty"`
}

// Ubus represents the ubus context
type Ubus struct {
	ctx *C.struct_ubus_context
}

// NewUbus creates a new ubus context
func NewUbus() (*Ubus, error) {
	ctx := C.ubus_connect(nil)
	if ctx == nil {
		return nil, errors.New("failed to connect to ubus")
	}
	return &Ubus{ctx: ctx}, nil
}

// Free releases the ubus context
func (u *Ubus) Free() {
	if u.ctx != nil {
		C.ubus_free(u.ctx)
		u.ctx = nil
	}
}

// Call invokes a ubus method
func (u *Ubus) Call(object, method string, args map[string]interface{}) (map[string]interface{}, error) {
	if u.ctx == nil {
		return nil, errors.New("ubus context is nil")
	}

	cObject := C.CString(object)
	defer C.free(unsafe.Pointer(cObject))

	cMethod := C.CString(method)
	defer C.free(unsafe.Pointer(cMethod))

	var blob C.struct_blob_buf
	C.blob_buf_init(&blob, 0)
	defer C.blob_buf_free(&blob)

	// Convert args to blob
	for key, value := range args {
		cKey := C.CString(key)
		defer C.free(unsafe.Pointer(cKey))

		switch v := value.(type) {
		case string:
			cVal := C.CString(v)
			C.blobmsg_add_string(&blob, cKey, cVal)
			C.free(unsafe.Pointer(cVal))
		case int:
			C.blobmsg_add_u32(&blob, cKey, C.uint32_t(v))
		case bool:
			if v {
				C.blobmsg_add_u8(&blob, cKey, C.uint8_t(1))
			} else {
				C.blobmsg_add_u8(&blob, cKey, C.uint8_t(0))
			}
		}
	}

	id := C.uint32_t(0)
	if C.ubus_lookup_id(u.ctx, cObject, &id) != 0 {
		return nil, errors.New("failed to lookup object ID for: " + object)
	}

	// Simple result handling - just return empty map for now
	resultMap := make(map[string]interface{})

	// Try to call the method, but don't expect complex result parsing yet
	ret := C.ubus_invoke(u.ctx, id, cMethod, blob.head, nil, nil, 1000)
	if ret != 0 {
		return nil, errors.New("failed to invoke method")
	}

	return resultMap, nil
}

func (u *Ubus) List() ([]string, error) {
	// ubus_lookup function with a NULL path to list all objects
	ctx := C.ubus_connect(nil)
	if ctx == nil {
		return nil, errors.New("failed to connect to ubus")
	}
	defer C.ubus_free(ctx)

	var results []string
	// Channel to collect results from callback
	resultChan := make(chan string, 100)

	// Prepare a pointer to store the result
	var cResult *C.char

	// Call ubus_lookup with NULL path to get all objects using the C callback
	ret := C.ubus_lookup(ctx, nil, (C.ubus_lookup_handler_t)(unsafe.Pointer(C.lookup_cb)), unsafe.Pointer(&cResult))
	if ret != 0 {
		return nil, errors.New(C.GoString(C.ubus_strerror(ret)))
	}

	// If we got a result, convert it to Go string and add to results
	if cResult != nil {
		results = append(results, C.GoString(cResult))
		C.free(unsafe.Pointer(cResult)) // Free the C string
	}

	// Close the channel to signal completion
	close(resultChan)

	// Collect results from the channel
	for result := range resultChan {
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, errors.New("no objects found")
	}

	return results, nil
}

// Subscribe subscribes to a ubus event
func (u *Ubus) Subscribe(object string) error {
	if u.ctx == nil {
		return errors.New("ubus context is nil")
	}

	cObject := C.CString(object)
	defer C.free(unsafe.Pointer(cObject))

	id := C.uint32_t(0)
	if C.ubus_lookup_id(u.ctx, cObject, &id) != 0 {
		return errors.New("failed to lookup object ID")
	}

	var subscriber C.struct_ubus_subscriber
	if C.ubus_subscribe(u.ctx, &subscriber, id) != 0 {
		return errors.New("failed to subscribe to object")
	}

	return nil
}

// Listen starts the ubus event loop
func (u *Ubus) Listen() error {
	if u.ctx == nil {
		return errors.New("ubus context is nil")
	}

	C.ubus_handle_event(u.ctx)
	return nil
}
