package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Configuration holds the client configuration
type Configuration struct {
	ACSURL           string        `yaml:"acs_url"`
	Username         string        `yaml:"username"`
	Password         string        `yaml:"password"`
	SerialNumber     string        `yaml:"serial_number"`
	PeriodicInterval time.Duration `yaml:"periodic_interval"`
	ProvisioningCode string        `yaml:"provisioning_code"`
}

// LoadConfig loads configuration from a YAML file /etc/cwmp/config.yaml
func LoadConfig() (*Configuration, error) {
	// Check if the cwmp directory exists
	if _, err := os.Stat("/etc/cwmp"); os.IsNotExist(err) {
		// Create the directory if it does not exist
		if err := os.MkdirAll("/etc/cwmp", 0755); err != nil {
			return nil, err
		}
	}
	// check if the config file exists
	if _, err := os.Stat("/etc/cwmp/config.yaml"); os.IsNotExist(err) {
		// Create a default config file if it does not exist
		defaultConfig := &Configuration{
			ACSURL:           "https://local.longshot-router.com/tr069",
			Username:         "",
			Password:         "",
			SerialNumber:     "1234567890",
			PeriodicInterval: 30 * time.Second, // Default periodic interval
			ProvisioningCode: "",
		}
		data, err := yaml.Marshal(defaultConfig)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile("/etc/cwmp/config.yaml", data, 0644); err != nil {
			return nil, err
		}
	}
	data, err := os.ReadFile("/etc/cwmp/config.yaml")
	if err != nil {
		return nil, err
	}
	cfg := &Configuration{
		PeriodicInterval: 30 * time.Second, // Default
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
