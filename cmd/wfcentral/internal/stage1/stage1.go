// Package stage1 implements the Stage 1 commands for basic device management.
package stage1

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/options"
)

// AddCommands adds all Stage 1 commands to the root command.
// This function serves as the main entry point for the stage1 package,
// organizing commands into logical groups and maintaining a clear hierarchy.
func AddCommands(root *cobra.Command, cfg *options.Config) error {
	// Server lifecycle commands
	startCmd, err := newStartCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating start command: %w", err)
	}
	root.AddCommand(startCmd)

	stopCmd, err := newStopCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating stop command: %w", err)
	}
	root.AddCommand(stopCmd)

	statusCmd, err := newStatusCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating status command: %w", err)
	}
	root.AddCommand(statusCmd)

	// Device management commands
	deviceCmd := &cobra.Command{
		Use:   "device",
		Short: "Manage devices in the fleet",
		Long: `Commands for device lifecycle management, status monitoring, and configuration.

Device commands provide comprehensive management capabilities including:
- Device registration and inventory
- Status monitoring and health checks
- Configuration management and deployment
- Operational control and maintenance`,
		Example: `  # List all registered devices
  wfcentral device list

  # Show detailed device status
  wfcentral device status device-1

  # Show device health metrics
  wfcentral device health device-1

  # Manage device configuration
  wfcentral device config show device-1
  wfcentral device config validate device-1 --config new-config.yaml
  wfcentral device config apply device-1 --config new-config.yaml`,
	}

	// Add device subcommands
	listCmd, err := newDeviceListCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating device list command: %w", err)
	}
	deviceCmd.AddCommand(listCmd)

	statusDeviceCmd, err := newDeviceStatusCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating device status command: %w", err)
	}
	deviceCmd.AddCommand(statusDeviceCmd)

	healthCmd, err := newDeviceHealthCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating device health command: %w", err)
	}
	deviceCmd.AddCommand(healthCmd)

	// Device configuration commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage device configurations",
		Long: `Commands for viewing and managing device configurations.

Configuration management includes:
- Viewing current configurations
- Validating new configurations
- Applying configuration changes
- Monitoring configuration status`,
	}

	// Add configuration subcommands
	showCmd, err := newConfigShowCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating config show command: %w", err)
	}
	configCmd.AddCommand(showCmd)

	validateCmd, err := newConfigValidateCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating config validate command: %w", err)
	}
	configCmd.AddCommand(validateCmd)

	applyCmd, err := newConfigApplyCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating config apply command: %w", err)
	}
	configCmd.AddCommand(applyCmd)

	// Add config commands to device command
	deviceCmd.AddCommand(configCmd)

	// Add device command to root
	root.AddCommand(deviceCmd)

	return nil
}
