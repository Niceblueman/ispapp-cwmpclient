package ubus_test

import (
	"testing"

	"github.com/Niceblueman/goispappd/internal/ubus"
)

func TestUbus(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "TestUbusConnect"},
		{name: "TestUbusList"},
		{name: "TestHelloMethod"},
		{name: "TestWatchMethod"},
		{name: "TestCountMethod"},
	}
	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.name {
			case "TestUbusConnect":
				ubus, err := ubus.NewUbus()
				if err != nil {
					t.Fatalf("Failed to create Ubus context: %v", err)
				}
				defer ubus.Free()
				if ubus == nil {
					t.Error("Ubus context is nil after creation")
				} else {
					t.Log("Ubus context created successfully")
				}
			case "TestUbusList":
				ubus, err := ubus.NewUbus()
				if err != nil {
					t.Fatalf("Failed to create Ubus context: %v", err)
				}
				defer ubus.Free()
				result, err := ubus.List()
				if err != nil {
					t.Errorf("Ubus list call failed: %v", err)
				} else {
					t.Logf("Ubus list call result: %+v", result)
				}
			case "TestHelloMethod":
				ubus, err := ubus.NewUbus()
				if err != nil {
					t.Fatalf("Failed to create Ubus context: %v", err)
				}
				defer ubus.Free()

				// Test hello method with id (Integer) and msg (String)
				args := map[string]interface{}{
					"id":  123,
					"msg": "Hello World",
				}
				result, err := ubus.Call("test", "hello", args)
				if err != nil {
					t.Logf("Ubus hello call failed (this may be expected): %v", err)
				} else {
					t.Logf("Ubus hello call result: %+v", result)
				}
			case "TestWatchMethod":
				ubus, err := ubus.NewUbus()
				if err != nil {
					t.Fatalf("Failed to create Ubus context: %v", err)
				}
				defer ubus.Free()

				// Test watch method with id (Integer) and counter (Integer)
				args := map[string]interface{}{
					"id":      456,
					"counter": 100,
				}
				result, err := ubus.Call("test", "watch", args)
				if err != nil {
					t.Logf("Ubus watch call failed (this may be expected): %v", err)
				} else {
					t.Logf("Ubus watch call result: %+v", result)
				}
			case "TestCountMethod":
				ubus, err := ubus.NewUbus()
				if err != nil {
					t.Fatalf("Failed to create Ubus context: %v", err)
				}
				defer ubus.Free()

				// Test count method with to (Integer) and string (String)
				args := map[string]interface{}{
					"to":     789,
					"string": "Count Test",
				}
				result, err := ubus.Call("test", "count", args)
				if err != nil {
					t.Logf("Ubus count call failed (this may be expected): %v", err)
				} else {
					t.Logf("Ubus count call result: %+v", result)
				}
			default:
				t.Errorf("Unknown test case: %s", tt.name)
			}
		})
	}
}
