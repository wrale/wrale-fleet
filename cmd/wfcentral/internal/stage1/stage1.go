// Package stage1 implements the Stage 1 commands for basic device management.
package stage1

import (
	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/options"
)

// AddCommands adds all Stage 1 commands to the root command.
// This function serves as the main entry point for the stage1 package,
// organizing commands into logical groups and maintaining a clear hierarchy.
func AddCommands(root *cobra.Command, cfg *options.Config) {
	// Server lifecycle commands
	root.AddCommand(newStartCmd(cfg))
	root.AddCommand(newStopCmd(cfg))
	root.AddCommand(newStatusCmd(cfg))

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
	deviceCmd.AddCommand(newDeviceListCmd(cfg))
	deviceCmd.AddCommand(newDeviceStatusCmd(cfg))
	deviceCmd.AddCommand(newDeviceHealthCmd(cfg))

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
	configCmd.AddCommand(newConfigShowCmd(cfg))
	configCmd.AddCommand(newConfigValidateCmd(cfg))
	configCmd.AddCommand(newConfigApplyCmd(cfg))

	// Add config commands to device command
	deviceCmd.AddCommand(configCmd)

	// Add device command to root
	root.AddCommand(deviceCmd)
}
