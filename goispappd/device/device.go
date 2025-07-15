package device

// Device represents the overall TR-069 data model structure for a CPE,
// as understood and potentially cached by the ACS. It aggregates various
// sub-models like DeviceInfo, ManagementServer, interfaces, etc.

type Device struct {
	*InternetGatewayDevice
	RootDataModelVersion          float64                  // Root data model version, e.g. '2.15'
	InterfaceStackNumberOfEntries int                      // Number of entries in the InterfaceStack table
	DeviceSummary                 string                   // Device summary for backward compatibility
	DeviceInfo                    DeviceInfo               // General device identification and status information.
	ManagementServer              ManagementServer         // Configuration related to the ACS connection.
	InterfaceStack                []InterfaceStackEntry    // Describes the layering of network interfaces.
	Cellular                      CellularDevice           // Cellular interface and related configurations.
	Ethernet                      EthernetDevice           // Ethernet interface configurations.
	WiFi                          WiFiDevice               // WiFi interface and access point configurations.
	PPP                           PPPDevice                // PPP interface configurations.
	IP                            IPDevice                 // IP layer configurations and diagnostics.
	Routing                       RoutingDevice            // Routing configurations.
	Hosts                         HostsDevice              // Information about hosts connected to the CPE.
	DNS                           DNSDevice                // DNS client configuration.
	DHCPv4                        DHCPv4Device             // DHCPv4 client and server configuration.
	Firewall                      FirewallDevice           // Firewall configuration (standard and Mikrotik extensions).
	WAN                           WANDevice                // WAN device configuration and status.
	X_ISPAPP_Interface            XMikrotikInterfaceDevice // Mikrotik generic interfaces.
	X_ISPAPP_Monitor              XMikrotikMonitorDevice   // Mikrotik traffic monitoring.
	NeedMethods                   bool
	NeedParams                    bool
}

// huawei
type InternetGatewayDevice = Device

// DeviceInfo contains general static and dynamic information about the CPE.
type DeviceInfo struct {
	OutsideIPAddress                string             // Public IP address of the CPE.
	Manufacturer                    string             // Name of the device manufacturer.
	ManufacturerOUI                 string             // OUI of the manufacturer.
	ModelName                       string             // Model name of the device.
	Description                     string             // Textual description of the device.
	ProductClass                    string             // Product class identifier.
	SerialNumber                    string             // Unique serial number of the device.
	SpecVersion                     string             // Specification version identifier.
	HardwareVersion                 string             // Hardware version identifier.
	SoftwareVersion                 string             // Software/firmware version identifier.
	ProvisioningCode                string             // Optional code for provisioning.
	UpTime                          int                // Time in seconds since the device last booted.
	VendorConfigFileNumberOfEntries int                // Number of entries in VendorConfigFile table
	X_ISPAPP_SystemIdentity         string             // Mikrotik system identity name.
	X_ISPAPP_ArchName               string             // Mikrotik architecture name.
	X_ISPAPP_BrandingPckgBuildTime  string             // Branding package build time
	X_ISPAPP_AutosupoutTime         string             // Generated autosupout.rif timestamp
	VendorConfigFile                []VendorConfigFile // Vendor config file entries
	MemoryStatus                    MemoryStatus       // Status of device's physical memory
	ProcessStatus                   ProcessStatus      // Status of the processes on the device
	// Geolocation fields for device location information
	Latitude               float64 // Device geographical latitude
	Longitude              float64 // Device geographical longitude
	Country                string  // Device country location
	City                   string  // Device city location
	Region                 string  // Device region/state location
	Timezone               string  // Device timezone
	GeolocationSource      string  // Source of geolocation data (cloudflare, api, manual)
	GeolocationLastUpdated string  // Last time geolocation was updated
}

// VendorConfigFile represents an entry in the VendorConfigFile table
type VendorConfigFile struct {
	Index               int    // TR-069 index for this vendor config file
	Name                string // Name of configuration file
	Description         string // Description of the vendor configuration file
	UseForBackupRestore bool   // Whether this file should be used for backup/restore
}

// MemoryStatus contains information about the device's physical memory
type MemoryStatus struct {
	Total int // Total physical volatile RAM in KiB
	Free  int // Free physical volatile RAM in KiB
}

// ProcessStatus contains information about the device's processes
type ProcessStatus struct {
	CPUUsage int // Total CPU usage as a percentage (0-100)
}

// ManagementServer contains parameters related to the CPE's connection
// and interaction with the ACS.
type ManagementServer struct {
	URL                            string            // URL of the ACS the CPE connects to.
	Username                       string            // Username for CPE authentication to the ACS.
	Password                       string            // Password for CPE authentication to the ACS (should be handled securely).
	PeriodicInformEnable           bool              // Whether periodic informs are enabled.
	PeriodicInformInterval         int               // Interval in seconds for periodic informs.
	ParameterKey                   string            // Key provided by ACS for tracking configuration changes. Read-only for ACS.
	ConnectionRequestURL           string            // URL on the CPE for ACS connection requests. Read-only for ACS.
	ConnectionRequestUsername      string            // Username for ACS authentication to the CPE.
	ConnectionRequestPassword      string            // Password for ACS authentication to the CPE (should be handled securely).
	AliasBasedAddressing           bool              // Whether the CPE supports alias-based addressing. Read-only for ACS.
	InformParameterNumberOfEntries int               // Number of entries in the InformParameter table
	InformParameter                []InformParameter // Inform parameter entries
}

// InformParameter represents an entry in the InformParameter table
type InformParameter struct {
	Index         int      // TR-069 index for this inform parameter
	Enable        bool     // Enables or disables this InformParameter
	ParameterName string   // Parameter pattern to be included in the Inform
	EventList     []string // Events for which this parameter should be included
}

