package uci_test

import (
	"testing"

	"github.com/Niceblueman/goispappd/internal/uci"
)

func TestUCI(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "TestUCIGet"},
		{name: "TestUCISet"},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.name {
			case "TestUCIGet":
				uci, err := uci.NewUCI()
				if err != nil {
					t.Fatalf("Failed to create UCI context: %v", err)
				}
				defer uci.Free()
				value, err := uci.Get("ispappd", "@device[0]", "manufacturer")
				if err != nil {
					t.Errorf("Failed to get UCI value: %v", err)
				} else {
					t.Logf("UCI value: %s", value)
				}
			case "TestUCISet":
				uci, err := uci.NewUCI()
				if err != nil {
					t.Fatalf("Failed to create UCI context: %v", err)
				}
				defer uci.Free()
				err = uci.Set("ispappd", "@device[0]", "manufacturer", "TestManufacturer")
				if err != nil {
					t.Errorf("Failed to set UCI value: %v", err)
				}
				value, err := uci.Get("ispappd", "@device[0]", "manufacturer")
				if err != nil {
					t.Errorf("Failed to get UCI value after set: %v", err)
				} else if value != "TestManufacturer" {
					t.Errorf("UCI value mismatch: expected 'TestManufacturer', got '%s'", value)
				} else {
					t.Logf("UCI value after set: %s", value)
				}
			default:
				t.Errorf("Unknown test case: %s", tt.name)
			}
		})
	}
}
