package soap_test

import (
	"testing"

	"github.com/Niceblueman/goispappd/internal/commands"
	"github.com/Niceblueman/goispappd/internal/exec"
	"github.com/Niceblueman/goispappd/soap"
)

func TestLoaders(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "TestLoadInformResponse"},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.name {
			case "TestLoadInformResponse":
				t.Run("LoadInformResponse", func(t *testing.T) {
					// Call the function to test
					executer := exec.NewExecutor(exec.ExecConfig{
						Credentials: &exec.SSHCredentials{
							Username:       "root",
							PrivateKeyPath: "/Users/kimo/.ssh/id_ed25519",
						},
					})
					informResponse := soap.NewRequestEnvelope()
					informResponse.Body.Inform = &soap.Inform{}
					if getter := commands.InformCommands["Device.DeviceInfo.ManufacturerOUI"]; getter != nil {
						ssh_host := "192.168.1.170:22"
						if result, err := getter(executer, &ssh_host); err == nil && result.Success {
							informResponse.Body.Inform.DeviceID.OUI = string(result.Raw)
							t.Logf("Successfully retrieved ManufacturerOUI: %s", informResponse.Body.Inform.DeviceID.OUI)
						} else {
							t.Errorf("Failed to get ManufacturerOUI: %v", err)
						}
					} else {
						t.Error("Failed to get function")
					}
				})
			default:
				t.Errorf("Unknown test case: %s", tt.name)
			}
		})
	}
}
