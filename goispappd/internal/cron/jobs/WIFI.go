package jobs

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Niceblueman/goispappd/internal/exec"
	"github.com/Niceblueman/goispappd/internal/uci"
)

// the job will collect from openwrt and save into uci with key/value pairs under wifi section
// Device.WiFi. is the root object for all WiFi related parameters in the data model:

// Device.WiFi.
//
//	SSIDNumberOfEntries                    type: uint32
//	SSID.{i}.
//	    Enable                             type: bool, access: W, default: "false"
//	    Status                             type: enum, default: "Down"
//	    LowerLayers                        type: list<strongRef>(1024), access: W, default: ""
//	    BSSID                              type: MACAddress
//	    MACAddress                         type: MACAddress
//	    SSID                               type: string(32), access: W
//	    Stats.
//	        BytesSent                      type: uint64, flags: deny-active-notif
//	        BytesReceived                  type: uint64, flags: deny-active-notif
//	        PacketsSent                    type: uint64, flags: deny-active-notif
//	        PacketsReceived                type: uint64, flags: deny-active-notif
//	        ErrorsSent                     type: uint32
//	        ErrorsReceived                 type: uint32, flags: deny-active-notif
//	        DiscardPacketsSent             type: uint32, flags: deny-active-notif
//	        DiscardPacketsReceived         type: uint32, flags: deny-active-notif
func SSIDCollectCmd(executor *exec.Executor) *error {
	_package := "Device"
	uci, err := uci.LoadConfig("/etc/config/tr069", &_package)
	if err != nil {
		return &err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if executor == nil {
		_err := fmt.Errorf("failed to create executor")
		return &_err
	}
	// Get all wifi interfaces from UCI
	interfaces, err := getWiFiInterfaces(ctx, executor, uci)
	if err != nil {
		return &err
	}

	// Save SSID count - Device.WiFi.SSIDNumberOfEntries
	err = uci.Set("WiFi", "SSIDNumberOfEntries", fmt.Sprintf("%d", len(interfaces)), false)

	if err != nil {
		log.Printf("Failed to set SSID count: %v", err)
	}

	// Process each interface - Device.WiFi.SSID.{i}.
	for i, iface := range interfaces {
		index := i + 1
		sectionName := fmt.Sprintf("SSID.%d.", index)

		// Basic interface properties following TR-069 naming
		uci.Set("WIFI", fmt.Sprintf("%sEnable", sectionName), fmt.Sprintf("%t", iface.Enable), false)
		uci.Set("WIFI", fmt.Sprintf("%sStatus", sectionName), iface.Status, false)
		uci.Set("WIFI", fmt.Sprintf("%sSSID", sectionName), iface.SSID, false)
		uci.Set("WIFI", fmt.Sprintf("%sBSSID", sectionName), iface.BSSID, false)
		uci.Set("WIFI", fmt.Sprintf("%sMACAddress", sectionName), iface.MACAddress, false)
		uci.Set("WIFI", fmt.Sprintf("%sLowerLayers", sectionName), fmt.Sprintf("Device.WiFi.Radio.%s.", iface.Device), false)

		// Get interface statistics
		stats, err := getInterfaceStats(ctx, executor, iface.Name)
		if err != nil {
			log.Printf("Failed to get stats for interface %s: %v", iface.Name, err)
			continue
		}

		// Save statistics - Device.WiFi.SSID.{i}.Stats.
		uci.Set("WIFI", fmt.Sprintf("%sStats.BytesSent", sectionName), fmt.Sprintf("%d", stats.BytesSent), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.BytesReceived", sectionName), fmt.Sprintf("%d", stats.BytesReceived), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.PacketsSent", sectionName), fmt.Sprintf("%d", stats.PacketsSent), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.PacketsReceived", sectionName), fmt.Sprintf("%d", stats.PacketsReceived), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.ErrorsSent", sectionName), fmt.Sprintf("%d", stats.ErrorsSent), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.ErrorsReceived", sectionName), fmt.Sprintf("%d", stats.ErrorsReceived), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.DiscardPacketsSent", sectionName), fmt.Sprintf("%d", stats.DiscardPacketsSent), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.DiscardPacketsReceived", sectionName), fmt.Sprintf("%d", stats.DiscardPacketsReceived), false)
	}

	// Commit changes
	err = uci.Save()
	if err != nil {
		return &err
	}
	return nil
}