// InterfaceStackEntry describes a single link in the interface stack,
// showing the relationship between a higher-layer and a lower-layer interface.
type InterfaceStackEntry struct {
	Index       int    // TR-069 index for this interface stack entry
	HigherLayer string // Path name of the higher-level interface object.
	LowerLayer  string // Path name of the lower-level interface object it runs on top of.
}

// CellularDevice aggregates all cellular-related configurations and status.
type CellularDevice struct {
	InterfaceNumberOfEntries   int                   // Number of entries in the Interface table
	AccessPointNumberOfEntries int                   // Number of entries in the AccessPoint table
	X_ISPAPP_Antenna           string                // Which antenna to use for modem
	X_ISPAPP_CurrentAntenna    string                // Currently selected antenna
	Interfaces                 []CellularInterface   // List of cellular modem interfaces.
	AccessPoints               []CellularAccessPoint // List of configured APNs.
	X_ISPAPP_CellDiagnostics   CellDiagnostics       // Cell diagnostics information
}

// CellDiagnostics represents Mikrotik specific cell diagnostics
type CellDiagnostics struct {
	DiagnosticsState      string           // State of cell diagnostics
	Interface             string           // Interface for diagnostics
	Seconds               int              // Seconds to perform scan
	ResultNumberOfEntries int              // Number of results
	Results               []CellDiagResult // Diagnostic results
}

// CellDiagResult represents a result from cell diagnostics
type CellDiagResult struct {
	Index          int // TR-069 index for this cell diagnostic result
	Band           int // Band number
	Fcn            int // Frequency Channel Number in MHz
	PhysicalCellId int // Physical Cell ID
	RSSI           int // Signal strength in dBm
	RSRP           int // Reference Signal Received Power in dBm
	RSRQ           int // Reference Signal Received Quality in dB
}

// CellularInterface represents a single cellular modem interface.
type CellularInterface struct {
	Index                                 int             // TR-069 index for this interface
	Enable                                bool            // Administrative status (enabled/disabled).
	Status                                string          // Operational status (Up, Down, Dormant, etc.).
	LowerLayers                           string          // Reference to lower layers
	IMEI                                  string          // International Mobile Equipment Identity.
	RSSI                                  int             // Received Signal Strength Indicator (dBm)
	X_ISPAPP_Model                        string          // Modem model
	X_ISPAPP_Revision                     string          // Modem revision
	X_ISPAPP_ExtRevision                  string          // Extended revision info
	X_ISPAPP_SupportedAccessTechnologies  []string        // Supported access technologies
	X_ISPAPP_AccessTechnologies           []string        // Enabled access technologies
	X_ISPAPP_CurrentAccessTechnology      string          // Currently used access technology
	X_ISPAPP_SupportedLteBands            []string        // Supported LTE bands
	X_ISPAPP_LteBands                     []string        // Configured LTE bands
	X_ISPAPP_LteCellLock                  []string        // Cell lock configuration
	X_ISPAPP_Supported5GBands             []string        // Supported 5G bands
	X_ISPAPP_5GBands                      []string        // Configured 5G bands
	X_ISPAPP_RSCP                         int             // Received Signal Code Power
	X_ISPAPP_ECNO                         int             // ECNO value
	X_ISPAPP_SINR                         int             // SINR value in dB
	X_ISPAPP_RSRP                         int             // RSRP value in dBm
	X_ISPAPP_MimoRSRP                     []int           // RSRP MIMO values
	X_ISPAPP_RSRQ                         int             // RSRQ value in dB
	X_ISPAPP_CQI                          int             // CQI value
	X_ISPAPP_RI                           int             // Rank Indicator
	X_ISPAPP_MCS                          int             // Modulation Coding Scheme
	X_ISPAPP_TBS                          int             // Transport Block Size
	X_ISPAPP_RBs                          int             // Number of allocated Resource Blocks
	X_ISPAPP_Modulation                   string          // Modulation type
	X_ISPAPP_5G_CQI                       int             // CQI value for 5G
	X_ISPAPP_5G_RI                        int             // Rank Indicator for 5G
	X_ISPAPP_5G_MCS                       int             // Modulation Coding Scheme for 5G
	X_ISPAPP_5G_TBS                       int             // Transport Block Size for 5G
	X_ISPAPP_5G_RBs                       int             // Number of Resource Blocks for 5G
	X_ISPAPP_5G_Modulation                string          // Modulation type for 5G
	X_ISPAPP_5G_DataPath                  string          // User Layer Data Path under NSA Network
	X_ISPAPP_TxPUCCH                      int             // TxPUCCH value
	X_ISPAPP_TxPUSCH                      int             // TxPUSCH value
	X_ISPAPP_TxSRS                        int             // TxSRS value
	X_ISPAPP_TxPRACH                      int             // TxPRACH value
	X_ISPAPP_5G_TxPUCCH                   int             // TxPUCCH value for 5G
	X_ISPAPP_5G_TxPUSCH                   int             // TxPUSCH value for 5G
	X_ISPAPP_5G_TxSRS                     int             // TxSRS value for 5G
	X_ISPAPP_5G_TxPRACH                   int             // TxPRACH value for 5G
	X_ISPAPP_5G_Band                      int             // Band for 5G
	X_ISPAPP_5G_Bandwidth                 int             // Bandwidth in MHz for 5G
	X_ISPAPP_5G_PhysicalCellId            int             // Physical Cell ID for 5G
	X_ISPAPP_5G_SINR                      int             // SINR value for 5G in dB
	X_ISPAPP_5G_RSRP                      int             // RSRP value for 5G in dBm
	X_ISPAPP_5G_RSRQ                      int             // RSRQ value for 5G in dB
	X_ISPAPP_CellId                       int             // Cell ID value
	X_ISPAPP_BandInfo                     string          // Human readable band info
	X_ISPAPP_LinkDowns                    int             // Number of link downs
	X_ISPAPP_AccessPoints                 []string        // List of AccessPoint profiles
	X_ISPAPP_CarrierInfoNumberOfEntries   int             // Number of entries in CarrierInfo table
	X_ISPAPP_CarrierInfo5GNumberOfEntries int             // Number of entries in CarrierInfo5G table
	USIM                                  SIMCard         // Details of the inserted SIM card.
	Stats                                 InterfaceStats  // Interface statistics
	X_ISPAPP_CarrierInfo                  []CarrierInfo   // Carrier information entries
	X_ISPAPP_CarrierInfo5G                []CarrierInfo5G // 5G carrier information entries
}

