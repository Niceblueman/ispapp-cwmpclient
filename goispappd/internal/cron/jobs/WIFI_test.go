package jobs_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Niceblueman/goispappd/internal/cron/jobs"
	"github.com/Niceblueman/goispappd/internal/exec"
	"github.com/Niceblueman/goispappd/internal/uci"
)

// Test configuration
const (
	TestHost       = "192.168.1.170:22" // Added SSH port
	Username       = "root"
	PrivateKeyPath = "/Users/kimo/.ssh/id_ed25519"
	TestTimeout    = 30 // Increased timeout for debugging
)

// TestConfig holds common test configuration
type TestConfig struct {
	Host     string
	Executor *exec.Executor
	Logger   *log.Logger
}

// setupTest creates a common test configuration with enhanced logging
func setupTest(testName string) *TestConfig {
	// Create custom logger for this test
	logger := log.New(os.Stdout, fmt.Sprintf("[%s] ", testName), log.LstdFlags|log.Lshortfile)

	host := TestHost
	executor := exec.NewExecutor(exec.ExecConfig{
		Credentials: &exec.SSHCredentials{
			Username:       Username,
			PrivateKeyPath: PrivateKeyPath,
			Host:           &host,
		},
	})

	return &TestConfig{
		Host:     host,
		Executor: executor,
		Logger:   logger,
	}
}

// debugExecutorCommands logs all commands being executed
func debugExecutorCommands(tc *TestConfig, description string) {
	tc.Logger.Printf("=== %s ===", description)
	tc.Logger.Printf("Host: %s", tc.Host)
	tc.Logger.Printf("Timeout: %ds", TestTimeout)
}

// validateUCIResults reads back UCI values to verify they were set correctly using UCI module
func validateUCIResults(tc *TestConfig, section string) {
	tc.Logger.Printf("--- Validating UCI Results for %s ---", section)

	// Use UCI module to read remote configuration
	_package := "Device"
	uciConfig, err := uci.LoadConfig("/etc/config/tr069", &_package)
	if err != nil {
		tc.Logger.Printf("ERROR loading UCI config via module: %v", err)
		// Fallback to direct remote UCI command
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()
		result, err := tc.Executor.Execute(ctx, "uci", "show", "tr069")
		if err != nil {
			tc.Logger.Printf("ERROR reading remote UCI config: %v", err)
			return
		}

		output, ok := result.Stdout.(string)
		if !ok {
			tc.Logger.Print("ERROR: unexpected UCI output type")
			return
		}

		tc.Logger.Print("Remote UCI tr069 configuration:")
		lines := strings.Split(output, "\n")
		relevantLines := 0
		for _, line := range lines {
			if strings.Contains(line, section) {
				tc.Logger.Printf("  %s", line)
				relevantLines++
			}
		}

		if relevantLines == 0 {
			tc.Logger.Printf("WARNING: No %s entries found in UCI config", section)
		} else {
			tc.Logger.Printf("Found %d %s related entries", relevantLines, section)
		}
		return
	}

	tc.Logger.Print("SUCCESS: UCI config loaded via module")
	// Note: The UCI module handles remote operations internally
	_ = uciConfig // Use the variable to avoid unused error
}

// debugWirelessConfig shows current wireless configuration for debugging
func debugWirelessConfig(tc *TestConfig) {
	tc.Logger.Print("--- Current Wireless Configuration ---")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Show remote wireless config via SSH
	result, err := tc.Executor.Execute(ctx, "uci", "show", "wireless")
	if err != nil {
		tc.Logger.Printf("ERROR reading remote wireless config: %v", err)
		return
	}

	output, ok := result.Stdout.(string)
	if !ok {
		tc.Logger.Print("ERROR: unexpected wireless output type")
		return
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i < 20 { // Limit output for readability
			tc.Logger.Printf("  %s", line)
		} else {
			tc.Logger.Printf("  ... (%d more lines)", len(lines)-20)
			break
		}
	}
}