// RadioNumberOfEntries                   type: uint32
// Radio.{i}.
//
//	Enable                             type: bool, access: W
//	Status                             type: enum
//	LowerLayers                        type: list<strongRef>(1024)
//	SupportedFrequencyBands            type: list<enum>
//	OperatingFrequencyBand             type: string, access: W
//	SupportedStandards                 type: list<enum>
//	OperatingStandards                 type: list<string>, access: W
//	PossibleChannels                   type: list<string>(1024)
//	Channel                            type: uint32[1:255], access: W
//	AutoChannelSupported               type: bool
//	AutoChannelEnable                  type: bool, access: W
//	X_ISAPP_SkipDFSChannels         type: enum, access: W
//	Stats.
//	    Noise                          type: int32
//	X_ISPAPP_Stats.
//	    OverallTxCCQ                   type: uint32[:100], flags: deny-active-notif
func RadiosCollectCmd(executor *exec.Executor) {
	section := "Device"
	uci, err := uci.LoadConfig("/etc/config/tr069", &section)
	if err != nil {
		log.Printf("Failed to create UCI context: %v", err)
		return
	}

	if executor == nil {
		log.Printf("Failed to create executor")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Get all WiFi radios
	radios, err := getWiFiRadios(ctx, executor, uci)
	if err != nil {
		log.Printf("Failed to get WiFi radios: %v", err)
		return
	}

	// Save radio count - Device.WiFi.RadioNumberOfEntries
	err = uci.Set("WIFI", "RadioNumberOfEntries", fmt.Sprintf("%d", len(radios)), false)
	if err != nil {
		log.Printf("Failed to set radio count: %v", err)
	}

	// Process each radio - Device.WiFi.Radio.{i}.
	for i, radio := range radios {
		index := i + 1
		sectionName := fmt.Sprintf("Radio.%d.", index)

		// Basic radio properties following TR-069 naming
		uci.Set("WIFI", fmt.Sprintf("%sEnable", sectionName), fmt.Sprintf("%t", radio.Enable), false)
		uci.Set("WIFI", fmt.Sprintf("%sStatus", sectionName), radio.Status, false)
		uci.Set("WIFI", fmt.Sprintf("%sName", sectionName), radio.Name, false)
		uci.Set("WIFI", fmt.Sprintf("%sSupportedFrequencyBands", sectionName), radio.SupportedFrequencyBands, false)
		uci.Set("WIFI", fmt.Sprintf("%sOperatingFrequencyBand", sectionName), radio.OperatingFrequencyBand, false)
		uci.Set("WIFI", fmt.Sprintf("%sSupportedStandards", sectionName), radio.SupportedStandards, false)
		uci.Set("WIFI", fmt.Sprintf("%sOperatingStandards", sectionName), radio.OperatingStandards, false)
		uci.Set("WIFI", fmt.Sprintf("%sPossibleChannels", sectionName), radio.PossibleChannels, false)
		uci.Set("WIFI", fmt.Sprintf("%sChannel", sectionName), fmt.Sprintf("%d", radio.Channel), false)
		uci.Set("WIFI", fmt.Sprintf("%sAutoChannelSupported", sectionName), fmt.Sprintf("%t", radio.AutoChannelSupported), false)
		uci.Set("WIFI", fmt.Sprintf("%sAutoChannelEnable", sectionName), fmt.Sprintf("%t", radio.AutoChannelEnable), false)
		uci.Set("WIFI", fmt.Sprintf("%sStats.Noise", sectionName), fmt.Sprintf("%d", radio.Noise), false)
	}

	// Commit changes
	err = uci.Save()
	if err != nil {
		log.Printf("Failed to commit UCI changes: %v", err)
	}
}

// AccessPointNumberOfEntries             type: uint32
// AccessPoint.{i}.
//     Enable                             type: bool, access: W, default: "false"
//     Status                             type: enum, default: "Disabled"
//     SSIDReference                      type: strongRef(256), access: W, default: ""
//     SSIDAdvertisementEnabled           type: bool, access: W, default: "true"
//     AssociatedDeviceNumberOfEntries    type: uint32
//     Security.
//         ModesSupported                 type: list<enum>, default: "None,WPA-Personal,WPA2-Personal,WPA-WPA2-Personal,WPA-Enterprise,WPA2-Enterprise,WPA-WPA2-Enterprise,X_ISPAPP_Specific"
//         ModeEnabled                    type: string, access: W, default: "None"
//         KeyPassphrase                  type: string(8:63), access: W, flags: hidden
//     AssociatedDevice.{i}.
//         MACAddress                     type: MACAddress, flags: deny-active-notif
//         AuthenticationState            type: bool, flags: deny-active-notif
//         SignalStrength                 type: int32[-200:]
//         Stats.
//             BytesSent                  type: StatsCounter64
//             BytesReceived              type: StatsCounter64
//             PacketsSent                type: StatsCounter64
//             PacketsReceived            type: StatsCounter64
//         X_ISPAPP_Stats.
//             TxFrames                   type: StatsCounter64, flags: deny-active-notif
//             RxFrames                   type: StatsCounter64, flags: deny-active-notif
//             TxFrameBytes               type: StatsCounter64, flags: deny-active-notif
//             RxFrameBytes               type: StatsCounter64, flags: deny-active-notif
//             TxHwFrames                 type: StatsCounter64, flags: deny-active-notif
//             RxHwFrames                 type: StatsCounter64, flags: deny-active-notif
//             TxHwFrameBytes             type: StatsCounter64, flags: deny-active-notif
//             RxHwFrameBytes             type: StatsCounter64, flags: deny-active-notif
//             TxCCQ                      type: uint32[:100], flags: deny-active-notif
//             RxCCQ                      type: uint32[:100], flags: deny-active-notif
//             SignalToNoise              type: int32, flags: deny-active-notif
//             RxRate                     type: string, flags: deny-active-notif
//             TxRate                     type: string, flags: deny-active-notif
//             LastActivity               type: uint32, flags: deny-active-notif
//             SignalStrengthCh0          type: int32, flags: deny-active-notif
//             SignalStrengthCh1          type: int32, flags: deny-active-notif
//             StrengthAtRates            type: string, flags: deny-active-notif
//             UpTime                     type: uint32, flags: deny-active-notif

func AccessPointCollectCmd(executor *exec.Executor) {
	_package := "Device"
	uci, err := uci.LoadConfig("/etc/config/tr069", &_package)
	if err != nil {
		log.Printf("Failed to create UCI context: %v", err)
		return
	}
	if executor == nil {
		log.Printf("Failed to create executor")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Get all WiFi interfaces (access points)
	interfaces, err := getWiFiInterfaces(ctx, executor, uci)
	if err != nil {
		log.Printf("Failed to get WiFi interfaces: %v", err)
		return
	}

	// Filter for AP interfaces only
	var accessPoints []WiFiInterface
	for _, iface := range interfaces {
		// Check if this is an AP interface by checking mode using executor
		mode := getWirelessInterfaceMode(ctx, executor, iface.Name)
		if mode == "ap" {
			accessPoints = append(accessPoints, iface)
		}
	}

	// Save access point count - Device.WiFi.AccessPointNumberOfEntries
	err = uci.Set("WiFi", "AccessPointNumberOfEntries", fmt.Sprintf("%d", len(accessPoints)), false)
	if err != nil {
		log.Printf("Failed to set access point count: %v", err)
	}

	// Process each access point - Device.WiFi.AccessPoint.{i}.
	for i, ap := range accessPoints {
		index := i + 1
		sectionName := fmt.Sprintf("AccessPoint.%d.", index)

		// Basic access point properties following TR-069 naming
		uci.Set("WIFI", fmt.Sprintf("%sEnable", sectionName), fmt.Sprintf("%t", ap.Enable), false)
		uci.Set("WIFI", fmt.Sprintf("%sStatus", sectionName), getAPStatus(ap.Status), false)
		uci.Set("WIFI", fmt.Sprintf("%sSSIDReference", sectionName), fmt.Sprintf("Device.WiFi.SSID.%d.", index), false)
		uci.Set("WIFI", fmt.Sprintf("%sSSIDAdvertisementEnabled", sectionName), "true", false) // Default value

		// Get security configuration
		encryption, keyPassphrase := getSecurityConfig(ctx, executor, ap.Name)

		// Save security configuration - Device.WiFi.AccessPoint.{i}.Security.
		uci.Set("WIFI", fmt.Sprintf("%sSecurity.ModesSupported", sectionName), "None,WPA-Personal,WPA2-Personal,WPA-WPA2-Personal,WPA-Enterprise,WPA2-Enterprise,WPA-WPA2-Enterprise", false)
		uci.Set("WIFI", fmt.Sprintf("%sSecurity.ModeEnabled", sectionName), encryption, false)
		if keyPassphrase != "" {
			uci.Set("WIFI", fmt.Sprintf("%sSecurity.KeyPassphrase", sectionName), keyPassphrase, false)
		}

		// Get associated devices
		associatedDevices, err := getAssociatedDevices(ctx, executor, ap.Name)
		if err != nil {
			log.Printf("Failed to get associated devices for %s: %v", ap.Name, err)
			associatedDevices = []AssociatedDevice{} // Continue with empty list
		}

		// Save associated device count - Device.WiFi.AccessPoint.{i}.AssociatedDeviceNumberOfEntries
		uci.Set("WIFI", fmt.Sprintf("%sAssociatedDeviceNumberOfEntries", sectionName), fmt.Sprintf("%d", len(associatedDevices)), false)

		// Process each associated device - Device.WiFi.AccessPoint.{i}.AssociatedDevice.{j}.
		for j, device := range associatedDevices {
			deviceIndex := j + 1
			deviceSectionName := fmt.Sprintf("AccessPoint.%d.AssociatedDevice.%d.", index, deviceIndex)

			// Basic device properties following TR-069 naming
			uci.Set("WIFI", fmt.Sprintf("%sMACAddress", deviceSectionName), device.MACAddress, false)
			uci.Set("WIFI", fmt.Sprintf("%sAuthenticationState", deviceSectionName), fmt.Sprintf("%t", device.AuthenticationState), false)
			uci.Set("WIFI", fmt.Sprintf("%sSignalStrength", deviceSectionName), fmt.Sprintf("%d", device.SignalStrength), false)

			// Save device statistics - Device.WiFi.AccessPoint.{i}.AssociatedDevice.{j}.Stats.
			uci.Set("WIFI", fmt.Sprintf("%sStats.BytesSent", deviceSectionName), fmt.Sprintf("%d", device.Stats.BytesSent), false)
			uci.Set("WIFI", fmt.Sprintf("%sStats.BytesReceived", deviceSectionName), fmt.Sprintf("%d", device.Stats.BytesReceived), false)
			uci.Set("WIFI", fmt.Sprintf("%sStats.PacketsSent", deviceSectionName), fmt.Sprintf("%d", device.Stats.PacketsSent), false)
			uci.Set("WIFI", fmt.Sprintf("%sStats.PacketsReceived", deviceSectionName), fmt.Sprintf("%d", device.Stats.PacketsReceived), false)

			// Save extended statistics - Device.WiFi.AccessPoint.{i}.AssociatedDevice.{j}.X_ISPAPP_Stats.
			uci.Set("WIFI", fmt.Sprintf("%sX_ISPAPP_Stats.RxRate", deviceSectionName), device.RxRate, false)
			uci.Set("WIFI", fmt.Sprintf("%sX_ISPAPP_Stats.TxRate", deviceSectionName), device.TxRate, false)
			uci.Set("WIFI", fmt.Sprintf("%sX_ISPAPP_Stats.LastActivity", deviceSectionName), fmt.Sprintf("%d", device.LastActivity), false)
		}
	}

	// Commit changes
	err = uci.Save()
	if err != nil {
		log.Printf("Failed to commit UCI changes: %v", err)
	}
}

// NeighboringWiFiDiagnostic.
//     DiagnosticsState                   type: DiagnosticsState, access: W
//     ResultNumberOfEntries              type: uint32
//     Result.{i}.
//         Radio                          type: strongRef
//         SSID                           type: string(32)
//         BSSID                          type: MACAddress
//         Channel                        type: uint32[1:255]
//         SignalStrength                 type: int32[-200:]
//         OperatingFrequencyBand         type: enum
//         OperatingStandards             type: list<string>
//         OperatingChannelBandwidth      type: enum
//         Noise                          type: int32[-200:]

func NeighboringWiFiCollectCmd(executor *exec.Executor) {
	_package := "Device"
	uci, err := uci.LoadConfig("/etc/config/tr069", &_package)
	if err != nil {
		log.Printf("Failed to create UCI context: %v", err)
		return
	}

	if executor == nil {
		log.Printf("Failed to create executor")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Get all WiFi interfaces to perform scan
	interfaces, err := getWiFiInterfaces(ctx, executor, uci)
	if err != nil {
		log.Printf("Failed to get WiFi interfaces: %v", err)
		return
	}

	var allScanResults []NeighboringScanResult

	// Perform scan on each interface
	for _, iface := range interfaces {
		if iface.Name != "" {
			scanResults, err := performWiFiScan(ctx, executor, iface.Name)
			if err != nil {
				log.Printf("Failed to scan on interface %s: %v", iface.Name, err)
				continue
			}
			allScanResults = append(allScanResults, scanResults...)
		}
	}

	// Save scan result count - Device.WiFi.NeighboringWiFiDiagnostic.ResultNumberOfEntries
	err = uci.Set("WiFi", "NeighboringWiFiDiagnostic.DiagnosticsState", "Complete", false)
	if err != nil {
		log.Printf("Failed to set diagnostics state: %v", err)
	}

	err = uci.Set("WiFi", "NeighboringWiFiDiagnostic.ResultNumberOfEntries", fmt.Sprintf("%d", len(allScanResults)), false)
	if err != nil {
		log.Printf("Failed to set scan result count: %v", err)
	}

	// Process each scan result - Device.WiFi.NeighboringWiFiDiagnostic.Result.{i}.
	for i, result := range allScanResults {
		index := i + 1
		sectionName := fmt.Sprintf("NeighboringWiFiDiagnostic.Result.%d.", index)

		// Basic scan result properties following TR-069 naming
		uci.Set("WIFI", fmt.Sprintf("%sRadio", sectionName), fmt.Sprintf("Device.WiFi.Radio.%d.", 1), false) // Default to radio 1
		uci.Set("WIFI", fmt.Sprintf("%sSSID", sectionName), result.SSID, false)
		uci.Set("WIFI", fmt.Sprintf("%sBSSID", sectionName), result.BSSID, false)
		uci.Set("WIFI", fmt.Sprintf("%sChannel", sectionName), fmt.Sprintf("%d", result.Channel), false)
		uci.Set("WIFI", fmt.Sprintf("%sSignalStrength", sectionName), fmt.Sprintf("%d", result.SignalStrength), false)
		uci.Set("WIFI", fmt.Sprintf("%sOperatingFrequencyBand", sectionName), result.OperatingFrequencyBand, false)
		uci.Set("WIFI", fmt.Sprintf("%sOperatingStandards", sectionName), result.OperatingStandards, false)
		uci.Set("WIFI", fmt.Sprintf("%sOperatingChannelBandwidth", sectionName), result.OperatingChannelBandwidth, false)
		uci.Set("WIFI", fmt.Sprintf("%sNoise", sectionName), fmt.Sprintf("%d", result.Noise), false)
	}

	// Commit changes
	err = uci.Save()
	if err != nil {
		log.Printf("Failed to commit UCI changes: %v", err)
	}
}

// WiFiInterface represents a WiFi interface with its properties
type WiFiInterface struct {
	Name       string
	Device     string
	SSID       string
	BSSID      string
	MACAddress string
	Enable     bool
	Status     string
	Network    string
}

// WiFiRadio represents a WiFi radio device
type WiFiRadio struct {
	Name                    string
	Enable                  bool
	Status                  string
	SupportedFrequencyBands string
	OperatingFrequencyBand  string
	SupportedStandards      string
	OperatingStandards      string
	PossibleChannels        string
	Channel                 int
	AutoChannelSupported    bool
	AutoChannelEnable       bool
	Noise                   int
}

// WiFiStats represents WiFi interface statistics
type WiFiStats struct {
	BytesSent              uint64
	BytesReceived          uint64
	PacketsSent            uint64
	PacketsReceived        uint64
	ErrorsSent             uint32
	ErrorsReceived         uint32
	DiscardPacketsSent     uint32
	DiscardPacketsReceived uint32
}

// AssociatedDevice represents a device associated to an access point
type AssociatedDevice struct {
	MACAddress          string
	AuthenticationState bool
	SignalStrength      int32
	TxRate              string
	RxRate              string
	LastActivity        uint32
	Stats               WiFiStats
}

// NeighboringScanResult represents a neighboring WiFi scan result
type NeighboringScanResult struct {
	SSID                      string
	BSSID                     string
	Channel                   int
	SignalStrength            int32
	OperatingFrequencyBand    string
	OperatingStandards        string
	OperatingChannelBandwidth string
	Noise                     int32
}

// getWiFiInterfaces retrieves all WiFi interfaces from UCI configuration
func getWiFiInterfaces(ctx context.Context, executor *exec.Executor, uci *uci.UCIConfig) ([]WiFiInterface, error) {
	var interfaces []WiFiInterface
	_ = uci // May be used in future for additional UCI queries

	// First, get available wireless interfaces from the system
	availableInterfaces, err := getAvailableWirelessInterfaces(ctx, executor)
	if err != nil {
		log.Printf("Warning: could not get available wireless interfaces: %v", err)
		availableInterfaces = make(map[string]string) // empty map as fallback
	}

	// Get wireless interfaces using UCI show command
	result, err := executor.Execute(ctx, "uci", "show", "wireless")
	if err != nil {
		return nil, fmt.Errorf("failed to get wireless config: %w", err)
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected output type from uci show")
	}

	// Parse UCI wireless configuration - support both named interfaces and @wifi-iface format
	ifaceRegex := regexp.MustCompile(`wireless\.(@wifi-iface\[(\d+)\]|(\w+))\.(\w+)=(.*)`)
	sectionTypeRegex := regexp.MustCompile(`wireless\.(\w+)=wifi-iface`)
	ifaceMap := make(map[string]map[string]string)
	wifiIfaceSections := make(map[string]bool)

	lines := strings.Split(output, "\n")

	// First pass: identify wifi-iface sections
	for _, line := range lines {
		if matches := sectionTypeRegex.FindStringSubmatch(line); matches != nil {
			sectionName := matches[1]
			wifiIfaceSections[sectionName] = true
		}
	}

	// Second pass: parse configuration for wifi-iface sections
	for _, line := range lines {
		if matches := ifaceRegex.FindStringSubmatch(line); matches != nil {
			var index string
			var key string
			var value string

			if matches[2] != "" {
				// @wifi-iface[n] format
				index = matches[2]
				key = matches[4]
				value = strings.Trim(matches[5], "'\"")
			} else if matches[3] != "" {
				// Named interface format (e.g., wifinet0)
				interfaceName := matches[3]
				key = matches[4]
				value = strings.Trim(matches[5], "'\"")

				// Only process if this is a wifi-iface section
				if wifiIfaceSections[interfaceName] {
					index = interfaceName
				} else {
					continue // Skip non-wifi-iface entries
				}
			}

			if index != "" && key != "" {
				if ifaceMap[index] == nil {
					ifaceMap[index] = make(map[string]string)
				}
				ifaceMap[index][key] = value
			}
		}
	}

	// Convert to WiFiInterface structs
	for sectionName, config := range ifaceMap {
		if config["ssid"] != "" {
			// Determine interface name - use ifname if set, otherwise try to find matching interface
			interfaceName := config["ifname"]
			if interfaceName == "" {
				// Try to find the actual interface name using various strategies
				interfaceName = findActualInterfaceName(config, sectionName, availableInterfaces)
			}

			iface := WiFiInterface{
				Name:    interfaceName,
				Device:  config["device"],
				SSID:    config["ssid"],
				Enable:  config["disabled"] != "1",
				Network: config["network"],
			}

			// Get interface status and additional info
			if iface.Name != "" {
				status, bssid, mac := getInterfaceInfo(ctx, executor, iface.Name)
				iface.Status = status
				iface.BSSID = bssid
				iface.MACAddress = mac
			}

			interfaces = append(interfaces, iface)
		}
	}

	return interfaces, nil
}

// getAvailableWirelessInterfaces gets list of available wireless interfaces from the system
func getAvailableWirelessInterfaces(ctx context.Context, executor *exec.Executor) (map[string]string, error) {
	interfaces := make(map[string]string)

	// Get wireless interfaces using iw
	result, err := executor.Execute(ctx, "iw", "dev")
	if err != nil {
		return interfaces, fmt.Errorf("failed to get wireless interfaces: %w", err)
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return interfaces, fmt.Errorf("unexpected output type from iw dev")
	}

	// Parse iw dev output to get interface names and their phy devices
	lines := strings.Split(output, "\n")
	var currentPhy string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "phy#") {
			// Extract phy name (e.g., "phy#0" -> "phy0")
			currentPhy = strings.Replace(line, "phy#", "phy", 1)
		} else if strings.HasPrefix(line, "Interface ") && currentPhy != "" {
			// Extract interface name
			interfaceName := strings.TrimPrefix(line, "Interface ")
			interfaces[interfaceName] = currentPhy
		}
	}

	return interfaces, nil
}

// findActualInterfaceName tries to determine the actual interface name using various strategies
func findActualInterfaceName(config map[string]string, sectionName string, availableInterfaces map[string]string) string {
	// Strategy 1: Look for interfaces that match the device
	deviceName := config["device"]
	if deviceName != "" {
		// Convert device name to expected phy name (e.g., "radio0" -> "phy0")
		expectedPhy := strings.Replace(deviceName, "radio", "phy", 1)

		// Find interface that belongs to this phy
		for ifaceName, phyName := range availableInterfaces {
			if phyName == expectedPhy {
				return ifaceName
			}
		}
	}

	// Strategy 2: Check if section name itself is a valid interface
	if _, exists := availableInterfaces[sectionName]; exists {
		return sectionName
	}

	// Strategy 3: Try common interface naming patterns
	if deviceName != "" {
		// Try wlan0, wlan1, etc. based on device number
		if strings.HasPrefix(deviceName, "radio") {
			radioNum := strings.TrimPrefix(deviceName, "radio")
			candidateName := "wlan" + radioNum
			if _, exists := availableInterfaces[candidateName]; exists {
				return candidateName
			}
		}
	}

	// Strategy 4: If all else fails, return first available interface (risky but better than empty)
	for ifaceName := range availableInterfaces {
		return ifaceName
	}

	return ""
}

// getInterfaceInfo gets interface status, BSSID and MAC address
func getInterfaceInfo(ctx context.Context, executor *exec.Executor, ifaceName string) (status, bssid, mac string) {
	// Get interface info using iw
	result, err := executor.Execute(ctx, "iw", "dev", ifaceName, "info")
	if err != nil {
		return "Down", "", ""
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return "Down", "", ""
	}

	status = "Down"
	if strings.Contains(output, "type AP") || strings.Contains(output, "type managed") {
		status = "Up"
	}

	// Extract BSSID
	bssidRegex := regexp.MustCompile(`addr ([0-9a-fA-F:]{17})`)
	if matches := bssidRegex.FindStringSubmatch(output); matches != nil {
		bssid = matches[1]
		mac = matches[1] // For AP mode, addr is both BSSID and MAC
	}

	return status, bssid, mac
}

// getInterfaceStats retrieves interface statistics
func getInterfaceStats(ctx context.Context, executor *exec.Executor, ifaceName string) (WiFiStats, error) {
	var stats WiFiStats

	// Get stats from /proc/net/dev
	result, err := executor.Execute(ctx, "cat", "/proc/net/dev")
	if err != nil {
		return stats, fmt.Errorf("failed to read /proc/net/dev: %w", err)
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return stats, fmt.Errorf("unexpected output type from /proc/net/dev")
	}

	// Parse interface statistics
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ifaceName+":") {
			fields := strings.Fields(line)
			if len(fields) >= 17 {
				// RX stats: bytes, packets, errs, drop, fifo, frame, compressed, multicast
				stats.BytesReceived, _ = strconv.ParseUint(fields[1], 10, 64)
				stats.PacketsReceived, _ = strconv.ParseUint(fields[2], 10, 64)
				if val, err := strconv.ParseUint(fields[3], 10, 32); err == nil {
					stats.ErrorsReceived = uint32(val)
				}
				if val, err := strconv.ParseUint(fields[4], 10, 32); err == nil {
					stats.DiscardPacketsReceived = uint32(val)
				}

				// TX stats: bytes, packets, errs, drop, fifo, colls, carrier, compressed
				stats.BytesSent, _ = strconv.ParseUint(fields[9], 10, 64)
				stats.PacketsSent, _ = strconv.ParseUint(fields[10], 10, 64)
				if val, err := strconv.ParseUint(fields[11], 10, 32); err == nil {
					stats.ErrorsSent = uint32(val)
				}
				if val, err := strconv.ParseUint(fields[12], 10, 32); err == nil {
					stats.DiscardPacketsSent = uint32(val)
				}
			}
			break
		}
	}

	return stats, nil
}

