package nl80211

// #cgo LDFLAGS: -lnl-3 -lnl-genl-3
// #include <netlink/netlink.h>
// #include <netlink/genl/genl.h>
// #include <netlink/genl/ctrl.h>
// #include <linux/nl80211.h>
// #include <stdlib.h>
// #include <string.h>
// #include <net/if.h>
//
// struct station_info {
//     char mac[18];
//     int signal;
//     unsigned int inactive_ms;
// };
//
// struct station_cb_data {
//     struct station_info *stations;
//     int count;
//     int max;
// };
//
// static int station_callback(struct nl_msg *msg, void *arg) {
//     struct station_cb_data *data = (struct station_cb_data *)arg;
//     struct genlmsghdr *gnlh = nlmsg_data(nlmsg_hdr(msg));
//     struct nlattr *tb[NL80211_ATTR_MAX + 1];
//     struct nlattr *sinfo[NL80211_STA_INFO_MAX + 1];
//
//     nla_parse(tb, NL80211_ATTR_MAX, genlmsg_attrdata(gnlh, 0), genlmsg_attrlen(gnlh, 0), NULL);
//
//     if (tb[NL80211_ATTR_STA_INFO]) {
//         if (nla_parse_nested(sinfo, NL80211_STA_INFO_MAX, tb[NL80211_ATTR_STA_INFO], NULL) < 0) {
//             return NL_SKIP;
//         }
//
//         if (data->count < data->max) {
//             struct station_info *info = &data->stations[data->count];
//
//             if (tb[NL80211_ATTR_MAC]) {
//                 unsigned char *mac = nla_data(tb[NL80211_ATTR_MAC]);
//                 snprintf(info->mac, sizeof(info->mac), "%02x:%02x:%02x:%02x:%02x:%02x",
//                          mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
//             }
//
//             if (sinfo[NL80211_STA_INFO_SIGNAL]) {
//                 info->signal = (int8_t)nla_get_u8(sinfo[NL80211_STA_INFO_SIGNAL]);
//             }
//
//             if (sinfo[NL80211_STA_INFO_INACTIVE_TIME]) {
//                 info->inactive_ms = nla_get_u32(sinfo[NL80211_STA_INFO_INACTIVE_TIME]);
//             }
//
//             data->count++;
//         }
//     }
//     return NL_OK;
// }
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Station represents a wireless station's information
type Station struct {
	MAC        string
	Signal     int
	InactiveMs uint32
}

// NL80211 represents the netlink context for nl80211
type NL80211 struct {
	sock     *C.struct_nl_sock
	familyID C.int
}

// NewNL80211 creates a new nl80211 context
func NewNL80211() (*NL80211, error) {
	sock := C.nl_socket_alloc()
	if sock == nil {
		return nil, errors.New("failed to allocate netlink socket")
	}

	if ret := C.genl_connect(sock); ret != 0 {
		C.nl_socket_free(sock)
		return nil, fmt.Errorf("failed to connect to generic netlink: %d", ret)
	}

	familyID := C.genl_ctrl_resolve(sock, C.CString("nl80211"))
	if familyID < 0 {
		C.nl_socket_free(sock)
		return nil, fmt.Errorf("failed to resolve nl80211 family: %d", familyID)
	}

	return &NL80211{
		sock:     sock,
		familyID: familyID,
	}, nil
}

// Free releases the netlink context
func (nl *NL80211) Free() {
	if nl.sock != nil {
		C.nl_socket_free(nl.sock)
		nl.sock = nil
	}
}

// GetStations retrieves station information for a given wireless interface
func (nl *NL80211) GetStations(ifname string, maxStations int) ([]Station, error) {
	if nl.sock == nil {
		return nil, errors.New("netlink socket is nil")
	}

	if maxStations <= 0 {
		return nil, errors.New("maxStations must be positive")
	}

	// Get interface index
	cIfname := C.CString(ifname)
	defer C.free(unsafe.Pointer(cIfname))
	ifindex := C.if_nametoindex(cIfname)
	if ifindex == 0 {
		return nil, fmt.Errorf("failed to get interface index for %s", ifname)
	}

	// Allocate message
	msg := C.nlmsg_alloc()
	if msg == nil {
		return nil, errors.New("failed to allocate netlink message")
	}
	defer C.nlmsg_free(msg)

	// Setup message for station dump
	C.genlmsg_put(msg, C.NL_AUTO_PID, C.NL_AUTO_SEQ, nl.familyID, 0, C.NLM_F_DUMP, C.NL80211_CMD_GET_STATION, 0)
	if ret := C.nla_put_u32(msg, C.NL80211_ATTR_IFINDEX, C.uint32_t(ifindex)); ret < 0 {
		return nil, fmt.Errorf("failed to add interface index: %d", ret)
	}

	// Allocate station info array
	stations := (*C.struct_station_info)(C.calloc(C.size_t(maxStations), C.size_t(unsafe.Sizeof(C.struct_station_info{}))))
	if stations == nil {
		return nil, errors.New("failed to allocate station info array")
	}
	defer C.free(unsafe.Pointer(stations))

	// Setup callback data
	data := C.struct_station_cb_data{
		stations: stations,
		count:    0,
		max:      C.int(maxStations),
	}

	// Set callback
	if ret := C.nl_socket_modify_cb(nl.sock, C.NL_CB_VALID, C.NL_CB_CUSTOM, (*[0]byte)(C.station_callback), unsafe.Pointer(&data)); ret < 0 {
		return nil, fmt.Errorf("failed to set callback: %d", ret)
	}

	// Send and receive message
	if ret := C.nl_send_auto(nl.sock, msg); ret < 0 {
		return nil, fmt.Errorf("failed to send message: %d", ret)
	}
	if ret := C.nl_recvmsgs_default(nl.sock); ret < 0 {
		return nil, fmt.Errorf("failed to receive messages: %d", ret)
	}

	// Convert results to Go slice
	result := make([]Station, data.count)
	for i := 0; i < int(data.count); i++ {
		info := (*C.struct_station_info)(unsafe.Pointer(uintptr(unsafe.Pointer(stations)) + uintptr(i)*unsafe.Sizeof(C.struct_station_info{})))
		result[i] = Station{
			MAC:        C.GoString(&info.mac[0]),
			Signal:     int(info.signal),
			InactiveMs: uint32(info.inactive_ms),
		}
	}

	return result, nil
}
