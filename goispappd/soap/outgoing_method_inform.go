package soap

import (
	"encoding/xml"

	"github.com/Niceblueman/goispappd/internal/commands"
	"github.com/Niceblueman/goispappd/internal/exec"
)

func NewRequestEnvelope() *RequestEnvelope {
	return &RequestEnvelope{
		XMLName: xml.Name{Space: "http://schemas.xmlsoap.org/soap/envelope/", Local: "Envelope"},
	}
}

func (e *RequestEnvelope) SetID(id string) {
	e.Header.ID = id
}

func (e *RequestEnvelope) LoadRPCMethods() {
	e.Body.GetRPCMethodsResponse = &GetRPCMethodsResponse{
		MethodList: []string{"GetParameterValues", "SetParameterValues", "Download", "Reboot", "FactoryReset", "AddObject", "DeleteObject", "InformResponse", "RequestXCommand", "TransferCompleteResponse", "GetParameterNames"},
	}
}

func (e *RequestEnvelope) LoadParameterNames(path string) {
	listavaiable := map[string]bool{
		"Device.InterfaceStackNumberOfEntries": false,
		"Device.DeviceSummary":                 false,
		"Device.Routing.":                      false,
		"Device.Firewall.":                     false,
		"Device.InterfaceStack.":               false,
		"Device.RootDataModelVersion":          false,
		"Device.DHCPv4.":                       false,
		"Device.X_ISPAPP_Interface.":           false,
		"Device.Hosts.":                        false,
		"Device.DeviceInfo.":                   false,
		"Device.ManagementServer.":             false,
		"Device.Cellular.":                     false,
		"Device.WiFi.":                         false,
		"Device.IP.":                           false,
		"Device.Ethernet.":                     false,
		"Device.PPP.":                          false,
		"Device.DNS.":                          false,
	}
	e.Body.GetParameterNamesResponse = &GetParameterNamesResponse{
		ParameterPath: path,
		NextLevel:     1,
		ParameterList: make([]struct {
			XMLName  xml.Name "xml:\"ParameterInfoStruct\""
			Name     string   "xml:\"Name\""
			Writable bool     "xml:\"Writable\""
		}, 0),
	}
	for key := range listavaiable {
		e.Body.GetParameterNamesResponse.ParameterList = append(e.Body.GetParameterNamesResponse.ParameterList, struct {
			XMLName  xml.Name "xml:\"ParameterInfoStruct\""
			Name     string   "xml:\"Name\""
			Writable bool     "xml:\"Writable\""
		}{
			Name:     key,
			Writable: false,
		})
	}
}

func (e *RequestEnvelope) LoadInformResponse() {
	e.Body.Inform = &Inform{}
	executer := exec.NewExecutor(exec.ExecConfig{})
	e.Body.Inform.ParameterList = ParameterList{}
	if getter := commands.InformCommands["Device.DeviceInfo.ManufacturerOUI"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.DeviceID.OUI = string(result.Raw)
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManufacturerOUI",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		} else {
			e.Body.Inform.DeviceID.OUI = "Unknown"
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.SerialNumber"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.SetID(string(result.Raw))
			e.Body.Inform.DeviceID.SerialNumber = string(result.Raw)
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.SerialNumber",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		} else {
			e.Body.Inform.DeviceID.SerialNumber = "Unknown"
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.Manufacturer"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.DeviceID.Manufacturer = string(result.Raw)
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.Manufacturer",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		} else {
			e.Body.Inform.DeviceID.Manufacturer = "Unknown"
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ModelName"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ModelName",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		} else {
			e.Body.Inform.DeviceID.ProductClass = "Unknown"
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ProductClass"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.DeviceID.ProductClass = string(result.Raw)
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ProductClass",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		} else {
			e.Body.Inform.DeviceID.ProductClass = "Unknown"
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.HardwareVersion"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.HardwareVersion",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.SoftwareVersion"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.SoftwareVersion",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ProvisioningCode"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ProvisioningCode",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ManagementServer.PeriodicInformEnable"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManagementServer.PeriodicInformEnable",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ManagementServer.PeriodicInformInterval"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManagementServer.PeriodicInformInterval",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ManagementServer.URL"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManagementServer.URL",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ManagementServer.Username"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManagementServer.Username",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ManagementServer.Password"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ManagementServer.Password",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.UpTime"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.UpTime",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFileNumberOfEntries"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFileNumberOfEntries",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.MemoryStatus.Total"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.MemoryStatus.Total",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.MemoryStatus.Free"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.MemoryStatus.Free",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.ProcessStatus.CPUUsage"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.ProcessStatus.CPUUsage",
				Value: Value{
					Type:    "unsignedInt",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.1.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.1.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.1.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.1.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.1.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.1.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.2.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.2.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.2.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.2.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.2.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.2.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.3.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.3.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.3.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.3.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.3.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.3.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.4.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.4.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.4.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.4.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.4.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.4.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.5.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.5.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.5.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.5.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.5.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.5.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.6.Name"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.6.Name",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.6.Description"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.6.Description",
				Value: Value{
					Type:    "string",
					Content: string(result.Raw),
				},
			})
		}
	}
	if getter := commands.InformCommands["Device.DeviceInfo.VendorConfigFile.6.UseForBackupRestore"]; getter != nil {
		if result, err := getter(executer, nil); err == nil && result.Success {
			e.Body.Inform.ParameterList.Parameters = append(e.Body.Inform.ParameterList.Parameters, ParameterValueStruct{
				Name: "Device.DeviceInfo.VendorConfigFile.6.UseForBackupRestore",
				Value: Value{
					Type:    "boolean",
					Content: string(result.Raw),
				},
			})
		}
	}
}