// getWiFiRadios retrieves all WiFi radio devices
func getWiFiRadios(ctx context.Context, executor *exec.Executor, uci *uci.UCIConfig) ([]WiFiRadio, error) {
	var radios []WiFiRadio
	_ = uci // May be used in future for additional UCI queries

	// Get wireless devices using UCI show command
	result, err := executor.Execute(ctx, "uci", "show", "wireless")
	if err != nil {
		return nil, fmt.Errorf("failed to get wireless config: %w", err)
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected output type from uci show")
	}

	// Parse UCI wireless configuration for wifi-device entries
	deviceRegex := regexp.MustCompile(`wireless\.@wifi-device\[(\d+)\]\.(\w+)=(.*)`)
	deviceMap := make(map[string]map[string]string)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if matches := deviceRegex.FindStringSubmatch(line); matches != nil {
			index := matches[1]
			key := matches[2]
			value := strings.Trim(matches[3], "'\"")

			if deviceMap[index] == nil {
				deviceMap[index] = make(map[string]string)
			}
			deviceMap[index][key] = value
		}
	}

	// Convert to WiFiRadio structs
	for _, config := range deviceMap {
		radio := WiFiRadio{
			Name:                 config["type"],
			Enable:               config["disabled"] != "1",
			AutoChannelSupported: true, // Most modern radios support auto-channel
		}

		// Parse channel
		if channel, err := strconv.Atoi(config["channel"]); err == nil {
			radio.Channel = channel
			radio.AutoChannelEnable = false
		} else if config["channel"] == "auto" {
			radio.AutoChannelEnable = true
		}

		// Get additional radio info
		phyName := "phy" + config["path"] // Approximate PHY name
		if phyName == "phy" {
			phyName = "phy0" // Default fallback
		}

		bands, standards, channels := getRadioCapabilities(ctx, executor, phyName)
		radio.SupportedFrequencyBands = bands
		radio.OperatingFrequencyBand = bands // Assume operating same as supported for now
		radio.SupportedStandards = standards
		radio.OperatingStandards = config["hwmode"]
		radio.PossibleChannels = channels

		// Get radio status
		radio.Status = getRadioStatus(ctx, executor, phyName)

		radios = append(radios, radio)
	}

	return radios, nil
}