// CarrierInfo represents carrier information for cellular interfaces
type CarrierInfo struct {
	Index          int  // TR-069 index for this carrier info
	Band           int  // Band number
	Fcn            int  // Frequency Channel Number in MHz
	Bandwidth      int  // Bandwidth in MHz
	PhysicalCellId int  // Physical Cell ID
	RSSI           int  // Signal strength in dBm
	SINR           int  // SINR value in dB
	RSRP           int  // RSRP value in dBm
	RSRQ           int  // RSRQ value in dB
	UplinkCA       bool // CA UL status
}

// CarrierInfo5G represents 5G carrier information
type CarrierInfo5G struct {
	Index          int // TR-069 index for this 5G carrier info
	Band           int // Band number
	Bandwidth      int // Bandwidth in MHz
	PhysicalCellId int // Physical Cell ID
	SINR           int // SINR value in dB
	RSRP           int // RSRP value in dBm
	RSRQ           int // RSRQ value in dB
	SNR            int // SNR value in dB
}

// SIMCard holds information about the Universal Subscriber Identity Module (USIM).
type SIMCard struct {
	IMSI  string // International Mobile Subscriber Identity.
	ICCID string // Integrated Circuit Card Identifier.
}

// CellularAccessPoint defines the configuration for a cellular Access Point Name (APN).
type CellularAccessPoint struct {
	Index    int    // TR-069 index for this access point
	APN      string // Access Point Name string.
	Username string // Username for APN authentication.
	Password string // Password for APN authentication (handle securely).
}

// EthernetDevice aggregates Ethernet interface configurations.
type EthernetDevice struct {
	InterfaceNumberOfEntries int                 // Number of entries in Interface table
	LinkNumberOfEntries      int                 // Number of entries in Link table
	Interfaces               []EthernetInterface // List of physical Ethernet interfaces.
	Links                    []EthernetLink      // Layer 2 Ethernet links.
}

// EthernetLink represents a layer 2 Ethernet link/connection
type EthernetLink struct {
	Index       int    // TR-069 index for this link
	Enable      bool   // Administrative status.
	Status      string // Operational status (Up, Down, etc.).
	LowerLayers string // Reference to underlying physical interface.
	// MACAddress not in the reference, but likely needed
}

// EthernetInterface represents a physical Ethernet port and its status.
type EthernetInterface struct {
	Index              int            // TR-069 index for this interface
	Enable             bool           // Administrative status.
	Status             string         // Operational status (Up, Down, etc.).
	LowerLayers        string         // Reference to lower layers (usually empty for physical Ethernet).
	MACAddress         string         // Burned-in MAC address of the interface.
	CurrentBitRate     int            // Current negotiated speed in Mbps.
	X_ISPAPP_LinkDowns int            // Number of link down events since boot.
	X_ISPAPP_Name      string         // Interface name in RouterOS.
	X_ISPAPP_Comment   string         // User comment for the interface.
	Stats              InterfaceStats // Interface statistics.
}

// WiFiDevice aggregates all WiFi related configurations.
type WiFiDevice struct {
	RadioNumberOfEntries       int                       // Number of entries in Radio table
	SSIDNumberOfEntries        int                       // Number of entries in SSID table
	AccessPointNumberOfEntries int                       // Number of entries in AccessPoint table
	Radios                     []WiFiRadio               // List of physical WiFi radios.
	SSIDs                      []WiFiSSID                // List of configured Service Set Identifiers (logical interfaces).
	AccessPoints               []WiFiAccessPoint         // List of Access Point configurations.
	NeighboringWiFiDiagnostic  NeighboringWiFiDiagnostic // Neighboring WiFi diagnostic information
}

// NeighboringWiFiDiagnostic represents WiFi scanning diagnostic information
type NeighboringWiFiDiagnostic struct {
	DiagnosticsState      string                  // State of diagnostics
	ResultNumberOfEntries int                     // Number of results
	Result                []NeighboringWiFiResult // Diagnostic results
}

// NeighboringWiFiResult represents a result from WiFi diagnostics
type NeighboringWiFiResult struct {
	Radio                     string   // Radio that detected the network
	SSID                      string   // Service Set Identifier
	BSSID                     string   // Basic Service Set Identifier (MAC)
	Channel                   int      // Operating channel
	SignalStrength            int      // Signal strength in dBm
	OperatingFrequencyBand    string   // Frequency band (2.4GHz, 5GHz)
	OperatingStandards        []string // Standards in use (a, b, g, n, ac, ax)
	OperatingChannelBandwidth string   // Channel bandwidth (20MHz, 40MHz, 80MHz)
	Noise                     int      // Noise level in dBm
}

// WiFiRadio represents a physical WiFi radio (e.g., 2.4GHz or 5GHz radio).
type WiFiRadio struct {
	Index                    int                     // TR-069 index for this radio
	Enable                   bool                    // Administrative status.
	Status                   string                  // Operational status.
	LowerLayers              string                  // Reference to lower layers (usually empty for physical radio).
	SupportedFrequencyBands  string                  // Bands supported (e.g., "2.4GHz", "5GHz").
	OperatingFrequencyBand   string                  // Currently active frequency band.
	SupportedStandards       string                  // Standards supported (e.g., "g,n,ac,ax").
	OperatingStandards       string                  // Currently active standards.
	PossibleChannels         string                  // Possible radio channels for the standard (comma-separated)
	Channel                  int                     // Current operating channel.
	AutoChannelSupported     bool                    // Whether automatic channel selection is supported
	AutoChannelEnable        bool                    // Whether automatic channel selection is enabled.
	X_ISPAPP_SkipDFSChannels string                  // DFS channel skipping configuration
	Stats                    WiFiRadioStats          // Radio statistics
	X_ISPAPP_Stats           XMikrotikWiFiRadioStats // Mikrotik-specific stats
}

