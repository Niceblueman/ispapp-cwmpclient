package device

import (
	"testing"

	"github.com/Niceblueman/goispappd/soap"
)

func TestCompareEnvelopeBasic(t *testing.T) {
	// Create a simple test device
	device := &Device{
		DeviceInfo: DeviceInfo{
			Manufacturer: "MikroTik",
			ModelName:    "RB952Ui-5ac2nD",
		},
		Hosts: HostsDevice{
			Hosts: []HostEntry{
				{
					Index:            153,
					AssociatedDevice: "Device.WiFi.AccessPoint.2.AssociatedDevice.120",
					IPAddress:        "192.168.1.1",
				},
			},
		},
		// Device.Routing.Router.1.IPv4Forwarding.2.DestIPAddress
		Routing: RoutingDevice{
			Routers: []Router{
				{
					Index: 1,
					IPv4Forwarding: []IPv4ForwardingEntry{
						{
							Index:         2,
							DestIPAddress: "192.168.1.1",
						},
					},
				},
			},
		},
	}

	// Create an envelope with one different value
	envelope := &soap.GetParameterValuesResponse{
		ParameterList: soap.ParameterList{
			Parameters: []soap.ParameterValueStruct{
				{
					Name: "Device.DeviceInfo.Manufacturer",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "MikroTik",
					},
				},
				{
					Name: "Device.DeviceInfo.ModelName",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "NEW-MODEL", // Different value
					},
				},
				{
					Name: "Device.Hosts.Host.153.AssociatedDevice",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "Device.WiFi.AccessPoint.2.AssociatedDevice.120",
					},
				},
				{
					Name: "Device.Hosts.Host.153.IPAddress",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "192.168.1.1",
					},
				},
				{
					Name: "Device.Routing.Router.1.IPv4Forwarding.2.DestIPAddress",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "192.168.1.2",
					},
				},
			},
		},
	}

	// Test the comparison
	result, _ := device.CompareEnvelope(envelope)

	// Validate basic functionality
	if result == nil {
		t.Fatal("Expected differences but got nil result")
	}

	if len(result.ParameterList.Params) == 0 {
		t.Error("Expected at least one difference")
	}

	t.Logf("Found %d differences", len(result.ParameterList.Params))
	for _, param := range result.ParameterList.Params {
		t.Logf("Difference: %s = %s", param.Name, param.Value)
	}
}

func TestCompareEnvelopeNil(t *testing.T) {
	device := &Device{}

	// Test with nil envelope
	result, _ := device.CompareEnvelope(nil)
	if result != nil {
		t.Error("Expected nil result for nil envelope")
	}

	// Test with nil device
	result, _ = (*Device)(nil).CompareEnvelope(&soap.GetParameterValuesResponse{})
	if result != nil {
		t.Error("Expected nil result for nil device")
	}
}

func TestCompareEnvelopeValidation(t *testing.T) {
	// Test validation features to avoid fake differences
	device := &Device{
		DeviceInfo: DeviceInfo{
			Manufacturer: "MikroTik",
		},
		ManagementServer: ManagementServer{
			PeriodicInformEnable:   true,
			PeriodicInformInterval: 5,
		},
	}

	envelope := &soap.GetParameterValuesResponse{
		ParameterList: soap.ParameterList{
			Parameters: []soap.ParameterValueStruct{
				{
					Name: "Device.DeviceInfo.Manufacturer",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "MikroTik", // Same value - should not create difference
					},
				},
				{
					Name: "Device.ManagementServer.PeriodicInformEnable",
					Value: soap.Value{
						Type:    "xsd:boolean",
						Content: "1", // Boolean equivalent of true - should not create difference
					},
				},
				{
					Name: "Device.ManagementServer.PeriodicInformInterval",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "5.0", // Numeric equivalent - should not create difference
					},
				},
				{
					Name: "Device.DeviceInfo.ModelName",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "   ", // Just whitespace - should not create difference for empty field
					},
				},
				{
					Name: "Device.NewParameter",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "NewValue", // Actually new parameter - should create difference
					},
				},
			},
		},
	}

	result, _ := device.CompareEnvelope(envelope)

	// Should only have one difference (the new parameter)
	if result == nil {
		t.Fatal("Expected differences but got nil result")
	}

	if len(result.ParameterList.Params) != 1 {
		t.Errorf("Expected 1 difference, got %d", len(result.ParameterList.Params))
		for _, param := range result.ParameterList.Params {
			t.Logf("Unexpected difference: %s = %s", param.Name, param.Value)
		}
	}

	// Verify it's the new parameter
	if result.ParameterList.Params[0].Name != "Device.NewParameter" {
		t.Errorf("Expected new parameter difference, got: %s", result.ParameterList.Params[0].Name)
	}
}

