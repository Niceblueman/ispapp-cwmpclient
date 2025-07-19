package jobs

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mdlayher/wifi"
)

// TestWiFiInterfacesAndStations tests collecting WiFi interface details and station details
// Note: StationInfo gets devices associated with this interface (when acting as AP)
func TestWiFiInterfacesAndStations(t *testing.T) {
	// Create a new WiFi client
	client, err := wifi.New()
	if err != nil {
		t.Fatalf("Failed to create WiFi client: %v", err)
	}
	defer client.Close()

	// Get all WiFi interfaces
	interfaces, err := client.Interfaces()
	if err != nil {
		t.Fatalf("Failed to get WiFi interfaces: %v", err)
	}

	if len(interfaces) == 0 {
		t.Skip("No WiFi interfaces found - this is expected on systems without WiFi hardware")
	}

	t.Logf("Found %d WiFi interface(s)", len(interfaces))

	// Iterate through each interface and collect details
	for i, iface := range interfaces {
		t.Logf("\n=== Interface %d ===", i+1)
		logInterfaceDetails(t, iface)

		// Get stations associated with this interface
		stations, err := client.StationInfo(iface)
		if err != nil {
			t.Logf("Failed to get station info for interface %s (normal if not in AP mode): %v", iface.Name, err)
			continue
		}

		t.Logf("Found %d station(s) on interface %s", len(stations), iface.Name)
		for j, station := range stations {
			t.Logf("\n--- Station %d on %s ---", j+1, iface.Name)
			logStationDetails(t, station)
		}
	}
}

// TestWiFiScanResults tests scanning for available networks
// Note: BSS() gets scan results of nearby networks, StationInfo() gets associated clients
func TestWiFiScanResults(t *testing.T) {
	client, err := wifi.New()
	if err != nil {
		t.Fatalf("Failed to create WiFi client: %v", err)
	}
	defer client.Close()

	interfaces, err := client.Interfaces()
	if err != nil {
		t.Fatalf("Failed to get WiFi interfaces: %v", err)
	}

	if len(interfaces) == 0 {
		t.Skip("No WiFi interfaces found")
	}

	// Try to scan on the first available interface
	iface := interfaces[1]
	t.Logf("Attempting to scan on interface: %s", iface.Name)

	// Get scan results directly - BSS method can get available networks
	bssEntrie, err := client.BSS(iface)
	if err != nil {
		t.Logf("Failed to get BSS entries (may require root privileges): %v", err)
		// Don't fail the test, as this is expected in many environments
	}
	if bssEntrie == nil {
		t.Logf("No BSS entries found on interface %s", iface.Name)
	}
	t.Logf("Found BSS entry for interface %s", iface.Name)
	logBSSDetails(t, bssEntrie)
	// Also test getting station info (this gets associated stations, not scan results)
	stations, err := client.StationInfo(iface)
	if err != nil {
		t.Logf("Failed to get station info (this is normal if interface is not in AP mode): %v", err)
	} else {
		t.Logf("Found %d associated station(s) on interface %s", len(stations), iface.Name)
		for _, station := range stations {
			t.Logf("Associated station: %s", station.HardwareAddr)
		}
	}
}

// logInterfaceDetails logs detailed information about a WiFi interface
func logInterfaceDetails(t *testing.T, iface *wifi.Interface) {
	t.Logf("Name: %s", iface.Name)
	t.Logf("Index: %d", iface.Index)
	t.Logf("Hardware Address: %s", iface.HardwareAddr)
	t.Logf("PHY: %d", iface.PHY)
	t.Logf("Device: %d", iface.Device)
	t.Logf("Type: %s", iface.Type)
	t.Logf("Frequency: %d MHz", iface.Frequency)
}

// logStationDetails logs detailed information about a station
func logStationDetails(t *testing.T, station *wifi.StationInfo) {
	t.Logf("Hardware Address: %s", station.HardwareAddr)
	t.Logf("Connected Time: %v", station.Connected)
	t.Logf("Inactive Time: %v", station.Inactive)
	t.Logf("Receive Bytes: %d", station.ReceivedBytes)
	t.Logf("Transmit Bytes: %d", station.TransmittedBytes)
	t.Logf("Receive Packets: %d", station.ReceivedPackets)
	t.Logf("Transmit Packets: %d", station.TransmittedPackets)
	t.Logf("Signal: %d dBm", station.Signal)
	t.Logf("Transmit Retries: %d", station.TransmitRetries)
	t.Logf("Transmit Failed: %d", station.TransmitFailed)
	t.Logf("Beacon Loss: %d", station.BeaconLoss)

	if station.ReceiveBitrate != 0 {
		t.Logf("Receive Bitrate: %d kbps", station.ReceiveBitrate)
	}
	if station.TransmitBitrate != 0 {
		t.Logf("Transmit Bitrate: %d kbps", station.TransmitBitrate)
	}
}