// WiFiRadioStats contains statistics for a WiFi radio
type WiFiRadioStats struct {
	Noise int // Average noise strength in dBm
}

// XMikrotikWiFiRadioStats contains Mikrotik-specific WiFi radio statistics
type XMikrotikWiFiRadioStats struct {
	OverallTxCCQ int // CCQ value in percent
}

// WiFiSSID represents a logical WiFi network (Service Set Identifier).
// It typically runs on top of a WiFiRadio.
type WiFiSSID struct {
	Index       int           // TR-069 index for this SSID
	Enable      bool          // Administrative status.
	Status      string        // Operational status.
	LowerLayers string        // Reference to the underlying WiFi.Radio
	BSSID       string        // Basic Service Set Identifier (MAC address of the AP for this SSID).
	MACAddress  string        // MAC address used by this logical interface.
	SSID        string        // The Service Set Identifier (network name).
	Stats       WiFiSSIDStats // SSID statistics
}

// Add WiFiSSIDStats struct for SSID stats
// WiFiSSIDStats contains statistics for a WiFi SSID
// (TR-181: Device.WiFi.SSID.{i}.Stats.*)
type WiFiSSIDStats struct {
	BytesSent              uint64
	BytesReceived          uint64
	PacketsSent            uint64
	PacketsReceived        uint64
	ErrorsSent             uint32
	ErrorsReceived         uint32
	DiscardPacketsSent     uint32
	DiscardPacketsReceived uint32
}

// WiFiAccessPoint configures an SSID to operate in Access Point mode.
type WiFiAccessPoint struct {
	Index                           int                    // TR-069 index for this access point
	Enable                          bool                   // Whether this AP configuration is active.
	Status                          string                 // Operational status (Enabled, Disabled, Error).
	SSIDReference                   string                 // Path name of the WiFi.SSID object this AP uses.
	SSIDAdvertisementEnabled        bool                   // Whether the SSID is broadcast in beacons.
	AssociatedDeviceNumberOfEntries int                    // Number of entries in AssociatedDevice table
	Security                        WiFiSecurity           // Security settings for this AP.
	AssociatedDevices               []WiFiAssociatedDevice // List of currently connected client devices.
}

// WiFiSecurity holds the security configuration for a WiFi Access Point.
type WiFiSecurity struct {
	ModesSupported string // Security modes supported by the device (comma-separated, e.g., "WPA2-Personal,WPA-Personal").
	ModeEnabled    string // Currently enabled security mode.
	KeyPassphrase  string // The WPA/WPA2 passphrase (handle securely).
}

// WiFiAssociatedDevice represents a client device connected to a WiFi Access Point.
type WiFiAssociatedDevice struct {
	Index               int                            // TR-069 index for this device
	MACAddress          string                         // MAC address of the connected client.
	AuthenticationState bool                           // Whether the client is authenticated.
	SignalStrength      int                            // Signal strength of the client's uplink (dBm).
	Stats               WiFiAssociatedDeviceStats      // Standard statistics
	X_ISPAPP_Stats      XMikrotikAssociatedDeviceStats // Mikrotik-specific statistics
}

// WiFiAssociatedDeviceStats contains statistics for an associated device
type WiFiAssociatedDeviceStats struct {
	BytesSent       uint64 // Total bytes sent to the device
	BytesReceived   uint64 // Total bytes received from the device
	PacketsSent     uint64 // Total packets sent to the device
	PacketsReceived uint64 // Total packets received from the device
}

// XMikrotikAssociatedDeviceStats contains Mikrotik-specific statistics for an associated device
type XMikrotikAssociatedDeviceStats struct {
	TxFrames          uint64 // Transmitted frames
	RxFrames          uint64 // Received frames
	TxFrameBytes      uint64 // Transmitted frame bytes
	RxFrameBytes      uint64 // Received frame bytes
	TxHwFrames        uint64 // Hardware transmitted frames
	RxHwFrames        uint64 // Hardware received frames
	TxHwFrameBytes    uint64 // Hardware transmitted frame bytes
	RxHwFrameBytes    uint64 // Hardware received frame bytes
	TxCCQ             uint64 // Client Connection Quality for transmit (percent)
	RxCCQ             uint64 // Client Connection Quality for receive (percent)
	SignalToNoise     int    // Signal-to-noise ratio (dB)
	RxRate            string // Receive rate
	TxRate            string // Transmit rate
	LastActivity      uint64 // Last activity time (ms)
	SignalStrengthCh0 int    // Signal strength chain 0 (dBm)
	SignalStrengthCh1 int    // Signal strength chain 1 (dBm)
	StrengthAtRates   string // Signal strength at different rates
	UpTime            uint64 // Client uptime (seconds)
}

// InterfaceStats holds common statistics for network interfaces.
type InterfaceStats struct {
	BytesSent              uint64 // Total bytes sent.
	BytesReceived          uint64 // Total bytes received.
	PacketsSent            uint64 // Total packets sent.
	PacketsReceived        uint64 // Total packets received.
	ErrorsSent             uint32 // Outbound errors.
	ErrorsReceived         uint32 // Inbound errors.
	DiscardPacketsSent     uint32 // Outbound discards.
	DiscardPacketsReceived uint32 // Inbound discards.
}

// PPPDevice aggregates PPP interface configurations.
type PPPDevice struct {
	InterfaceNumberOfEntries int            // Number of entries in Interface table
	Interfaces               []PPPInterface // List of PPP interfaces.
}

