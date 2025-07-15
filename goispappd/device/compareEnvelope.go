package device

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Niceblueman/goispappd/soap"
)

// CompareEnvelope compares device with soap.GetParameterValuesResponse parameters and generates
// a SetParameterValues containing only the actual differences with proper validation
// Time complexity: O(n) where n is the number of parameters in the envelope
// Space complexity: O(d) where d is the number of differences found
func (dev *Device) CompareEnvelope(_envelope *soap.GetParameterValuesResponse) (*soap.SetParameterValues, *soap.GetParameterValuesResponse) {
	if dev == nil || _envelope == nil {
		return nil, nil
	}
	envelope := _envelope
	// Pre-allocate slice with reasonable capacity to minimize allocations
	var differences []ParameterValueStruct

	// Build a map of current device values by TR-069 parameter path for efficient lookup
	deviceParameterMap := dev.buildParameterMap()

	// Process each parameter from the envelope
	for _, envParam := range envelope.ParameterList.Parameters {
		paramName := strings.TrimSpace(envParam.Name)
		envValue := strings.TrimSpace(envParam.Value.Content)

		// Skip empty or invalid parameter names
		if paramName == "" {
			continue
		}

		// Skip read-only parameters to avoid TR-069 faults
		if !isWritableParameter(paramName) {
			continue
		}

		// Get the current device value for this parameter path
		deviceValue, exists := deviceParameterMap[paramName]
		if exists && deviceValue != "" { // Only add non-empty values to avoid noise
			// Validate and compare values to avoid fake differences
			if isValidDifference(envValue, deviceValue) {
				addDifference(&differences, paramName, deviceValue)
			}
		}
	}

	if len(differences) == 0 {
		return nil, envelope
	}

	// Build SetParameterValues response
	result := &soap.SetParameterValues{
		ParameterList: struct {
			Params []struct {
				Name  string `xml:"cwmp:Name"`
				Value string `xml:"cwmp:Value"`
			} `xml:"cwmp:ParameterValueStruct"`
		}{},
		ParameterKey: "EnvelopeCompare_" + strconv.Itoa(len(differences)) + "_changes",
	}

	// Convert our internal format to SOAP format
	result.ParameterList.Params = make([]struct {
		Name  string `xml:"cwmp:Name"`
		Value string `xml:"cwmp:Value"`
	}, len(differences))

	for i, diff := range differences {
		result.ParameterList.Params[i].Name = diff.Name
		result.ParameterList.Params[i].Value = diff.Value
		for j, param := range envelope.ParameterList.Parameters {
			if param.Name == diff.Name {
				// Copy the xsi:type from the envelope to maintain consistency
				envelope.ParameterList.Parameters[j].Value.Content = diff.Value
				break
			}
		}
	}
	return result, envelope
}

// buildParameterMap creates a map of TR-069 parameter paths to their current device values
// Uses the same logic as the Compare function for consistency
func (dev *Device) buildParameterMap() map[string]string {
	paramMap := make(map[string]string)

	// Use reflection to traverse the device structure
	currentVal := reflect.ValueOf(dev).Elem()
	currentType := currentVal.Type()

	// Build parameter map recursively
	buildParameterMapRecursive(currentVal, currentType, "Device", paramMap, make(map[reflect.Type]bool))

	return paramMap
}

// buildParameterMapRecursive recursively builds the parameter map using the same logic as compareStructFields
func buildParameterMapRecursive(currentVal reflect.Value, structType reflect.Type, pathPrefix string, paramMap map[string]string, visited map[reflect.Type]bool) {
	// Prevent infinite recursion
	if visited[structType] {
		return
	}
	visited[structType] = true
	defer func() { delete(visited, structType) }()

	// Process all struct fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldVal := currentVal.Field(i)
		fieldPath := buildFieldPath(pathPrefix, field.Name, field.Type)

		// Add field value to parameter map based on type
		addFieldToParameterMap(fieldVal, fieldPath, field.Type, paramMap, visited)
	}
}