// logBSSDetails logs detailed information about a BSS (Basic Service Set)
func logBSSDetails(t *testing.T, bss *wifi.BSS) {
	if bss != nil {
		t.Logf("BSSID: %s", bss.BSSID)
		t.Logf("SSID: %s", bss.SSID)
		t.Logf("Frequency: %d MHz", bss.Frequency)
		t.Logf("Beacon Interval: %v", bss.BeaconInterval)
		t.Logf("Last Seen: %v", bss.LastSeen)
		t.Logf("Status: %s", bss.Status)
	} else {
		t.Log("No BSS information available")
		return
	}
}

// TestWiFiCapabilities tests getting WiFi capabilities
func TestWiFiCapabilities(t *testing.T) {
	client, err := wifi.New()
	if err != nil {
		t.Fatalf("Failed to create WiFi client: %v", err)
	}
	defer client.Close()

	interfaces, err := client.Interfaces()
	if err != nil {
		t.Fatalf("Failed to get WiFi interfaces: %v", err)
	}

	if len(interfaces) == 0 {
		t.Skip("No WiFi interfaces found")
	}

	for _, iface := range interfaces {
		t.Logf("\n=== Capabilities for %s ===", iface.Name)

		// This would show what the interface is capable of
		t.Logf("Interface Type: %s", iface.Type)
		t.Logf("Current Frequency: %d MHz", iface.Frequency)
		t.Logf("PHY Index: %d", iface.PHY)
		t.Logf("Device Index: %d", iface.Device)
		t.Logf("Hardware Address: %s", iface.HardwareAddr)

		// Additional interface information
		if iface.HardwareAddr != nil {
			t.Logf("MAC Address: %s", iface.HardwareAddr.String())
		}
	}
}

// Helper function to run WiFi collection in production code
func CollectWiFiInfo() (*WiFiInfo, error) {
	client, err := wifi.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create WiFi client: %w", err)
	}
	defer client.Close()

	info := &WiFiInfo{
		Interfaces: make([]InterfaceInfo, 0),
		Timestamp:  time.Now(),
	}

	interfaces, err := client.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range interfaces {
		ifaceInfo := InterfaceInfo{
			Name:         iface.Name,
			Index:        iface.Index,
			HardwareAddr: iface.HardwareAddr.String(),
			PHY:          iface.PHY,
			Device:       iface.Device,
			Type:         iface.Type.String(),
			Frequency:    iface.Frequency,
			Stations:     make([]StationInfo, 0),
		}

		// Get stations for this interface
		stations, err := client.StationInfo(iface)
		if err != nil {
			log.Printf("Failed to get station info for %s: %v", iface.Name, err)
		} else {
			for _, station := range stations {
				stationInfo := StationInfo{
					HardwareAddr:    station.HardwareAddr.String(),
					ConnectedTime:   station.Connected,
					InactiveTime:    station.Inactive,
					ReceiveBytes:    station.ReceivedBytes,
					TransmitBytes:   station.TransmittedBytes,
					ReceivePackets:  station.ReceivedPackets,
					TransmitPackets: station.TransmittedPackets,
					Signal:          station.Signal,
					TransmitRetries: station.TransmitRetries,
					TransmitFailed:  station.TransmitFailed,
					BeaconLoss:      station.BeaconLoss,
				}

				ifaceInfo.Stations = append(ifaceInfo.Stations, stationInfo)
			}
		}

		info.Interfaces = append(info.Interfaces, ifaceInfo)
	}

	return info, nil
}

// Data structures for WiFi information
type WiFiInfo struct {
	Interfaces []InterfaceInfo `json:"interfaces"`
	Timestamp  time.Time       `json:"timestamp"`
}

type InterfaceInfo struct {
	Name         string        `json:"name"`
	Index        int           `json:"index"`
	HardwareAddr string        `json:"hardware_addr"`
	PHY          int           `json:"phy"`
	Device       int           `json:"device"`
	Type         string        `json:"type"`
	Frequency    int           `json:"frequency"`
	Stations     []StationInfo `json:"stations"`
}

type StationInfo struct {
	HardwareAddr    string        `json:"hardware_addr"`
	ConnectedTime   time.Duration `json:"connected_time"`
	InactiveTime    time.Duration `json:"inactive_time"`
	ReceiveBytes    int           `json:"receive_bytes"`
	TransmitBytes   int           `json:"transmit_bytes"`
	ReceivePackets  int           `json:"receive_packets"`
	TransmitPackets int           `json:"transmit_packets"`
	Signal          int           `json:"signal"`
	TransmitRetries int           `json:"transmit_retries"`
	TransmitFailed  int           `json:"transmit_failed"`
	BeaconLoss      int           `json:"beacon_loss"`
	ReceiveBitrate  int           `json:"receive_bitrate,omitempty"`
	TransmitBitrate int           `json:"transmit_bitrate,omitempty"`
}