// PPPInterface represents a PPP connection interface.
type PPPInterface struct {
	Index              int            // TR-069 index for this interface
	Enable             bool           // Administrative status.
	Status             string         // Operational status (Up, Down, etc.).
	LowerLayers        string         // Reference to the lower layer interface (e.g., Ethernet Link, Cellular).
	ConnectionStatus   string         // PPP connection state (Connecting, Connected, Disconnected, etc.).
	AutoDisconnectTime int            // Time in seconds after connection before auto-disconnect (0=disabled).
	IdleDisconnectTime int            // Time in seconds of inactivity before auto-disconnect (0=disabled).
	Username           string         // Username for PPP authentication.
	Password           string         // Password for PPP authentication (handle securely).
	EncryptionProtocol string         // Encryption protocol used (None, MPPE).
	ConnectionTrigger  string         // Trigger for establishing connection (OnDemand, AlwaysOn).
	X_ISPAPP_Type      string         // Mikrotik specific type (e.g., PPPoE).
	PPPoE              PPPoESettings  // PPPoE specific settings.
	IPCP               IPCPSettings   // IPCP settings (for IPv4).
	Stats              InterfaceStats // Interface statistics.
}

// PPPoESettings contains parameters specific to PPPoE connections.
type PPPoESettings struct {
	ACName      string // Access Concentrator name.
	ServiceName string // Service Name identifier.
}

// IPCPSettings contains parameters related to IP Control Protocol (IPv4).
type IPCPSettings struct {
	LocalIPAddress  string // Local IPv4 address assigned via IPCP.
	RemoteIPAddress string // Remote IPv4 address assigned via IPCP.
}

// IPDevice aggregates IP layer configurations and diagnostics.
type IPDevice struct {
	InterfaceNumberOfEntries int           // Number of entries in Interface table
	Interfaces               []IPInterface // List of IP interfaces.
	Diagnostics              IPDiagnostics // IP layer diagnostic tools.
}

// IPInterface represents a layer 3 IP interface.
type IPInterface struct {
	Index                      int                // TR-069 index for this interface
	Enable                     bool               // Administrative status.
	Status                     string             // Operational status (Up, Down, etc.).
	LowerLayers                string             // Reference to the lower layer interface (e.g., Ethernet Link, PPP).
	Type                       string             // Type of IP interface (Normal, Loopback, Tunnel, Tunneled).
	IPv4AddressNumberOfEntries int                // Number of entries in IPv4Address table
	IPv4Addresses              []IPv4AddressEntry // List of configured IPv4 addresses.
}

// IPv4AddressEntry represents a single IPv4 address configuration on an IP interface.
type IPv4AddressEntry struct {
	Index          int    // TR-069 index for this address entry
	Enable         bool   // Administrative status.
	Status         string // Operational status (Enabled, Disabled, Error).
	IPAddress      string // The IPv4 address.
	SubnetMask     string // The subnet mask.
	AddressingType string // How the address was assigned (Static, DHCP, IPCP, X_ISPAPP_Dynamic).
}

// IPDiagnostics contains parameters for running IP layer diagnostic tests.
type IPDiagnostics struct {
	IPPing              IPPingDiagnostics     // IP Ping test parameters and results.
	TraceRoute          TraceRouteDiagnostics // Trace Route test parameters and results.
	DownloadDiagnostics DownloadDiagnostics   // Download test parameters and results.
	UploadDiagnostics   UploadDiagnostics     // Upload test parameters and results.
}

// IPPingDiagnostics holds configuration and results for an IP Ping test.
type IPPingDiagnostics struct {
	DiagnosticsState            string // State of the diagnostic (None, Requested, Complete, Error_*).
	Interface                   string // Interface to perform the test over.
	Host                        string // Target host name or IP address.
	NumberOfRepetitions         int    // Number of pings to send.
	Timeout                     int    // Timeout in milliseconds per ping.
	DataBlockSize               int    // Size of ping data payload in bytes.
	DSCP                        int    // DSCP value for test packets.
	SuccessCount                int    // Number of successful pings.
	FailureCount                int    // Number of failed pings.
	AverageResponseTime         int    // Average RTT in milliseconds.
	MinimumResponseTime         int    // Minimum RTT in milliseconds.
	MaximumResponseTime         int    // Maximum RTT in milliseconds.
	AverageResponseTimeDetailed int    // Average RTT in microseconds.
	MinimumResponseTimeDetailed int    // Minimum RTT in microseconds.
	MaximumResponseTimeDetailed string // Maximum RTT in microseconds.
}

// TraceRouteDiagnostics holds configuration and results for a Trace Route test.
type TraceRouteDiagnostics struct {
	DiagnosticsState         string          // State of the diagnostic.
	Interface                string          // Interface for the test.
	Host                     string          // Target host name or IP address.
	NumberOfTries            int             // Number of probes per hop.
	Timeout                  int             // Timeout in milliseconds per probe.
	DataBlockSize            int             // Size of probe data payload.
	DSCP                     int             // DSCP value for probes.
	MaxHopCount              int             // Maximum TTL value.
	ResponseTime             int             // Overall response time in milliseconds (if successful).
	RouteHopsNumberOfEntries int             // Number of entries in RouteHops table
	RouteHops                []TraceRouteHop // List of hops discovered.
}

// TraceRouteHop represents a single hop result in a Trace Route test.
type TraceRouteHop struct {
	Host        string // Host name or IP address of the hop.
	HostAddress string // IP address of the hop (if Host is name).
	ErrorCode   int    // ICMP error code (if any).
	RTTimes     []int  // List of Round Trip Times in milliseconds for probes to this hop.
}