// TestCompareEnvelopeRealData tests with real data from the XML file
func TestCompareEnvelopeRealData(t *testing.T) {
	// Create a device based on real XML data
	device := &Device{
		DeviceInfo: DeviceInfo{
			Manufacturer:    "MikroTik",
			ModelName:       "RB952Ui-5ac2nD",
			ManufacturerOUI: "4C:5E:0C",
		},
		Hosts: HostsDevice{
			HostNumberOfEntries: 6,
			Hosts: []HostEntry{
				{
					Index:            153,
					PhysAddress:      "B8:27:EB:89:CC:2D",
					IPAddress:        "192.168.1.100",
					AssociatedDevice: "Device.WiFi.AccessPoint.2.AssociatedDevice.120",
					HostName:         "raspberry-pi",
					Layer1Interface:  "Device.WiFi.SSID.2",
					Layer3Interface:  "Device.IP.Interface.1",
					DHCPClient:       "",
				},
			},
		},
		Routing: RoutingDevice{
			RouterNumberOfEntries: 1,
			Routers: []Router{
				{
					Index:                         1,
					Enable:                        true,
					Status:                        "Enabled",
					IPv4ForwardingNumberOfEntries: 2,
					IPv4Forwarding: []IPv4ForwardingEntry{
						{
							Index:            2,
							Interface:        "Device.IP.Interface.1",
							Enable:           true,
							Status:           "Enabled",
							StaticRoute:      false,
							DestIPAddress:    "192.168.1.0",
							DestSubnetMask:   "255.255.255.0",
							GatewayIPAddress: "",
							Origin:           "X_MIKROTIK_Connect",
						},
					},
				},
			},
		},
		WiFi: WiFiDevice{
			AccessPointNumberOfEntries: 1,
			AccessPoints: []WiFiAccessPoint{
				{
					Index:                           2,
					AssociatedDeviceNumberOfEntries: 2,
					AssociatedDevices: []WiFiAssociatedDevice{
						{
							Index:               120,
							MACAddress:          "AA:BB:CC:DD:EE:FF",
							AuthenticationState: true,
							SignalStrength:      -69,
							Stats: WiFiAssociatedDeviceStats{
								BytesSent:       12345,
								BytesReceived:   67890,
								PacketsSent:     123,
								PacketsReceived: 456,
							},
						},
						{
							Index:               167,
							MACAddress:          "5A:66:5F:4E:2F:96",
							AuthenticationState: true,
							SignalStrength:      -74,
						},
					},
				},
			},
		},
		IP: IPDevice{
			InterfaceNumberOfEntries: 1,
			Interfaces: []IPInterface{
				{
					Index:                      1,
					Enable:                     true,
					LowerLayers:                "Device.X_MIKROTIK_Interface.Generic.1",
					Status:                     "Up",
					Type:                       "Normal",
					IPv4AddressNumberOfEntries: 1,
					IPv4Addresses: []IPv4AddressEntry{
						{
							Index:          1,
							Enable:         true,
							Status:         "Enabled",
							IPAddress:      "192.168.1.114",
							SubnetMask:     "255.255.255.0",
							AddressingType: "X_MIKROTIK_Dynamic",
						},
					},
				},
			},
			Diagnostics: IPDiagnostics{
				IPPing: IPPingDiagnostics{
					Host: "",
				},
				TraceRoute: TraceRouteDiagnostics{
					Host: "",
				},
			},
		},
	}

	// Create envelope with real XML data including both writable and read-only parameters
	envelope := &soap.GetParameterValuesResponse{
		ParameterList: soap.ParameterList{
			Parameters: []soap.ParameterValueStruct{
				// Writable parameters - these should be processed
				{
					Name: "Device.DeviceInfo.ModelName",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "CHANGED-MODEL", // Different value - should trigger difference
					},
				},
				{
					Name: "Device.Routing.Router.1.IPv4Forwarding.2.DestIPAddress",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "192.168.2.0", // Different value - should trigger difference
					},
				},
				{
					Name: "Device.IP.Interface.1.IPv4Address.1.IPAddress",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "192.168.1.114", // Same value - should not trigger difference
					},
				},
				// Read-only parameters - these should be filtered out
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.Stats.BytesSent",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "123456789", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.Stats.PacketsReceived",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "4023909", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.SignalStrength",
					Value: soap.Value{
						Type:    "xsd:int",
						Content: "-70", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.X_MIKROTIK_Stats.TxFrames",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "2697159", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.DeviceInfo.UpTime",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "123456", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.DeviceInfo.MemoryStatus.Free",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "789012", // This should be ignored (read-only)
					},
				},
				{
					Name: "Device.Hosts.HostNumberOfEntries",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "7", // This should be ignored (read-only NumberOfEntries)
					},
				},
			},
		},
	}

	// Test the comparison
	result, _ := device.CompareEnvelope(envelope)

	// Validate that only writable parameters with differences are included
	if result == nil {
		t.Fatal("Expected differences but got nil result")
	}

	t.Logf("Found %d differences (should only include writable parameters)", len(result.ParameterList.Params))

	// We should only have 2 differences: ModelName and DestIPAddress
	expectedDifferences := 2
	if len(result.ParameterList.Params) != expectedDifferences {
		t.Errorf("Expected %d differences but got %d", expectedDifferences, len(result.ParameterList.Params))
	}

	// Check that read-only parameters are not included
	for _, param := range result.ParameterList.Params {
		t.Logf("Difference: %s = %s", param.Name, param.Value)

		// Verify no read-only parameters made it through
		if !isWritableParameter(param.Name) {
			t.Errorf("Read-only parameter should not be in differences: %s", param.Name)
		}

		// Verify specific expected differences
		switch param.Name {
		case "Device.DeviceInfo.ModelName":
			if param.Value != "CHANGED-MODEL" {
				t.Errorf("Expected ModelName to be 'CHANGED-MODEL', got '%s'", param.Value)
			}
		case "Device.Routing.Router.1.IPv4Forwarding.2.DestIPAddress":
			if param.Value != "192.168.2.0" {
				t.Errorf("Expected DestIPAddress to be '192.168.2.0', got '%s'", param.Value)
			}
		default:
			t.Errorf("Unexpected difference parameter: %s", param.Name)
		}
	}
}