// debugNetworkInterfaces shows available network interfaces on remote device
func debugNetworkInterfaces(tc *TestConfig) {
	tc.Logger.Print("--- Available Network Interfaces (Remote) ---")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Show remote interfaces via SSH
	result, err := tc.Executor.Execute(ctx, "ip", "link", "show")
	if err != nil {
		tc.Logger.Printf("ERROR reading remote interfaces: %v", err)
		return
	}

	output, ok := result.Stdout.(string)
	if !ok {
		tc.Logger.Print("ERROR: unexpected interface output type")
		return
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "wlan") || strings.Contains(line, "wifi") || strings.Contains(line, "ath") {
			tc.Logger.Printf("  WiFi: %s", line)
		}
	}
}

// debugWiFiScanCapabilities tests WiFi scanning capabilities on remote device
func debugWiFiScanCapabilities(tc *TestConfig) {
	tc.Logger.Print("--- WiFi Scan Capabilities (Remote) ---")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Check remote iw capabilities via SSH
	result, err := tc.Executor.Execute(ctx, "iw", "dev")
	if err != nil {
		tc.Logger.Printf("ERROR: remote iw dev failed: %v", err)
		return
	}

	output, ok := result.Stdout.(string)
	if !ok {
		tc.Logger.Print("ERROR: unexpected iw dev output type")
		return
	}

	tc.Logger.Print("Available WiFi devices on remote:")
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Interface") || strings.Contains(line, "phy") {
			tc.Logger.Printf("  %s", line)
		}
	}
}

// measureExecutionTime measures and logs function execution time
func measureExecutionTime(tc *TestConfig, functionName string, fn func()) {
	tc.Logger.Printf("Starting %s...", functionName)
	start := time.Now()
	fn()
	duration := time.Since(start)
	tc.Logger.Printf("%s completed in %v", functionName, duration)
}

func TestSSIDCollectCmd(t *testing.T) {
	tc := setupTest("TestSSIDCollectCmd")

	debugExecutorCommands(tc, "SSID Collection Test")
	debugWirelessConfig(tc)
	debugNetworkInterfaces(tc)

	// Test the actual function with timing
	measureExecutionTime(tc, "SSIDCollectCmd", func() {
		errPtr := jobs.SSIDCollectCmd(tc.Executor)
		if errPtr != nil {
			t.Errorf("SSIDCollectCmd failed: %v", *errPtr)
			return
		}
		tc.Logger.Print("SSIDCollectCmd executed successfully")
	})

	// Show detailed results by checking what was written to UCI
	tc.Logger.Print("--- Analyzing SSID Collection Results ---")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Check the UCI tr069 config to see what was written
	result, err := tc.Executor.Execute(ctx, "uci", "show", "tr069")
	if err != nil {
		tc.Logger.Printf("ERROR: Failed to read tr069 config: %v", err)
	} else if output, ok := result.Stdout.(string); ok {
		tc.Logger.Print("TR069 UCI Configuration after SSID collection:")
		lines := strings.Split(output, "\n")
		ssidCount := 0
		for _, line := range lines {
			if strings.Contains(line, "SSID") {
				tc.Logger.Printf("  %s", line)
				if strings.Contains(line, "SSIDNumberOfEntries") {
					// Extract the number
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						count := strings.Trim(parts[1], "'\"")
						tc.Logger.Printf("*** FOUND SSIDNumberOfEntries = %s ***", count)
					}
				}
				ssidCount++
			}
		}
		tc.Logger.Printf("Total SSID-related UCI entries: %d", ssidCount)

		// Also check if there are any WiFi interface entries
		interfaceCount := 0
		for _, line := range lines {
			if strings.Contains(line, "WiFi.") && strings.Contains(line, "SSID.") {
				tc.Logger.Printf("  WiFi Interface: %s", line)
				interfaceCount++
			}
		}
		tc.Logger.Printf("Total WiFi interface entries: %d", interfaceCount)
	}

	// Validate results
	validateUCIResults(tc, "SSID")

	tc.Logger.Print("=== SSID Collection Test Completed ===")
}

