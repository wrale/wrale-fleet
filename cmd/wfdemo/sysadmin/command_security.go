package sysadmin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"unicode"
)

// CommandType represents a validated command type
type CommandType int

const (
	CmdDeviceRegister CommandType = iota
	CmdDeviceGet
	CmdDeviceStatus
	CmdDeviceHealth
	CmdDeviceAlerts
	CmdDeviceConfigGet
	CmdDeviceConfigSet
	CmdDeviceConfigHistory
)

// commandDef defines a command's structure and validation rules
type commandDef struct {
	args     []string
	maxArgs  int
	validate func([]string) error
}

// commandExecutor provides secure command execution for demo scenarios
type commandExecutor struct {
	wfcentralPath string
	commands      map[CommandType]commandDef
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

// newCommandExecutor creates a new command executor with security validation
func newCommandExecutor(wfcentralPath string) (*commandExecutor, error) {
	// Validate wfcentral path
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

	// Define command definitions with validation rules
	commands := map[CommandType]commandDef{
		CmdDeviceRegister: {
			args:    []string{"device", "register", "--name"},
			maxArgs: 4,
			validate: func(args []string) error {
				if len(args) != 4 {
					return fmt.Errorf("invalid argument count for device register")
				}
				return validateDeviceName(args[3])
			},
		},
		CmdDeviceGet: {
			args:    []string{"device", "get"},
			maxArgs: 3,
			validate: func(args []string) error {
				if len(args) != 3 {
					return fmt.Errorf("invalid argument count for device get")
				}
				return validateDeviceName(args[2])
			},
		},
		CmdDeviceStatus: {
			args:    []string{"device", "status"},
			maxArgs: 3,
			validate: func(args []string) error {
				if len(args) != 3 {
					return fmt.Errorf("invalid argument count for device status")
				}
				return validateDeviceName(args[2])
			},
		},
		CmdDeviceHealth: {
			args:    []string{"device", "health"},
			maxArgs: 3,
			validate: func(args []string) error {
				if len(args) != 3 {
					return fmt.Errorf("invalid argument count for device health")
				}
				return validateDeviceName(args[2])
			},
		},
		CmdDeviceAlerts: {
			args:    []string{"device", "alerts"},
			maxArgs: 3,
			validate: func(args []string) error {
				if len(args) != 3 {
					return fmt.Errorf("invalid argument count for device alerts")
				}
				return validateDeviceName(args[2])
			},
		},
		CmdDeviceConfigGet: {
			args:    []string{"device", "config", "get"},
			maxArgs: 4,
			validate: func(args []string) error {
				if len(args) != 4 {
					return fmt.Errorf("invalid argument count for config get")
				}
				return validateDeviceName(args[3])
			},
		},
		CmdDeviceConfigSet: {
			args:    []string{"device", "config", "set"},
			maxArgs: 6,
			validate: func(args []string) error {
				if len(args) != 6 || args[4] != "--file" {
					return fmt.Errorf("invalid arguments for config set")
				}
				if err := validateDeviceName(args[3]); err != nil {
					return err
				}
				return validateFilePath(args[5])
			},
		},
		CmdDeviceConfigHistory: {
			args:    []string{"device", "config", "history"},
			maxArgs: 4,
			validate: func(args []string) error {
				if len(args) != 4 {
					return fmt.Errorf("invalid argument count for config history")
				}
				return validateDeviceName(args[3])
			},
		},
	}

	return &commandExecutor{
		wfcentralPath: absPath,
		commands:      commands,
	}, nil
}

// buildCommand safely constructs a command with validated arguments
func (c *commandExecutor) buildCommand(cmdType CommandType, args []string) ([]string, error) {
	def, ok := c.commands[cmdType]
	if !ok {
		return nil, fmt.Errorf("unknown command type")
	}

	if len(args) > def.maxArgs {
		return nil, fmt.Errorf("too many arguments for command")
	}

	if err := def.validate(args); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	return args, nil
}

// executeCommand safely executes a whitelisted command
func (c *commandExecutor) executeCommand(ctx context.Context, args []string) error {
	// Determine command type from args
	var cmdType CommandType
	found := false

	for ct, def := range c.commands {
		if len(args) >= len(def.args) {
			match := true
			for i, arg := range def.args {
				if i >= len(args) || args[i] != arg {
					match = false
					break
				}
			}
			if match {
				cmdType = ct
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("command not allowed: %v", args)
	}

	// Build and validate command
	validatedArgs, err := c.buildCommand(cmdType, args)
	if err != nil {
		return fmt.Errorf("command building failed: %w", err)
	}

	// Execute command with proper context and output handling
	cmd := exec.CommandContext(ctx, c.wfcentralPath, validatedArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
