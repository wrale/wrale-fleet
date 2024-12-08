// Package sysadmin provides secure command execution for fleet management demos
package sysadmin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unicode"
)

// DeviceCommand represents a validated device management command
type Command interface {
	// Execute runs the command with the given context
	Execute(context.Context) error
}

// CommandExecutor manages device management commands
type CommandExecutor struct {
	execPath string   // Validated path to wfcentral binary
	baseArgs []string // Immutable base arguments
}

// NewCommandExecutor creates a command executor with a validated executable path
func NewCommandExecutor(wfcentralPath string) (*CommandExecutor, error) {
	// Validate and resolve absolute path
	absPath, err := filepath.Abs(wfcentralPath)
	if err != nil {
		return nil, fmt.Errorf("invalid wfcentral path: %w", err)
	}

	// Verify file exists and is executable
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("wfcentral not found: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("wfcentral is not executable")
	}

	return &CommandExecutor{
		execPath: absPath,
		baseArgs: []string{"device"}, // Immutable base command
	}, nil
}

// validateDeviceName ensures device names contain only allowed characters
func validateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	for i, r := range name {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-') {
			return fmt.Errorf("invalid character in device name at position %d: %c", i, r)
		}
	}

	return nil
}

// validateFilePath ensures file paths are safe
func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check for directory traversal attempts
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return fmt.Errorf("invalid file path: possible directory traversal attempt")
	}

	// Additional file path validation
	for _, r := range path {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' || r == '_') {
			return fmt.Errorf("invalid character in file path: %c", r)
		}
	}

	return nil
}

// baseDeviceCommand provides common functionality for device commands
type baseDeviceCommand struct {
	executor *CommandExecutor
	name     string   // Validated device name
	subCmd   []string // Static sub-command
}

// execute safely executes a command built from validated components
func (c *baseDeviceCommand) execute(ctx context.Context) error {
	args := append(append([]string{}, c.executor.baseArgs...), c.subCmd...)
	args = append(args, c.name)

	cmd := exec.CommandContext(ctx, c.executor.execPath)
	cmd.Args = append([]string{c.executor.execPath}, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RegisterDeviceCommand represents a validated device registration command
type RegisterDeviceCommand struct {
	baseDeviceCommand
}

// NewRegisterDeviceCommand creates a new device registration command
func (e *CommandExecutor) NewRegisterDeviceCommand(name string) (*RegisterDeviceCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &RegisterDeviceCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"register", "--name"},
		},
	}, nil
}

// Execute runs the registration command
func (c *RegisterDeviceCommand) Execute(ctx context.Context) error {
	return c.execute(ctx)
}

// GetDeviceCommand represents a validated device info retrieval command
type GetDeviceCommand struct {
	baseDeviceCommand
}

// NewGetDeviceCommand creates a new device info retrieval command
func (e *CommandExecutor) NewGetDeviceCommand(name string) (*GetDeviceCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &GetDeviceCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"get"},
		},
	}, nil
}

// Execute runs the get command
func (c *GetDeviceCommand) Execute(ctx context.Context) error {
	return c.execute(ctx)
}

// DeviceStatusCommand represents a validated device status command
type DeviceStatusCommand struct {
	baseDeviceCommand
}

// NewDeviceStatusCommand creates a new device status command
func (e *CommandExecutor) NewDeviceStatusCommand(name string) (*DeviceStatusCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &DeviceStatusCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"status"},
		},
	}, nil
}

// Execute runs the status command
func (c *DeviceStatusCommand) Execute(ctx context.Context) error {
	return c.execute(ctx)
}

// DeviceHealthCommand represents a validated device health command
type DeviceHealthCommand struct {
	baseDeviceCommand
}

// NewDeviceHealthCommand creates a new device health command
func (e *CommandExecutor) NewDeviceHealthCommand(name string) (*DeviceHealthCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &DeviceHealthCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"health"},
		},
	}, nil
}

// Execute runs the health command
func (c *DeviceHealthCommand) Execute(ctx context.Context) error {
	return c.execute(ctx)
}

// DeviceAlertsCommand represents a validated device alerts command
type DeviceAlertsCommand struct {
	baseDeviceCommand
}

// NewDeviceAlertsCommand creates a new device alerts command
func (e *CommandExecutor) NewDeviceAlertsCommand(name string) (*DeviceAlertsCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &DeviceAlertsCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"alerts"},
		},
	}, nil
}

// Execute runs the alerts command
func (c *DeviceAlertsCommand) Execute(ctx context.Context) error {
	return c.execute(ctx)
}

// DeviceConfigCommand represents a validated device config command
type DeviceConfigCommand struct {
	baseDeviceCommand
	action string
	file   string // Optional config file path
}

// NewDeviceConfigGetCommand creates a new device config get command
func (e *CommandExecutor) NewDeviceConfigGetCommand(name string) (*DeviceConfigCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &DeviceConfigCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"config", "get"},
		},
		action: "get",
	}, nil
}

// NewDeviceConfigSetCommand creates a new device config set command
func (e *CommandExecutor) NewDeviceConfigSetCommand(name, configFile string) (*DeviceConfigCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}
	if err := validateFilePath(configFile); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &DeviceConfigCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"config", "set"},
		},
		action: "set",
		file:   configFile,
	}, nil
}

// NewDeviceConfigHistoryCommand creates a new device config history command
func (e *CommandExecutor) NewDeviceConfigHistoryCommand(name string) (*DeviceConfigCommand, error) {
	if err := validateDeviceName(name); err != nil {
		return nil, fmt.Errorf("invalid device name: %w", err)
	}

	return &DeviceConfigCommand{
		baseDeviceCommand: baseDeviceCommand{
			executor: e,
			name:     name,
			subCmd:   []string{"config", "history"},
		},
		action: "history",
	}, nil
}

// Execute runs the config command with appropriate arguments
func (c *DeviceConfigCommand) Execute(ctx context.Context) error {
	if c.action == "set" {
		args := append(append([]string{}, c.executor.baseArgs...), c.subCmd...)
		args = append(args, c.name, "--file", c.file)

		cmd := exec.CommandContext(ctx, c.executor.execPath)
		cmd.Args = append([]string{c.executor.execPath}, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return c.execute(ctx)
}