func TestRadiosCollectCmd(t *testing.T) {
	tc := setupTest("TestRadiosCollectCmd")

	debugExecutorCommands(tc, "Radio Collection Test")
	debugWirelessConfig(tc)

	// Test radio capabilities first
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	result, err := tc.Executor.Execute(ctx, "iw", "phy")
	if err != nil {
		tc.Logger.Printf("WARNING: iw phy command failed: %v", err)
	} else if output, ok := result.Stdout.(string); ok {
		tc.Logger.Printf("Available PHY devices: %s", strings.TrimSpace(output))
	}

	// Test the actual function with timing
	measureExecutionTime(tc, "RadiosCollectCmd", func() {
		jobs.RadiosCollectCmd(tc.Executor)
	})

	// Validate results
	validateUCIResults(tc, "Radio")

	tc.Logger.Print("=== Radio Collection Test Completed ===")
}

func TestAccessPointCollectCmd(t *testing.T) {
	tc := setupTest("TestAccessPointCollectCmd")

	debugExecutorCommands(tc, "Access Point Collection Test")
	debugWirelessConfig(tc)
	debugNetworkInterfaces(tc)

	// Test wlanconfig capabilities
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Check for wlanconfig command availability
	result, err := tc.Executor.Execute(ctx, "which", "wlanconfig")
	if err != nil {
		tc.Logger.Printf("WARNING: wlanconfig not found: %v", err)
	} else {
		tc.Logger.Printf("wlanconfig available at: %v", result.Stdout)
	}

	// Test the actual function with timing
	measureExecutionTime(tc, "AccessPointCollectCmd", func() {
		jobs.AccessPointCollectCmd(tc.Executor)
	})

	// Validate results
	validateUCIResults(tc, "AccessPoint")

	tc.Logger.Print("=== Access Point Collection Test Completed ===")
}

func TestNeighboringWiFiCollectCmd(t *testing.T) {
	tc := setupTest("TestNeighboringWiFiCollectCmd")

	debugExecutorCommands(tc, "Neighboring WiFi Collection Test")
	debugWiFiScanCapabilities(tc)

	// Test the actual function with timing
	measureExecutionTime(tc, "NeighboringWiFiCollectCmd", func() {
		jobs.NeighboringWiFiCollectCmd(tc.Executor)
	})

	// Validate results
	validateUCIResults(tc, "NeighboringWiFi")

	tc.Logger.Print("=== Neighboring WiFi Collection Test Completed ===")
}

