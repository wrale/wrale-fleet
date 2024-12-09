package stage1

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

func newStartCmd(cfg *options.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the device agent",
		Long: `Start the device agent with the specified configuration.

The agent will:
- Initialize system components
- Connect to the control plane if registered
- Begin health monitoring
- Handle device operations
- Report status and metrics

The agent runs until stopped by either:
- The stop command
- SIGINT (Ctrl+C)
- SIGTERM`,
		Example: `  # Start with default settings
  wfdevice start

  # Start with custom port and data directory
  wfdevice start --port 9091 --data-dir /opt/wfdevice

  # Start with debug logging
  wfdevice start --log-level debug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd.Context(), cfg)
		},
	}

	return cmd, nil
}

func runStart(ctx context.Context, cfg *options.Config) error {
	// Validate required configuration for starting the agent
	if err := validateStartConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize and run the server
	srv, err := options.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("initializing server: %w", err)
	}

	return srv.Run(ctx)
}

func validateStartConfig(cfg *options.Config) error {
	if cfg.Port == "" {
		return fmt.Errorf("port is required")
	}
	if cfg.DataDir == "" {
		return fmt.Errorf("data directory is required")
	}
	return nil
}
