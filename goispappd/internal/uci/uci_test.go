package uci_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Niceblueman/goispappd/internal/uci"
)

func TestUCI(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "TestUCISet"},
		{name: "TestUCIGet"},
		{name: "TestLoadConfigWithPackage"},
		{name: "TestNewFormat"},
		{name: "TestSetCreatesSection"},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.name {
			case "TestUCIGet":
				t.Run("UCIGet", func(t *testing.T) {
					// Call the function to test
					_package := "Device"
					uci, err := uci.LoadConfig("/etc/config/tr069", &_package)
					if err != nil {
						t.Errorf("Failed to create UCI context: %v", err)
						return
					}
					ssid, err := uci.Get("WIFI", "keypath")
					if err != nil {
						t.Logf("Key not found (expected in empty config): %v", err)
					} else {
						var _ssid string
						switch ssid := ssid.(type) {
						case string:
							_ssid = ssid
						default:
							_ssid = ""
						}
						t.Logf("Value: %s", _ssid)
					}
				})
			case "TestUCISet":
				t.Run("UCISet", func(t *testing.T) {
					// Call the function to test
					_package := "Device"
					uci, err := uci.LoadConfig("/etc/config/tr069", &_package)
					if err != nil {
						t.Errorf("Failed to create UCI context: %v", err)
						return
					}

					t.Logf("Initial state - Package: '%s', Sections: %d", uci.Package, len(uci.Sections))

					err = uci.Set("WIFI", "keypath", "value", false)
					if err != nil {
						t.Errorf("Failed to set UCI value: %v", err)
					}

					t.Logf("After Set - Package: '%s', Sections: %d", uci.Package, len(uci.Sections))
					for i, sec := range uci.Sections {
						t.Logf("Section %d: Type='%s', Name='%s', Options=%d", i, sec.SectionType, sec.Name, len(sec.Options))
					}

					if err := uci.Save(); err != nil {
						t.Errorf("Failed to save UCI config: %v", err)
					}

					// Read back what was saved
					content, err := os.ReadFile("/etc/config/tr069")
					if err != nil {
						t.Logf("Could not read saved file: %v", err)
					} else {
						t.Logf("Saved content (%d bytes): '%s'", len(content), string(content))
					}
				})
			case "TestLoadConfigWithPackage":
				t.Run("LoadConfigWithPackage", func(t *testing.T) {
					// Create a temporary file with the new format
					tmpDir := t.TempDir()
					testFile := filepath.Join(tmpDir, "test.conf")

					testConfig := `package Device
config WIFI
	option keypath value
	list servers server1
	list servers server2
`
					err := os.WriteFile(testFile, []byte(testConfig), 0644)
					if err != nil {
						t.Fatalf("Failed to create test file: %v", err)
					}

					// Test loading without specifying a section
					cfg, err := uci.LoadConfig(testFile, nil)
					if err != nil {
						t.Fatalf("Failed to load config: %v", err)
					}

					if cfg.Package != "Device" {
						t.Errorf("Expected package 'Device', got '%s'", cfg.Package)
					}

					if len(cfg.Sections) != 1 {
						t.Errorf("Expected 1 section, got %d", len(cfg.Sections))
					}

					if cfg.Sections[0].SectionType != "WIFI" {
						t.Errorf("Expected section type 'WIFI', got '%s'", cfg.Sections[0].SectionType)
					}

					// Test getting values
					val, err := cfg.Get("WIFI", "keypath")
					if err != nil {
						t.Errorf("Failed to get keypath: %v", err)
					}
					if val != "value" {
						t.Errorf("Expected 'value', got '%v'", val)
					}

					// Test getting list values
					listVal, err := cfg.Get("WIFI", "servers")
					if err != nil {
						t.Errorf("Failed to get servers list: %v", err)
					}
					servers, ok := listVal.([]string)
					if !ok {
						t.Errorf("Expected []string, got %T", listVal)
					}
					if len(servers) != 2 || servers[0] != "server1" || servers[1] != "server2" {
						t.Errorf("Expected [server1 server2], got %v", servers)
					}
				})
			case "TestNewFormat":
				t.Run("NewFormat", func(t *testing.T) {
					// Test with the new format: package Device, config WIFI, option keypath value
					tmpDir := t.TempDir()
					testFile := filepath.Join(tmpDir, "device.conf")

					// Create a config with the specified section
					sectionType := "WIFI"
					cfg, err := uci.LoadConfig(testFile, &sectionType)
					if err != nil {
						t.Fatalf("Failed to load config: %v", err)
					}

					// Set a value
					err = cfg.Set("WIFI", "keypath", "value", false)
					if err != nil {
						t.Errorf("Failed to set value: %v", err)
					}

					// Save the config
					err = cfg.Save()
					if err != nil {
						t.Errorf("Failed to save config: %v", err)
					}

					// Read back and verify format
					content, err := os.ReadFile(testFile)
					if err != nil {
						t.Errorf("Failed to read saved file: %v", err)
					}

					expected := "package WIFI\nconfig WIFI \"WIFI\"\n\toption keypath value\n\n"
					if string(content) != expected {
						t.Errorf("Expected:\n%s\nGot:\n%s", expected, string(content))
					}
				})
			case "TestSetCreatesSection":
				t.Run("SetCreatesSection", func(t *testing.T) {
					// Create a temporary file for testing
					tmpDir := t.TempDir()
					testFile := filepath.Join(tmpDir, "test.conf")

					// Call the function to test with a temporary file
					_package := "Device"
					uci, err := uci.LoadConfig(testFile, &_package)
					if err != nil {
						t.Errorf("Failed to create UCI context: %v", err)
						return
					}

					// Debug: check what sections exist
					t.Logf("Number of sections: %d", len(uci.Sections))
					for i, sec := range uci.Sections {
						t.Logf("Section %d: Type=%s, Name=%s", i, sec.SectionType, sec.Name)
					}

					// Ensure the section does not exist before the test
					_, err = uci.Get("WIFI", "keypath")
					if err == nil {
						t.Errorf("Expected section 'WIFI' to be absent")
						return
					}

					// Set a value, which should create the section
					err = uci.Set("WIFI", "keypath", "value", false)
					if err != nil {
						t.Errorf("Failed to set UCI value: %v", err)
						return
					}

					// Verify the section was created
					val, err := uci.Get("WIFI", "keypath")
					if err != nil {
						t.Errorf("Failed to get keypath: %v", err)
						return
					}
					if val != "value" {
						t.Errorf("Expected 'value', got '%v'", val)
					}
				})
			default:
				t.Errorf("Unknown test case: %s", tt.name)
			}
		})
	}
}