// getRadioCapabilities gets supported frequency bands, standards and channels
func getRadioCapabilities(ctx context.Context, executor *exec.Executor, phyName string) (bands, standards, channels string) {
	// Get PHY info using iw
	result, err := executor.Execute(ctx, "iw", "phy", phyName, "info")
	if err != nil {
		return "2.4GHz", "b,g,n", "1,2,3,4,5,6,7,8,9,10,11"
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return "2.4GHz", "b,g,n", "1,2,3,4,5,6,7,8,9,10,11"
	}

	// Parse supported bands and channels
	has24GHz := strings.Contains(output, "2412 MHz") || strings.Contains(output, "2.4 GHz")
	has5GHz := strings.Contains(output, "5180 MHz") || strings.Contains(output, "5 GHz")

	if has24GHz && has5GHz {
		bands = "2.4GHz,5GHz"
		standards = "a,b,g,n,ac"
		channels = "1,2,3,4,5,6,7,8,9,10,11,36,40,44,48,149,153,157,161,165"
	} else if has5GHz {
		bands = "5GHz"
		standards = "a,n,ac"
		channels = "36,40,44,48,149,153,157,161,165"
	} else {
		bands = "2.4GHz"
		standards = "b,g,n"
		channels = "1,2,3,4,5,6,7,8,9,10,11"
	}

	return bands, standards, channels
}