// DownloadDiagnostics holds configuration and results for an HTTP/FTP download test.
type DownloadDiagnostics struct {
	DiagnosticsState                   string                        // State of the diagnostic // Requested or Completed or None or Error_Other
	DownloadURL                        string                        // URL to download from.
	DownloadDiagnosticMaxConnections   int                           // Max supported connections.
	DSCP                               int                           // DSCP value for test packets.
	EthernetPriority                   int                           // Ethernet priority value.
	NumberOfConnections                int                           // Number of connections to use.
	ROMTime                            string                        // Request time (UTC).
	BOMTime                            string                        // Begin of transmission time (UTC).
	EOMTime                            string                        // End of transmission time (UTC).
	TestBytesReceived                  int                           // Bytes received during test period.
	TotalBytesReceived                 int                           // Total bytes received on interface during test.
	TotalBytesSent                     int                           // Total bytes sent on interface during test.
	TestBytesReceivedUnderFullLoading  int                           // Bytes received during full loading period.
	TotalBytesReceivedUnderFullLoading int                           // Total bytes received on interface during full loading.
	TotalBytesSentUnderFullLoading     int                           // Total bytes sent on interface during full loading.
	PeriodOfFullLoading                int                           // Duration of full loading period (microseconds).
	TCPOpenRequestTime                 string                        // TCP connection request time (UTC).
	TCPOpenResponseTime                string                        // TCP connection response time (UTC).
	PerConnectionResultNumberOfEntries int                           // Number of entries in PerConnectionResult table
	EnablePerConnectionResults         bool                          // Flag to enable per-connection results.
	PerConnectionResult                []PerConnectionDownloadResult // Results per connection.
}

// PerConnectionDownloadResult holds download results for a single connection.
type PerConnectionDownloadResult struct {
	ROMTime             string // Request time (UTC).
	BOMTime             string // Begin of transmission time (UTC).
	EOMTime             string // End of transmission time (UTC).
	TestBytesReceived   int    // Bytes received on this connection.
	TCPOpenRequestTime  string // TCP connection request time (UTC).
	TCPOpenResponseTime string // TCP connection response time (UTC).
}

// UploadDiagnostics holds configuration and results for an HTTP/FTP upload test.
type UploadDiagnostics struct {
	DiagnosticsState                   string                      // State of the diagnostic.
	UploadURL                          string                      // URL to upload to.
	UploadDiagnosticsMaxConnections    int                         // Max supported connections.
	DSCP                               int                         // DSCP value for test packets.
	EthernetPriority                   int                         // Ethernet priority value.
	TestFileLength                     int                         // Size of the file to upload in bytes.
	NumberOfConnections                int                         // Number of connections to use.
	ROMTime                            string                      // Request time (UTC).
	BOMTime                            string                      // Begin of transmission time (UTC).
	EOMTime                            string                      // End of transmission time (UTC).
	TestBytesSent                      int                         // Bytes sent during test period.
	TotalBytesReceived                 int                         // Total bytes received on interface during test.
	TotalBytesSent                     int                         // Total bytes sent on interface during test.
	TestBytesSentUnderFullLoading      int                         // Bytes sent during full loading period.
	TotalBytesReceivedUnderFullLoading int                         // Total bytes received on interface during full loading.
	TotalBytesSentUnderFullLoading     int                         // Total bytes sent on interface during full loading.
	PeriodOfFullLoading                int                         // Duration of full loading period (microseconds).
	TCPOpenRequestTime                 string                      // TCP connection request time (UTC).
	TCPOpenResponseTime                string                      // TCP connection response time (UTC).
	PerConnectionResultNumberOfEntries int                         // Number of entries in PerConnectionResult table
	EnablePerConnectionResults         bool                        // Flag to enable per-connection results.
	PerConnectionResult                []PerConnectionUploadResult // Results per connection.
}

// PerConnectionUploadResult holds upload results for a single connection.
type PerConnectionUploadResult struct {
	ROMTime             string // Request time (UTC).
	BOMTime             string // Begin of transmission time (UTC).
	EOMTime             string // End of transmission time (UTC).
	TestBytesSent       int    // Bytes sent on this connection.
	TCPOpenRequestTime  string // TCP connection request time (UTC).
	TCPOpenResponseTime string // TCP connection response time (UTC).
}

// RoutingDevice aggregates routing configurations.
type RoutingDevice struct {
	RouterNumberOfEntries int      // Number of entries in Router table
	Routers               []Router // List of routers (typically one).
}

// Router represents a routing instance on the device.
type Router struct {
	Index                         int                   // TR-069 index for this router
	Enable                        bool                  // Administrative status (usually always true for the main router).
	Status                        string                // Operational status (Enabled, Disabled, Error).
	IPv4ForwardingNumberOfEntries int                   // Number of entries in IPv4Forwarding table
	IPv4Forwarding                []IPv4ForwardingEntry // IPv4 routing table entries.
}

// IPv4ForwardingEntry represents a single entry in the IPv4 routing table.
type IPv4ForwardingEntry struct {
	Index            int    // TR-069 index for this forwarding entry
	Enable           bool   // Administrative status.
	Status           string // Operational status (Enabled, Disabled, Error_Misconfigured).
	StaticRoute      bool   // Indicates if this is a static route.
	DestIPAddress    string // Destination IP address.
	DestSubnetMask   string // Destination subnet mask.
	GatewayIPAddress string // Next hop gateway IP address.
	Interface        string // Egress interface path name.
	Origin           string // How the route was learned (Static, DHCP, RIP, OSPF, X_ISPAPP_*).
}

// HostsDevice provides information about hosts detected by the CPE.
type HostsDevice struct {
	HostNumberOfEntries int         // Number of entries in Host table
	Hosts               []HostEntry // List of detected hosts.
}