// addFieldToParameterMap adds a field value to the parameter map based on its type
func addFieldToParameterMap(fieldVal reflect.Value, paramPath string, fieldType reflect.Type, paramMap map[string]string, visited map[reflect.Type]bool) {
	// Handle nil pointers
	if fieldVal.Kind() == reflect.Ptr {
		if fieldVal.IsNil() {
			paramMap[paramPath] = ""
			return
		}
		fieldVal = fieldVal.Elem()
		fieldType = fieldType.Elem()
	}

	// Handle different field types
	switch fieldType.Kind() {
	case reflect.String:
		paramMap[paramPath] = fieldVal.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		paramMap[paramPath] = strconv.FormatInt(fieldVal.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		paramMap[paramPath] = strconv.FormatUint(fieldVal.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		paramMap[paramPath] = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
	case reflect.Bool:
		paramMap[paramPath] = strconv.FormatBool(fieldVal.Bool())
	case reflect.Slice:
		// For slices, use the parent path directly (don't add the field name again)
		// This handles cases like HostsDevice.Hosts where the TR-069 path should be
		// Device.Hosts.Host.{index} not Device.Hosts.Hosts.Host.{index}
		addSliceToParameterMap(fieldVal, paramPath, paramMap, visited)
	case reflect.Struct:
		// Recursively process nested structs
		buildParameterMapRecursive(fieldVal, fieldType, paramPath, paramMap, visited)
	default:
		// Fallback for other types
		paramMap[paramPath] = getStringValue(fieldVal)
	}
}

// addSliceToParameterMap handles slice fields in parameter map building
func addSliceToParameterMap(sliceVal reflect.Value, pathPrefix string, paramMap map[string]string, visited map[reflect.Type]bool) {
	if sliceVal.Len() == 0 {
		return
	}

	elementType := sliceVal.Index(0).Type()

	// Handle struct elements with Index field (sparse array approach)
	if elementType.Kind() == reflect.Struct {
		if indexField, hasIndex := elementType.FieldByName("Index"); hasIndex && indexField.Type.Kind() == reflect.Int {
			addIndexedSliceToParameterMap(sliceVal, pathPrefix, paramMap, visited)
			return
		}
	}

	// Traditional array handling for non-indexed elements
	for i := 0; i < sliceVal.Len(); i++ {
		indexPath := pathPrefix + "." + strconv.Itoa(i+1) // TR-069 uses 1-based indexing
		elem := sliceVal.Index(i)
		addFieldToParameterMap(elem, indexPath, elem.Type(), paramMap, visited)
	}
}

// addIndexedSliceToParameterMap handles slices with Index fields
// Uses TR-069 naming convention where indexed arrays use singular form in path
func addIndexedSliceToParameterMap(sliceVal reflect.Value, pathPrefix string, paramMap map[string]string, visited map[reflect.Type]bool) {
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Struct {
			indexField := elem.FieldByName("Index")
			if indexField.IsValid() && indexField.Kind() == reflect.Int {
				index := int(indexField.Int())

				// Convert plural field name to singular for TR-069 indexed array paths
				// Replace the last part of the path (plural) with singular form
				singularPath := replacePluralWithSingular(pathPrefix)
				indexPath := singularPath + "." + strconv.Itoa(index)
				addFieldToParameterMap(elem, indexPath, elem.Type(), paramMap, visited)
			}
		}
	}
}

// replacePluralWithSingular replaces the last plural field name in a path with its singular form
func replacePluralWithSingular(pathPrefix string) string {
	parts := strings.Split(pathPrefix, ".")
	if len(parts) == 0 {
		return pathPrefix
	}

	lastPart := parts[len(parts)-1]
	singularName := ""

	// Convert plural to singular following TR-069 conventions
	switch lastPart {
	case "Hosts":
		singularName = "Host"
	case "Interfaces":
		singularName = "Interface"
	case "AccessPoints":
		singularName = "AccessPoint"
	case "Radios":
		singularName = "Radio"
	case "SSIDs":
		singularName = "SSID"
	case "AssociatedDevices":
		singularName = "AssociatedDevice"
	case "IPv4Addresses":
		singularName = "IPv4Address"
	case "Routers":
		singularName = "Router"
	case "IPv4Forwardings":
		singularName = "IPv4Forwarding"
	default:
		// Default: remove 's' if it ends with 's'
		if strings.HasSuffix(lastPart, "s") && len(lastPart) > 1 {
			singularName = lastPart[:len(lastPart)-1]
		} else {
			singularName = lastPart
		}
	}

	// Replace the last part with singular form
	parts[len(parts)-1] = singularName
	return strings.Join(parts, ".")
}

// isValidDifference validates whether a difference is genuine and not a false positive
// This helps avoid fake differences due to type conversions, formatting, etc.
func isValidDifference(deviceValue, envelopeValue string) bool {
	// Both empty - no difference
	if deviceValue == "" && envelopeValue == "" {
		return false
	}

	// Exact match - no difference
	if deviceValue == envelopeValue {
		return false
	}

	// Normalize and compare boolean values
	if normalizedDeviceVal := normalizeBooleanValue(deviceValue); normalizedDeviceVal != "" {
		if normalizedEnvVal := normalizeBooleanValue(envelopeValue); normalizedEnvVal != "" {
			return normalizedDeviceVal != normalizedEnvVal
		}
	}

	// Normalize and compare numeric values
	if normalizedDeviceVal := normalizeNumericValue(deviceValue); normalizedDeviceVal != "" {
		if normalizedEnvVal := normalizeNumericValue(envelopeValue); normalizedEnvVal != "" {
			return normalizedDeviceVal != normalizedEnvVal
		}
	}

	// For string values, trim whitespace and compare
	deviceTrimmed := strings.TrimSpace(deviceValue)
	envelopeTrimmed := strings.TrimSpace(envelopeValue)

	// Consider empty and missing values as equivalent
	if deviceTrimmed == "" && envelopeTrimmed == "" {
		return false
	}

	return deviceTrimmed != envelopeTrimmed
}

// normalizeBooleanValue normalizes boolean representations to a standard format
func normalizeBooleanValue(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "true", "1", "yes", "on", "enabled":
		return "true"
	case "false", "0", "no", "off", "disabled":
		return "false"
	default:
		return "" // Not a boolean value
	}
}