// getRadioStatus gets the status of a radio
func getRadioStatus(ctx context.Context, executor *exec.Executor, phyName string) string {
	// Check if any interface on this PHY is up
	result, err := executor.Execute(ctx, "iw", "dev")
	if err != nil {
		return "Down"
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return "Down"
	}

	// Look for interfaces on this PHY
	lines := strings.Split(output, "\n")
	inPhy := false
	for _, line := range lines {
		if strings.Contains(line, phyName) {
			inPhy = true
			continue
		}
		if inPhy && strings.Contains(line, "Interface") {
			return "Up"
		}
		if inPhy && strings.HasPrefix(line, "phy") {
			break
		}
	}

	return "Down"
}

// getAssociatedDevices gets devices associated with an access point interface
func getAssociatedDevices(ctx context.Context, executor *exec.Executor, ifaceName string) ([]AssociatedDevice, error) {
	var devices []AssociatedDevice

	// Use wlanconfig to get associated stations
	result, err := executor.Execute(ctx, "wlanconfig", ifaceName, "list")
	if err != nil {
		return devices, nil // No error if no stations or interface not found
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return devices, nil
	}

	// Parse wlanconfig output
	lines := strings.Split(output, "\n")
	var currentDevice *AssociatedDevice

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// New client entry starts with MAC address (format: xx:xx:xx:xx:xx:xx)
		macRegex := regexp.MustCompile(`^([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)
		if macMatch := macRegex.FindStringSubmatch(line); macMatch != nil {
			// Save previous device if exists
			if currentDevice != nil {
				devices = append(devices, *currentDevice)
			}

			// Initialize new device
			currentDevice = &AssociatedDevice{
				MACAddress:          macMatch[1],
				AuthenticationState: true, // If listed, assume authenticated
				SignalStrength:      0,
				TxRate:              "0M",
				RxRate:              "0M",
				LastActivity:        0,
			}

			// Parse the basic info from the main line
			parts := strings.Fields(line)
			if len(parts) >= 7 {
				// Format: MAC AID CHAN TXRATE RXRATE RSSI MINRSSI MAXRSSI [IDLE] ...
				// Extract RSSI (signal strength)
				if rssi, err := strconv.ParseInt(parts[5], 10, 32); err == nil {
					currentDevice.SignalStrength = int32(rssi)
				}
				// Extract TX rate
				if len(parts) > 3 {
					currentDevice.TxRate = parts[3]
				}
				// Extract RX rate
				if len(parts) > 4 {
					currentDevice.RxRate = parts[4]
				}
				// Extract idle time if available
				if len(parts) > 8 {
					if idle, err := strconv.ParseUint(parts[8], 10, 32); err == nil {
						currentDevice.LastActivity = uint32(idle)
					}
				}
			}
		} else if currentDevice != nil {
			// Parse additional device info from subsequent lines
			if strings.Contains(line, "SNR") {
				if snr := regexp.MustCompile(`SNR\s*:\s*(\d+)`).FindStringSubmatch(line); snr != nil {
					// SNR can be used to improve signal strength calculation if needed
					if snrVal, err := strconv.ParseInt(snr[1], 10, 32); err == nil {
						// SNR is available but we're using RSSI as primary signal strength
						_ = snrVal // Keep for potential future use
					}
				}
			}

			if strings.Contains(line, "Minimum Tx Power") {
				// Additional power info if needed for extended stats
			}

			if strings.Contains(line, "Maximum Tx Power") {
				// Additional power info if needed for extended stats
			}
		}
	}

	// Add the last device if exists
	if currentDevice != nil {
		devices = append(devices, *currentDevice)
	}

	return devices, nil
}

// getAPStatus converts interface status to AP status
func getAPStatus(status string) string {
	if status == "Up" {
		return "Enabled"
	}
	return "Disabled"
}

// getSecurityConfig gets encryption and key passphrase using executor
func getSecurityConfig(ctx context.Context, executor *exec.Executor, ifaceName string) (encryption, keyPassphrase string) {
	// Get encryption type using executor
	_encType := getWirelessInterfaceEncryption(ctx, executor, ifaceName)
	if _encType == "" {
		return "None", ""
	}
	// Convert UCI encryption to TR-069 format
	switch {
	case strings.HasPrefix(_encType, "psk2"):
		encryption = "WPA2-Personal"
	case strings.HasPrefix(_encType, "psk-mixed"):
		encryption = "WPA-WPA2-Personal"
	case strings.HasPrefix(_encType, "psk"):
		encryption = "WPA-Personal"
	case strings.HasPrefix(_encType, "wpa2"):
		encryption = "WPA2-Enterprise"
	case strings.HasPrefix(_encType, "wpa-mixed"):
		encryption = "WPA-WPA2-Enterprise"
	case strings.HasPrefix(_encType, "wpa"):
		encryption = "WPA-Enterprise"
	case strings.HasPrefix(_encType, "wep"):
		encryption = "WEP-128" // Assume WEP-128 as default
	default:
		encryption = "None"
	}

	// Get key passphrase if available
	if encryption != "None" {
		_key := getWirelessInterfaceKey(ctx, executor, ifaceName)
		if _key != "" {
			keyPassphrase = _key
		}
	}

	return encryption, keyPassphrase
}

// performWiFiScan performs WiFi scan on an interface and returns neighboring APs
func performWiFiScan(ctx context.Context, executor *exec.Executor, ifaceName string) ([]NeighboringScanResult, error) {
	var results []NeighboringScanResult

	// Use iw scan to get neighboring WiFi networks
	result, err := executor.Execute(ctx, "iw", "dev", ifaceName, "scan")
	if err != nil {
		return results, nil // No error if scan fails, just return empty results
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return results, nil
	}

	// Parse scan results
	lines := strings.Split(output, "\n")
	var currentResult *NeighboringScanResult

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// New BSS entry
		if strings.HasPrefix(line, "BSS ") {
			if currentResult != nil {
				results = append(results, *currentResult)
			}

			// Extract BSSID
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bssid := strings.TrimSuffix(parts[1], "(on")
				currentResult = &NeighboringScanResult{
					BSSID: bssid,
					Noise: -95, // Default noise floor
				}
			}
		} else if currentResult != nil {
			// Parse BSS parameters
			if strings.Contains(line, "freq:") {
				freqStr := strings.TrimSpace(strings.Split(line, ":")[1])
				if freq, err := strconv.Atoi(freqStr); err == nil {
					currentResult.Channel = freqToChannel(freq)
					currentResult.OperatingFrequencyBand = freqToBand(freq)
				}
			} else if strings.Contains(line, "signal:") {
				signalStr := strings.TrimSpace(strings.Split(line, ":")[1])
				signalStr = strings.Fields(signalStr)[0] // Remove "dBm"
				if signal, err := strconv.ParseInt(signalStr, 10, 32); err == nil {
					currentResult.SignalStrength = int32(signal)
				}
			} else if strings.Contains(line, "SSID:") {
				ssid := strings.TrimSpace(strings.Split(line, ":")[1])
				currentResult.SSID = ssid
			}
		}
	}

	// Add the last result if exists
	if currentResult != nil {
		results = append(results, *currentResult)
	}

	return results, nil
}

// freqToChannel converts frequency to channel number
func freqToChannel(freq int) int {
	if freq >= 2412 && freq <= 2484 {
		// 2.4 GHz band
		if freq == 2484 {
			return 14
		}
		return (freq-2412)/5 + 1
	} else if freq >= 5170 && freq <= 5825 {
		// 5 GHz band
		return (freq - 5000) / 5
	}
	return 0
}

// freqToBand converts frequency to band string
func freqToBand(freq int) string {
	if freq >= 2412 && freq <= 2484 {
		return "2.4GHz"
	} else if freq >= 5170 && freq <= 5825 {
		return "5GHz"
	}
	return "Unknown"
}

// getWirelessInterfaceMode gets the mode of a wireless interface using executor
func getWirelessInterfaceMode(ctx context.Context, executor *exec.Executor, ifaceName string) string {
	result, err := executor.Execute(ctx, "uci", "get", fmt.Sprintf("wireless.%s.mode", ifaceName))
	if err != nil {
		return ""
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(output)
}

// getWirelessInterfaceEncryption gets the encryption type of a wireless interface using executor
func getWirelessInterfaceEncryption(ctx context.Context, executor *exec.Executor, ifaceName string) string {
	result, err := executor.Execute(ctx, "uci", "get", fmt.Sprintf("wireless.%s.encryption", ifaceName))
	if err != nil {
		return ""
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(output)
}

// getWirelessInterfaceKey gets the key/passphrase of a wireless interface using executor
func getWirelessInterfaceKey(ctx context.Context, executor *exec.Executor, ifaceName string) string {
	result, err := executor.Execute(ctx, "uci", "get", fmt.Sprintf("wireless.%s.key", ifaceName))
	if err != nil {
		return ""
	}

	output, ok := result.Stdout.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(output)
}