// HostEntry represents a single host detected on the network.
type HostEntry struct {
	Index            int    // TR-069 index for this host
	PhysAddress      string // Physical address (e.g., MAC address).
	IPAddress        string // Primary IP address (IPv4 or IPv6).
	DHCPClient       string // Reference to DHCP client entry (if applicable).
	AssociatedDevice string // Reference to the WiFi AssociatedDevice entry (if applicable).
	Layer1Interface  string // Reference to the Layer 1 interface the host is connected to.
	Layer3Interface  string // Reference to the Layer 3 interface the host is connected to.
	HostName         string // Host name (if known).
}

// DNSDevice aggregates DNS client configuration.
type DNSDevice struct {
	Client DNSClient // DNS client settings.
}

// DNSClient contains settings for the CPE's internal DNS resolver.
type DNSClient struct {
	ServerNumberOfEntries int              // Number of entries in Server table
	Servers               []DNSServerEntry // List of DNS servers to use.
}

// DNSServerEntry represents a single DNS server configuration.
type DNSServerEntry struct {
	Index     int    // TR-069 index for this DNS server
	Enable    bool   // Administrative status.
	Status    string // Operational status (Enabled, Disabled, Error).
	DNSServer string // IP address of the DNS server.
	Type      string // How the server was configured (Static, DHCPv4, DHCPv6, IPCP, RouterAdvertisement, X_ISPAPP_Dynamic).
}

// DHCPv4Device aggregates DHCPv4 client and server configurations.
type DHCPv4Device struct {
	ClientNumberOfEntries int            // Number of entries in Client table
	Clients               []DHCPv4Client // List of DHCPv4 client configurations.
	Server                DHCPv4Server   // DHCPv4 server configuration.
}

// DHCPv4Client represents a DHCPv4 client configured on a specific interface.
type DHCPv4Client struct {
	Index      int    // TR-069 index for this client
	Enable     bool   // Administrative status.
	Interface  string // IP Interface the client runs on.
	Status     string // Operational status (Enabled, Disabled, Error_Misconfigured).
	DHCPStatus string // DHCP state machine status (Init, Bound, Renewing, etc.).
	IPAddress  string // IP address obtained from server.
	SubnetMask string // Subnet mask obtained from server.
	IPRouters  string // Router IP addresses obtained from server (comma-separated).
	DNSServers string // DNS server IP addresses obtained from server (comma-separated).
	DHCPServer string // IP address of the DHCP server providing the lease.
}

// DHCPv4Server contains the overall DHCPv4 server configuration, including pools.
type DHCPv4Server struct {
	PoolNumberOfEntries int                // Number of entries in Pool table
	Pools               []DHCPv4ServerPool // List of DHCPv4 address pools.
}

// DHCPv4ServerPool represents a pool of IPv4 addresses managed by the DHCP server.
// Note: Mikrotik maps this differently, often linking one pool per server instance.
type DHCPv4ServerPool struct {
	Index                        int                   // TR-069 index for this pool
	Enable                       bool                  // Administrative status.
	Status                       string                // Operational status (Enabled, Disabled, Error_Misconfigured).
	Interface                    string                // IP Interface the server listens on for this pool.
	MinAddress                   string                // Start of the address pool range.
	MaxAddress                   string                // End of the address pool range.
	SubnetMask                   string                // Subnet mask for clients in this pool.
	DNSServers                   []string              // DNS servers offered to clients.
	DomainName                   string                // Domain name offered to clients.
	IPRouters                    []string              // Routers (gateways) offered to clients.
	LeaseTime                    int                   // Lease duration in seconds.
	StaticAddressNumberOfEntries int                   // Number of entries in StaticAddress table
	ClientNumberOfEntries        int                   // Number of entries in Client table
	StaticAddresses              []DHCPv4StaticAddress // Statically assigned addresses within this pool.
	Clients                      []DHCPv4ServerClient  // Currently active client leases from this pool.
}

// DHCPv4StaticAddress represents a manually allocated IP address for a specific client MAC.
type DHCPv4StaticAddress struct {
	Index  int    // TR-069 index for this static address
	Enable bool   // Administrative status.
	Chaddr string // Client hardware address (MAC address).
	Yiaddr string // Statically assigned IP address for this client.
}

// DHCPv4ServerClient represents an active lease held by a DHCP client.
type DHCPv4ServerClient struct {
	Index                      int                       // TR-069 index for this server client
	Chaddr                     string                    // Client hardware address (MAC address).
	IPv4AddressNumberOfEntries int                       // Number of entries in IPv4Address table
	IPv4Addresses              []DHCPv4ClientIPv4Address // List of IPs leased to this client (usually one).
}

// DHCPv4ClientIPv4Address holds details about a specific IP address leased to a client.
type DHCPv4ClientIPv4Address struct {
	Index              int    // TR-069 index for this client IPv4 address
	IPAddress          string // The leased IPv4 address.
	LeaseTimeRemaining string // Time remaining on the lease (ISO 8601 format or duration).
}

// XMikrotikConnTrack holds connection tracking information.
type XMikrotikConnTrack struct {
	TotalEntries int // Current number of tracked connections.
}

// XMikrotikFilter aggregates firewall filter chains and rules.
type XMikrotikFilter struct {
	ChainNumberOfEntries int                      // Number of entries in Chain table
	Chains               []XMikrotikFirewallChain // List of filter chains (input, output, forward, custom).
}

// XMikrotikNAT aggregates firewall NAT chains and rules.
type XMikrotikNAT struct {
	ChainNumberOfEntries int                      // Number of entries in Chain table
	Chains               []XMikrotikFirewallChain // List of NAT chains (srcnat, dstnat, custom).
}

// XMikrotikFirewallChain represents a firewall chain (filter or NAT).
type XMikrotikFirewallChain struct {
	Enable              bool                    // Administrative status.
	Name                string                  // Name of the chain (e.g., "input", "forward", "srcnat").
	RuleNumberOfEntries int                     // Number of entries in Rule table
	Rules               []XMikrotikFirewallRule // Ordered list of rules in this chain.
}