// TestWiFiHelperFunctions tests individual helper functions for debugging
func TestWiFiHelperFunctions(t *testing.T) {
	tc := setupTest("TestWiFiHelperFunctions")

	debugExecutorCommands(tc, "WiFi Helper Functions Test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Test UCI configuration loading (this uses the UCI module, not local commands)
	tc.Logger.Print("--- Testing UCI Configuration Access (via Module) ---")
	_package := "Device"
	uciConfig, err := uci.LoadConfig("/etc/config/tr069", &_package)
	if err != nil {
		tc.Logger.Printf("ERROR: Failed to load UCI config via module: %v", err)
	} else {
		tc.Logger.Print("SUCCESS: UCI config loaded successfully via module")
		_ = uciConfig // Use the variable to avoid unused error
	}

	// Test basic remote wireless commands (these execute on remote device via SSH)
	tc.Logger.Print("--- Testing Remote Wireless Commands ---")

	commands := [][]string{
		{"uci", "show", "wireless"},
		{"iw", "dev"},
		{"ip", "link", "show"},
		{"cat", "/proc/net/dev"},
	}

	for _, cmd := range commands {
		tc.Logger.Printf("Testing remote command: %v", cmd)
		result, err := tc.Executor.Execute(ctx, cmd[0], cmd[1:]...)
		if err != nil {
			tc.Logger.Printf("  ERROR: %v", err)
		} else {
			if output, ok := result.Stdout.(string); ok {
				lines := strings.Split(output, "\n")
				tc.Logger.Printf("  SUCCESS: %d lines of output from remote", len(lines))
				if len(lines) > 0 && len(lines[0]) > 100 {
					tc.Logger.Printf("  Sample: %s...", lines[0][:100])
				} else if len(lines) > 0 {
					tc.Logger.Printf("  Sample: %s", lines[0])
				}
			}
		}
	}

	tc.Logger.Print("=== Helper Functions Test Completed ===")
}

// TestWiFiDataStructures tests the data structures and parsing
func TestWiFiDataStructures(t *testing.T) {
	tc := setupTest("TestWiFiDataStructures")

	tc.Logger.Print("--- Testing WiFi Data Structure Parsing ---")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Test parsing remote wireless configuration
	result, err := tc.Executor.Execute(ctx, "uci", "show", "wireless")
	if err != nil {
		tc.Logger.Printf("ERROR: Failed to get remote wireless config: %v", err)
		return
	}
	output, ok := result.Stdout.(string)
	if !ok {
		tc.Logger.Print("ERROR: Unexpected output type")
		return
	}

	lines := strings.Split(output, "\n")
	interfaceCount := 0
	deviceCount := 0

	for _, line := range lines {
		if strings.Contains(line, "wifi-iface") {
			interfaceCount++
		}
		if strings.Contains(line, "wifi-device") {
			deviceCount++
		}
	}

	tc.Logger.Printf("Found %d wifi interfaces and %d wifi devices in remote config", interfaceCount, deviceCount)

	// Test remote network statistics parsing
	result, err = tc.Executor.Execute(ctx, "cat", "/proc/net/dev")
	if err != nil {
		tc.Logger.Printf("WARNING: Failed to read remote /proc/net/dev: %v", err)
	} else if output, ok := result.Stdout.(string); ok {
		lines := strings.Split(output, "\n")
		wifiInterfaces := 0
		for _, line := range lines {
			if strings.Contains(line, "wlan") || strings.Contains(line, "ath") || strings.Contains(line, "wifi") {
				wifiInterfaces++
				tc.Logger.Printf("  Remote WiFi interface stats: %s", strings.TrimSpace(line))
			}
		}
		tc.Logger.Printf("Found %d WiFi interfaces with statistics on remote device", wifiInterfaces)
	}

	tc.Logger.Print("=== Data Structures Test Completed ===")
}

// TestWiFiComprehensive runs all tests in sequence for comprehensive debugging
func TestWiFiComprehensive(t *testing.T) {
	tc := setupTest("TestWiFiComprehensive")
	tc.Logger.Print("=== COMPREHENSIVE WIFI DEBUGGING TEST ===")

	// Run all tests in sequence
	tests := []struct {
		name string
		fn   func()
	}{
		{"Helper Functions", func() { TestWiFiHelperFunctions(t) }},
		{"Data Structures", func() { TestWiFiDataStructures(t) }},
		{"SSID Collection", func() { TestSSIDCollectCmd(t) }},
		{"Radio Collection", func() { TestRadiosCollectCmd(t) }},
		{"Access Point Collection", func() { TestAccessPointCollectCmd(t) }},
		{"Neighboring WiFi Collection", func() { TestNeighboringWiFiCollectCmd(t) }},
	}

	for _, test := range tests {
		tc.Logger.Printf("\n" + strings.Repeat("=", 60))
		tc.Logger.Printf("Running: %s", test.name)
		tc.Logger.Print(strings.Repeat("=", 60))

		measureExecutionTime(tc, test.name, test.fn)

		// Small delay between tests
		time.Sleep(1 * time.Second)
	}

	tc.Logger.Printf("\n" + strings.Repeat("=", 60))
	tc.Logger.Print("=== COMPREHENSIVE TEST COMPLETED ===")
	tc.Logger.Print(strings.Repeat("=", 60))
}