// TestReadOnlyParameterFilter demonstrates that read-only parameters are properly filtered
func TestReadOnlyParameterFilter(t *testing.T) {
	device := &Device{
		DeviceInfo: DeviceInfo{
			Manufacturer: "MikroTik",
			ModelName:    "RB952Ui-5ac2nD",
		},
	}

	// Create envelope with mix of writable and read-only parameters
	envelope := &soap.GetParameterValuesResponse{
		ParameterList: soap.ParameterList{
			Parameters: []soap.ParameterValueStruct{
				// WRITABLE - should be included
				{
					Name: "Device.DeviceInfo.ModelName",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "DIFFERENT-MODEL",
					},
				},
				// READ-ONLY - should be filtered out
				{
					Name: "Device.DeviceInfo.UpTime",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "123456",
					},
				},
				{
					Name: "Device.DeviceInfo.MemoryStatus.Free",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "789012",
					},
				},
				{
					Name: "Device.WiFi.SSID.1.Stats.BytesSent",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "999999",
					},
				},
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.SignalStrength",
					Value: soap.Value{
						Type:    "xsd:int",
						Content: "-50",
					},
				},
				{
					Name: "Device.WiFi.AccessPoint.2.AssociatedDevice.120.X_MIKROTIK_Stats.TxFrames",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "12345",
					},
				},
				{
					Name: "Device.Hosts.HostNumberOfEntries",
					Value: soap.Value{
						Type:    "xsd:unsignedInt",
						Content: "5",
					},
				},
				{
					Name: "Device.InterfaceStack.1.HigherLayer",
					Value: soap.Value{
						Type:    "xsd:string",
						Content: "Device.WiFi.SSID.1",
					},
				},
			},
		},
	}

	result, _ := device.CompareEnvelope(envelope)

	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Should only have 1 difference (the writable ModelName parameter)
	if len(result.ParameterList.Params) != 1 {
		t.Errorf("Expected 1 difference but got %d", len(result.ParameterList.Params))
		for _, param := range result.ParameterList.Params {
			t.Logf("Unexpected difference: %s = %s", param.Name, param.Value)
		}
	}

	// Verify the one difference is the ModelName
	if len(result.ParameterList.Params) > 0 {
		param := result.ParameterList.Params[0]
		if param.Name != "Device.DeviceInfo.ModelName" {
			t.Errorf("Expected ModelName difference but got %s", param.Name)
		}
		if param.Value != "DIFFERENT-MODEL" {
			t.Errorf("Expected 'DIFFERENT-MODEL' but got '%s'", param.Value)
		}
	}
}
