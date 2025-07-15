package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// OutputType defines the type of command output
type OutputType string

const (
	TypeString OutputType = "string"
	TypeJSON   OutputType = "json"
	TypeXML    OutputType = "xml"
	TypeBinary OutputType = "binary"
)

// CommandResult holds the result of a command execution
type CommandResult struct {
	Type    OutputType
	Stdout  interface{}
	Stderr  string
	Raw     []byte
	Success bool
}

// ExecConfig holds configuration for command execution
type ExecConfig struct {
	Timeout     time.Duration
	Credentials *SSHCredentials
}

// SSHCredentials holds SSH authentication details
type SSHCredentials struct {
	Username       string
	Password       string
	PrivateKey     []byte
	PrivateKeyPath string // NEW: path to private key file
}

// Executor manages command execution
type Executor struct {
	config ExecConfig
}

// NewExecutor creates a new Executor instance
func NewExecutor(config ExecConfig) *Executor {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &Executor{config: config}
}

// Execute runs a local command
func (e *Executor) Execute(ctx context.Context, command string, args ...string) (*CommandResult, error) {
	if command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &CommandResult{
		Raw:     stdout.Bytes(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timed out after %v", e.config.Timeout)
	}

	if err != nil {
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return e.parseOutput(result)
}

// SSHExecute runs a command over SSH
func (e *Executor) SSHExecute(ctx context.Context, host, command string) (*CommandResult, error) {
	if host == "" || command == "" {
		return nil, fmt.Errorf("host and command cannot be empty")
	}

	config, err := e.buildSSHConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	ctx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(command)
	}()

	select {
	case err = <-errChan:
		if err != nil {
			return &CommandResult{
				Stderr:  stderr.String(),
				Success: false,
			}, fmt.Errorf("SSH command execution failed: %w", err)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("SSH command timed out after %v", e.config.Timeout)
	}

	result := &CommandResult{
		Raw:     stdout.Bytes(),
		Stderr:  stderr.String(),
		Success: true,
	}

	return e.parseOutput(result)
}

// buildSSHConfig creates SSH client configuration
func (e *Executor) buildSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		Timeout:         e.config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if e.config.Credentials == nil {
		config.User = "root"
		config.Auth = []ssh.AuthMethod{ssh.Password("")}
		return config, nil
	}

	config.User = e.config.Credentials.Username
	var authMethods []ssh.AuthMethod

	if e.config.Credentials.Password != "" {
		authMethods = append(authMethods, ssh.Password(e.config.Credentials.Password))
	}

	// NEW: If PrivateKeyPath is set, read the file
	if e.config.Credentials.PrivateKeyPath != "" {
		keyData, err := os.ReadFile(e.config.Credentials.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if len(e.config.Credentials.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(e.config.Credentials.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods provided")
	}

	config.Auth = authMethods
	return config, nil
}

// parseOutput determines the output type and parses accordingly
func (e *Executor) parseOutput(result *CommandResult) (*CommandResult, error) {
	if len(result.Raw) == 0 {
		result.Type = TypeString
		result.Stdout = ""
		return result, nil
	}

	// Try JSON first
	var jsonData interface{}
	if json.Unmarshal(result.Raw, &jsonData) == nil {
		result.Type = TypeJSON
		result.Stdout = jsonData
		return result, nil
	}

	// Try XML
	var xmlData interface{}
	if xml.Unmarshal(result.Raw, &xmlData) == nil {
		result.Type = TypeXML
		result.Stdout = xmlData
		return result, nil
	}

	// Check if binary (non-printable characters)
	isBinary := false
	for _, b := range result.Raw {
		if b < 32 && b != 9 && b != 10 && b != 13 {
			isBinary = true
			break
		}
	}

	if isBinary {
		result.Type = TypeBinary
		result.Stdout = result.Raw
	} else {
		result.Type = TypeString
		result.Stdout = strings.TrimSpace(string(result.Raw))
	}

	return result, nil
}
