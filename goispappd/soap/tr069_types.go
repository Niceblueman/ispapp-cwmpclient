package soap

// TR069DataTypes defines the XSD types used in TR-069 parameter values
const (
	TR069TypeString      = "xsd:string"
	TR069TypeInt         = "xsd:int"
	TR069TypeUnsignedInt = "xsd:unsignedInt"
	TR069TypeBoolean     = "xsd:boolean"
	TR069TypeDateTime    = "xsd:dateTime"
	TR069TypeFloat       = "xsd:float"
	TR069TypeDouble      = "xsd:double"
	TR069TypeBase64      = "xsd:base64Binary"
)

// TR069ParameterTypeRules maps parameter path patterns to their expected TR-069 types
var TR069ParameterTypeRules = map[string]string{
	// Boolean parameters
	`.*\.Enable$`:                   TR069TypeBoolean,
	`.*\.Status$`:                   TR069TypeString, // Status is typically enum string
	`.*\.X_MIKROTIK_.*Enable$`:      TR069TypeBoolean,
	`.*\.PeriodicInformEnable$`:     TR069TypeBoolean,
	`.*\.AutoChannelEnable$`:        TR069TypeBoolean,
	`.*\.AutoChannelSupported$`:     TR069TypeBoolean,
	`.*\.SSIDAdvertisementEnabled$`: TR069TypeBoolean,
	`.*\.AliasBasedAddressing$`:     TR069TypeBoolean,
	`.*\.AuthenticationState$`:      TR069TypeBoolean,
	`.*\.StaticRoute$`:              TR069TypeBoolean,
	`.*\.UseForBackupRestore$`:      TR069TypeBoolean,
	`.*\.Log$`:                      TR069TypeBoolean,
	`.*\..*Exclude$`:                TR069TypeBoolean,

	// Unsigned integer parameters (counters, indices, ports, etc.)
	`.*NumberOfEntries$`:          TR069TypeUnsignedInt,
	`.*\.Index$`:                  TR069TypeUnsignedInt,
	`.*\.PeriodicInformInterval$`: TR069TypeUnsignedInt,
	`.*\.Channel$`:                TR069TypeUnsignedInt,
	`.*\.CurrentBitRate$`:         TR069TypeUnsignedInt,
	`.*\.Port$`:                   TR069TypeUnsignedInt,
	`.*\..*Port.*$`:               TR069TypeUnsignedInt,
	`.*\.UpTime$`:                 TR069TypeUnsignedInt,
	`.*\.Total$`:                  TR069TypeUnsignedInt,
	`.*\.Free$`:                   TR069TypeUnsignedInt,
	`.*\.CPUUsage$`:               TR069TypeUnsignedInt,
	`.*\.LeaseTime$`:              TR069TypeUnsignedInt,
	`.*\.TestFileLength$`:         TR069TypeUnsignedInt,
	`.*\.NumberOfRepetitions$`:    TR069TypeUnsignedInt,
	`.*\.Timeout$`:                TR069TypeUnsignedInt,
	`.*\.DataBlockSize$`:          TR069TypeUnsignedInt,
	`.*\.DSCP$`:                   TR069TypeUnsignedInt,
	`.*\.EthernetPriority$`:       TR069TypeUnsignedInt,
	`.*\.NumberOfConnections$`:    TR069TypeUnsignedInt,
	`.*\.NumberOfTries$`:          TR069TypeUnsignedInt,
	`.*\.MaxHopCount$`:            TR069TypeUnsignedInt,
	`.*\.Order$`:                  TR069TypeUnsignedInt,
	`.*\.Protocol$`:               TR069TypeUnsignedInt, // Protocol numbers are unsigned
	`.*\.X_MIKROTIK_LinkDowns$`:   TR069TypeUnsignedInt,

	// Stats parameters (all unsigned counters)
	`.*\.Stats\..*$`:            TR069TypeUnsignedInt,
	`.*\.X_MIKROTIK_Stats\..*$`: TR069TypeUnsignedInt,

	// Signal strength and cellular parameters (signed integers)
	`.*\.SignalStrength$`: TR069TypeInt,
	`.*\.RSSI$`:           TR069TypeInt,
	`.*\.RSCP$`:           TR069TypeInt,
	`.*\.ECNO$`:           TR069TypeInt,
	`.*\.SINR$`:           TR069TypeInt,
	`.*\.RSRP$`:           TR069TypeInt,
	`.*\.RSRQ$`:           TR069TypeInt,
	`.*\.SNR$`:            TR069TypeInt,
	`.*\.SignalToNoise$`:  TR069TypeInt,
	`.*\.Noise$`:          TR069TypeInt,
	`.*\.TxPUCCH$`:        TR069TypeInt,
	`.*\.TxPUSCH$`:        TR069TypeInt,
	`.*\.TxSRS$`:          TR069TypeInt,
	`.*\.TxPRACH$`:        TR069TypeInt,

	// Cellular specific unsigned parameters
	`.*\.Band$`:           TR069TypeUnsignedInt,
	`.*\.Fcn$`:            TR069TypeUnsignedInt,
	`.*\.Bandwidth$`:      TR069TypeUnsignedInt,
	`.*\.PhysicalCellId$`: TR069TypeUnsignedInt,
	`.*\.CQI$`:            TR069TypeUnsignedInt,
	`.*\.RI$`:             TR069TypeUnsignedInt,
	`.*\.MCS$`:            TR069TypeUnsignedInt,
	`.*\.TBS$`:            TR069TypeUnsignedInt,
	`.*\.RBs$`:            TR069TypeUnsignedInt,
	`.*\.CellId$`:         TR069TypeUnsignedInt,

	// Rates and traffic monitoring
	`.*Rate$`:       TR069TypeUnsignedInt,
	`.*\..*Rate.*$`: TR069TypeUnsignedInt,

	// DateTime parameters
	`.*Time$`:       TR069TypeDateTime,
	`.*\..*Time.*$`: TR069TypeDateTime,

	// Version numbers as float
	`.*Version$`: TR069TypeString, // Usually string format like "2.4"

	// Default to string for most other parameters
	`.*\..*$`: TR069TypeString,
}

// BooleanValues maps string representations to boolean values
var BooleanValues = map[string]bool{
	"true":     true,
	"false":    false,
	"1":        true,
	"0":        false,
	"yes":      true,
	"no":       false,
	"on":       true,
	"off":      false,
	"enabled":  true,
	"disabled": false,
}

// EnumParameters defines parameters that should be treated as string enums
var EnumParameters = map[string]bool{
	"Status":                             true,
	"ConnectionStatus":                   true,
	"DHCPStatus":                         true,
	"DiagnosticsState":                   true,
	"OperatingFrequencyBand":             true,
	"Target":                             true,
	"ModeEnabled":                        true,
	"Type":                               true,
	"Origin":                             true,
	"AddressingType":                     true,
	"EncryptionProtocol":                 true,
	"ConnectionTrigger":                  true,
	"OperatingChannelBandwidth":          true,
	"X_MIKROTIK_Modulation":              true,
	"X_MIKROTIK_5G_Modulation":           true,
	"X_MIKROTIK_CurrentAccessTechnology": true,
	"X_MIKROTIK_Type":                    true,
	"X_MIKROTIK_Antenna":                 true,
	"X_MIKROTIK_SkipDFSChannels":         true,
}
