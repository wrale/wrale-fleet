package sysadmin

import (
	"context"
	"fmt"
)

// commandExecutor provides backward compatibility for the demo scenarios
type commandExecutor struct {
	cmd *DeviceCommand
}

// newCommandExecutor creates a new command executor with backward compatibility
func newCommandExecutor(wfcentralPath string) (*commandExecutor, error) {
	cmd, err := NewDeviceCommand(wfcentralPath)
	if err != nil {
		return nil, err
	}
	return &commandExecutor{cmd: cmd}, nil
}

// executeCommand provides backward compatibility for string slice-based command execution
func (c *commandExecutor) executeCommand(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("invalid command: insufficient arguments")
	}

	// Map string slice commands to type-safe methods
	switch args[0] {
	case "device":
		switch args[1] {
		case "register":
			if len(args) != 4 || args[2] != "--name" {
				return fmt.Errorf("invalid device register command")
			}
			return c.cmd.RegisterDevice(ctx, args[3])

		case "get":
			if len(args) != 3 {
				return fmt.Errorf("invalid device get command")
			}
			return c.cmd.GetDevice(ctx, args[2])

		case "status":
			if len(args) != 3 {
				return fmt.Errorf("invalid device status command")
			}
			return c.cmd.GetDeviceStatus(ctx, args[2])

		case "health":
			if len(args) != 3 {
				return fmt.Errorf("invalid device health command")
			}
			return c.cmd.GetDeviceHealth(ctx, args[2])

		case "alerts":
			if len(args) != 3 {
				return fmt.Errorf("invalid device alerts command")
			}
			return c.cmd.GetDeviceAlerts(ctx, args[2])

		case "config":
			if len(args) < 4 {
				return fmt.Errorf("invalid device config command")
			}

			switch args[2] {
			case "get":
				return c.cmd.GetDeviceConfig(ctx, args[3])

			case "set":
				if len(args) != 6 || args[4] != "--file" {
					return fmt.Errorf("invalid device config set command")
				}
				return c.cmd.SetDeviceConfig(ctx, args[3], args[5])

			case "history":
				return c.cmd.GetDeviceConfigHistory(ctx, args[3])

			default:
				return fmt.Errorf("unknown config subcommand: %s", args[2])
			}

		default:
			return fmt.Errorf("unknown device subcommand: %s", args[1])
		}

	default:
		return fmt.Errorf("unknown command type: %s", args[0])
	}
}