// normalizeNumericValue normalizes numeric representations
func normalizeNumericValue(value string) string {
	value = strings.TrimSpace(value)

	// Try to parse as integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return strconv.FormatInt(intVal, 10)
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		// Use same formatting as getStringValue function
		return strconv.FormatFloat(floatVal, 'f', -1, 64)
	}

	return "" // Not a numeric value
}

// isWritableParameter checks if a TR-069 parameter is writable
// Read-only parameters should not be included in SetParameterValues requests
func isWritableParameter(paramPath string) bool {
	readonlyList := []string{
		"*ProcessStatus*",
		"*.Stats.*",
		"*.X_MIKROTIK_Stats.*",
		"*.SignalStrength",
		"*Signal*",
		"*Rate*",
		"*MemoryStatus*",
		"*DiagnosticsState*",
		"*InterfaceStack*",
		"*UpTime*",
		"*SoftwareVersion*",
		"*HardwareVersion*",
		"*SerialNumber*",
		"*NumberOfEntries*",
		"*.Diagnostics.*",
		"*.X_MIKROTIK_ConnTrack.*",
		"*.X_MIKROTIK_Filter.*",
		"*.X_MIKROTIK_NAT.*",
		"*.AssociatedDevice*",
		"*.Status",
		"*AuthenticationState*",
		"*.IPv4F",
		"*.AddressingType",
		"*.IPv4Forwarding.*",
		"*.IP.Interface.*.IPv4Address.*.Enable",
		"*.PhysAddress*",
		"Hosts.*.Type",
		"Device.DNS.Client.Server.*", // dynamic parameter
		"Device.IP.Interface.*.IPv4Address.*.IPAddress",  // dynamic parameter
		"Device.IP.Interface.*.IPv4Address.*.SubnetMask", // dynamic parameter
		"*DNS.Client.Server.*.Type",
		"*.ModelName",
		"*.DeviceSummary",
		"*.X_MIKROTIK_ArchName",
		"*.X_MIKROTIK_BrandingPckgBuildTime",
		"*.X_MIKROTIK_AutosupoutTime",
		"*.VendorConfigFile.*.Name",
		"*.VendorConfigFile.*.Description",
		"*.VendorConfigFile.*.UseForBackupRestore",
		"*.MACAddress",
		"*.Ethernet.Interface.*.MACAddress",
		"*.Ethernet.Interface.*.X_MIKROTIK_MediaType",
		"*.Ethernet.Interface.*.X_MIKROTIK_Speed",
		"*.Ethernet.Interface.*.X_MIKROTIK_LinkStatus",
		"*.Ethernet.Interface.*.X_MIKROTIK_Stats.*",
		"*.Ethernet.Interface.*.X_MIKROTIK_ConnTrack.*",
		"*.DeviceInfo.Description",
		"Device.WiFi.Radio.*",
		"Device.WiFi.SSID.*",
		"Device.WiFi.AccessPoint.*.Security.ModesSupported",
		"Device.IP.Interface.*.Type",
		"Device.IP.Interface.*.Type",
		"Device.Routing.Router.*.Enable",
		"Device.DHCPv4.Client.*.IPAddress",
		"Device.DHCPv4.Client.*.SubnetMask",
		"Device.DHCPv4.Client.*.IPRouters",
		"Device.DHCPv4.Client.*.DNSServers",
		"Device.DHCPv4.Client.*.DHCPServer",
		"Device.X_ISPAPP_Interface.Generic.*.Enable",
		"Device.X_ISPAPP_Interface.Generic.*.Name",
		"Device.WiFi.AccessPoint.*.SSIDReference", // This is a special case, it cannot be changed
		"Device.Ethernet.Interface.*.X_MIKROTIK_LinkDowns",
		"*.NeighboringWiFiDiagnostic.*",
		"Device.IP.Interface.*.LowerLayers",
		"Device.DHCPv4.Client.*.Interface",
		"Device.Ethernet.Interface.*",
		"Device.Ethernet.Link.*",
	}
	// Stats parameters are always read-only
	for _, roParam := range readonlyList {
		// Convert wildcard patterns to regex patterns
		regexPattern := strings.ReplaceAll(roParam, "*", ".*?")
		regexPattern = "^" + regexPattern + "$"

		if matched, _ := regexp.MatchString(regexPattern, paramPath); matched {
			return false
		}
	}
	// Default to writable for other parameters
	return true
}
