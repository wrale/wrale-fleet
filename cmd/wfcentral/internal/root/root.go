// Package root provides the root command for the wfcentral CLI.
package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/internal/stage1"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/options"
)

// New creates and configures the root command for wfcentral.
func New() *cobra.Command {
	cfg := options.New()

	cmd := &cobra.Command{
		Use:           "wfcentral",
		Short:         "Enterprise control plane for the Wrale Fleet Management Platform",
		Long:          `wfcentral provides global fleet management capabilities including device registration, configuration management, monitoring, and multi-region coordination.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Configure global flags
	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel,
		"logging level (debug, info, warn, error)")
	cmd.PersistentFlags().StringVar(&cfg.LogFile, "log-file", "",
		"log file path (defaults to stdout)")

	// Add staged command groups
	stage1.AddCommands(cmd, cfg)

	// Custom error handling to maintain consistent error reporting
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("invalid flag: %w", err)
	})

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once.
func Execute() {
	if err := New().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
