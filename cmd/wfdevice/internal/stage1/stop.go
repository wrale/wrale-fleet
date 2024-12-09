package stage1

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

const (
	stopTimeout = 5 * time.Second
)

func newStopCmd(cfg *options.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the device agent gracefully",
		Long: `Stop the device agent, allowing time for graceful shutdown.

The stop command:
- Signals the agent to begin shutdown
- Waits for operations to complete
- Ensures clean termination
- Times out after 5 seconds

A graceful stop ensures:
- Active operations complete
- Resources are cleaned up
- Control plane is notified
- State is properly saved`,
		Example: `  # Stop the agent
  wfdevice stop`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(cmd.Context(), cfg)
		},
	}

	return cmd, nil
}

func runStop(ctx context.Context, cfg *options.Config) error {
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(ctx, stopTimeout)
	defer cancel()

	// Get handle to running server
	srv, err := options.GetRunningServer()
	if err != nil {
		return fmt.Errorf("getting server handle: %w", err)
	}

	// Initiate graceful shutdown
	if err := srv.Stop(ctx); err != nil {
		return fmt.Errorf("stopping server: %w", err)
	}

	return nil
}
