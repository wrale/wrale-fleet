package main

import (
	"errors"
	"os/exec"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/wrale/fleet/cmd/wfdemo/sysadmin"
)

// sysadminCmd creates the command tree for SysAdmin persona demos
func sysadminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sysadmin",
		Short: "System Administrator demonstrations",
		Long: `Demonstrations targeting System Administrators, showing core device
management and monitoring capabilities of the platform.`,
	}

	// Add stage-specific commands
	cmd.AddCommand(sysadminStage1Cmd())

	return cmd
}

func sysadminStage1Cmd() *cobra.Command {
	var wfcentralPath string

	cmd := &cobra.Command{
		Use:   "stage1",
		Short: "Stage 1: Basic Device Management",
		Long: `Demonstrates core device management capabilities including:
- Device registration and provisioning
- Status monitoring and health checks
- Basic configuration management`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Verify wfcentral is available
			if wfcentralPath == "" {
				return errors.New("wfcentral path not specified")
			}
			if _, err := exec.LookPath(wfcentralPath); err != nil {
				return errors.New("wfcentral not found in path")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := cmd.Context().Value("logger").(*zap.Logger)
			scenarios := sysadmin.Stage1Scenarios(logger, wfcentralPath)
			runner := NewRunner(logger, scenarios)
			return runner.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&wfcentralPath, "wfcentral", "wfcentral", "path to wfcentral binary")

	return cmd
}