// XMikrotikFirewallRule represents a single firewall rule (filter or NAT).
// Combines parameters from both filter and NAT contexts where applicable.
type XMikrotikFirewallRule struct {
	Enable                 bool     // Administrative status.
	Order                  int      // Rule order within the chain.
	Description            string   // User comment for the rule.
	Target                 string   // Action to take (Accept, Drop, Reject, Return, Log, Masquerade, SNAT, DNAT, etc.).
	TargetChain            string   // Target chain if Target is TargetChain/Jump.
	Log                    bool     // Whether to log packets matching this rule.
	SourceInterfaceGroup   string   // Source interface group (all, all-ethernet, etc.).
	SourceInterface        string   // Specific source interface path name.
	SourceInterfaceExclude bool     // Negate source interface match.
	DestInterfaceGroup     string   // Destination interface group.
	DestInterface          string   // Specific destination interface path name.
	DestInterfaceExclude   bool     // Negate destination interface match.
	DestIPRange            string   // Destination IP address/range/subnet.
	DestIPExclude          bool     // Negate destination IP match.
	SourceIPRange          string   // Source IP address/range/subnet.
	SourceIPExclude        bool     // Negate source IP match.
	Protocol               int      // Protocol number (-1 for any).
	ProtocolExclude        bool     // Negate protocol match.
	DestPortList           string   // Destination port(s) or range(s).
	DestPortExclude        bool     // Negate destination port match.
	SourcePortList         string   // Source port(s) or range(s).
	SourcePortExclude      bool     // Negate source port match.
	ConnState              []string // Connection states to match (New, Established, Related, Invalid, Untracked).
	ConnStateExclude       bool     // Negate connection state match.
	// NAT specific parameters
	ToAddresses string // Target address(es) for NAT actions.
	ToPorts     string // Target port(s) for NAT actions.
}

// XMikrotikInterfaceDevice aggregates Mikrotik generic interfaces.
type XMikrotikInterfaceDevice struct {
	GenericNumberOfEntries int                         // Number of entries in Generic table
	Generics               []XMikrotikGenericInterface // List of generic interfaces.
}

// XMikrotikGenericInterface represents RouterOS interfaces not covered by standard TR-069 models.
type XMikrotikGenericInterface struct {
	Index       int    // TR-069 index for this generic interface
	Enable      bool   // Administrative status.
	Status      string // Operational status.
	Name        string // Interface name in RouterOS.
	LowerLayers string // Reference to lower layer interfaces.
}

// XMikrotikMonitorDevice aggregates traffic monitoring configurations.
type XMikrotikMonitorDevice struct {
	TrafficNumberOfEntries int                       // Number of entries in Traffic table
	TrafficMonitors        []XMikrotikTrafficMonitor // List of traffic monitors.
}

// XMikrotikTrafficMonitor represents traffic monitoring on a specific interface.
type XMikrotikTrafficMonitor struct {
	Index     int    // TR-069 index for this traffic monitor
	Enable    bool   // Whether monitoring is enabled for this interface.
	Interface string // Path name of the interface being monitored.
	RxRate    int    // Current receive rate (Kbps).
	TxRate    int    // Current transmit rate (Kbps).
	MaxRxRate int    // Maximum recorded receive rate (Kbps).
	MaxTxRate int    // Maximum recorded transmit rate (Kbps).
}

// WANDevice represents the WAN device configuration and contains WAN connection devices.
type WANDevice struct {
	WANConnectionDeviceNumberOfEntries int                   // Number of entries in the WANConnectionDevice table
	WANConnectionDevices               []WANConnectionDevice // List of WAN connection devices.
}

// WANConnectionDevice represents a WAN connection device that contains WAN IP connections.
type WANConnectionDevice struct {
	Index                          int               // TR-069 index for this WAN connection device
	WANIPConnectionNumberOfEntries int               // Number of entries in the WANIPConnection table
	WANIPConnections               []WANIPConnection // List of WAN IP connections.
}

// WANIPConnection represents a WAN IP connection with its external IP address.
type WANIPConnection struct {
	Index                      int    // TR-069 index for this WAN IP connection
	Enable                     bool   // Administrative status of the connection.
	ConnectionStatus           string // Current status of the connection (Connected, Disconnected, etc.).
	ConnectionType             string // Type of connection (IP_Routed, IP_Bridged, etc.).
	Name                       string // User-readable name for the connection.
	LastConnectionError        string // Last connection error.
	AutoDisconnectTime         int    // Time in seconds before auto-disconnect (0=disabled).
	IdleDisconnectTime         int    // Time in seconds of inactivity before auto-disconnect (0=disabled).
	ExternalIPAddress          string // Current external IPv4 address assigned to the connection.
	SubnetMask                 string // Subnet mask associated with the external IP address.
	DefaultGateway             string // Default gateway IP address.
	DNSEnabled                 bool   // Whether DNS is enabled for this connection.
	DNSOverrideAllowed         bool   // Whether DNS override is allowed.
	DNSServers                 string // DNS server addresses (comma-separated).
	MaxMTUSize                 int    // Maximum Transmission Unit size.
	MACAddress                 string // MAC address used for this connection.
	ConnectionTrigger          string // Trigger mechanism for connection (AlwaysOn, OnDemand, Manual).
	RouteProtocolRx            string // Route protocol for received routes.
	ShapingRate                int    // Traffic shaping rate in bits per second.
	ShapingBurstSize           int    // Traffic shaping burst size in bytes.
	PortMappingNumberOfEntries int    // Number of port mapping entries.
}

// FirewallDevice aggregates firewall configurations. Includes Mikrotik extensions.
type FirewallDevice struct {
	X_ISPAPP_ConnTrack XMikrotikConnTrack // Mikrotik connection tracking info.
	X_ISPAPP_Filter    XMikrotikFilter    // Mikrotik filter rules.
	X_ISPAPP_NAT       XMikrotikNAT       // Mikrotik NAT rules.
}
