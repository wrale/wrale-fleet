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

The agent requires both a main API port for device operations and a separate
management port for health and readiness endpoints. The management port must
be explicitly configured for security reasons.

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
		Example: `  # Start with default settings (requires management port)
  wfdevice start --management-port 9091

  # Start with custom ports and data directory
  wfdevice start --port 9090 --management-port 9091 --data-dir /data/wfdevice

  # Start with full health endpoint exposure
  wfdevice start --management-port 9091 --health-exposure full

  # Start with device name and control plane connection
  wfdevice start --management-port 9091 --name device1 --control-plane localhost:8600`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd.Context(), cfg)
		},
	}

	// Add flags that align with wfcentral's configuration
	cmd.Flags().StringVar(&cfg.Port, "port", cfg.Port,
		"main API port for device operations")
	cmd.Flags().StringVar(&cfg.ManagementPort, "management-port", "",
		"management API port for health and readiness endpoints")
	cmd.Flags().StringVar(&cfg.HealthExposure, "health-exposure", cfg.HealthExposure,
		"level of information exposed in health endpoints (minimal, standard, full)")
	cmd.Flags().StringVar(&cfg.Name, "name", cfg.Name,
		"device name for identification")
	cmd.Flags().StringVar(&cfg.ControlPlane, "control-plane", cfg.ControlPlane,
		"control plane address for registration")

	// Mark management port as required for security
	if err := cmd.MarkFlagRequired("management-port"); err != nil {
		return nil, fmt.Errorf("marking management-port flag as required: %w", err)
	}

	return cmd, nil
}

func runStart(ctx context.Context, cfg *options.Config) error {
	// Initialize and run the server with validated configuration
	srv, err := options.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("initializing server: %w", err)
	}

	// Store as running server
	options.SetRunningServer(srv)
	defer options.ClearRunningServer()

	return srv.Run(ctx)
}
