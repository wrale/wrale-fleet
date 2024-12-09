// Package root provides the root command for the wfdevice CLI.
package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/internal/stage1"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

// New creates and configures the root command for wfdevice.
func New() (*cobra.Command, error) {
	cfg := options.New()

	cmd := &cobra.Command{
		Use:   "wfdevice",
		Short: "Local device management for the Wrale Fleet Management Platform",
		Long: `wfdevice provides local device management capabilities including status reporting,
configuration management, health monitoring, and secure communication with the
control plane.

Device operations include:
  - Starting and stopping the device agent
  - Registering with the control plane
  - Reporting device status and health
  - Applying and validating configurations
  - Managing secure communication`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Configure global flags
	flags := cmd.PersistentFlags()

	// Logging flags
	flags.StringVar(&cfg.LogLevel, "log-level", "info",
		"logging level (debug, info, warn, error)")
	flags.StringVar(&cfg.LogFile, "log-file", "",
		"log file path (defaults to stdout)")
	flags.BoolVar(&cfg.LogJSON, "log-json", false,
		"enable JSON log format")
	flags.IntVar(&cfg.LogStage, "log-stage", 1,
		"enable stage-aware logging (1-6)")

	// Server flags
	flags.StringVar(&cfg.DataDir, "data-dir", "/var/lib/wfdevice",
		"data directory path")
	flags.StringVar(&cfg.Port, "port", "9090",
		"agent port")

	// Add staged command groups
	if err := stage1.AddCommands(cmd, cfg); err != nil {
		return nil, fmt.Errorf("adding stage1 commands: %w", err)
	}

	// Custom error handling to maintain consistent error reporting
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("invalid flag: %w", err)
	})

	return cmd, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once.
func Execute() {
	cmd, err := New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
