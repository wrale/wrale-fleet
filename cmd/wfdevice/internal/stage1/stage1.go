// Package stage1 implements the Stage 1 commands for basic device management.
package stage1

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

// AddCommands adds all Stage 1 commands to the root command.
// This function serves as the main entry point for the stage1 package,
// organizing commands into logical groups and maintaining a clear hierarchy.
func AddCommands(root *cobra.Command, cfg *options.Config) error {
	// Agent lifecycle commands
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

	// Registration commands
	registerCmd, err := newRegisterCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating register command: %w", err)
	}
	root.AddCommand(registerCmd)

	// Shutdown notification
	notifyShutdownCmd, err := newNotifyShutdownCmd(cfg)
	if err != nil {
		return fmt.Errorf("creating notify-shutdown command: %w", err)
	}
	root.AddCommand(notifyShutdownCmd)

	return nil
}
