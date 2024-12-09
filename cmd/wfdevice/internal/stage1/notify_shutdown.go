package stage1

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

func newNotifyShutdownCmd(cfg *options.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "notify-shutdown",
		Short: "Signal planned shutdown to control plane",
		Long: `Notify the control plane of a planned device shutdown.

This command ensures graceful handling of planned shutdowns by:
- Informing the control plane of shutdown intent
- Allowing pending operations to complete
- Ensuring state is properly saved
- Preventing unnecessary alerts

This should be called before planned maintenance or
controlled shutdowns to maintain system health.`,
		Example: `  # Notify planned shutdown
  wfdevice notify-shutdown`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotifyShutdown(cmd.Context(), cfg)
		},
	}

	return cmd, nil
}

func runNotifyShutdown(ctx context.Context, cfg *options.Config) error {
	// Get handle to running server
	srv, err := options.GetRunningServer()
	if err != nil {
		return fmt.Errorf("getting server handle: %w", err)
	}

	// Send shutdown notification
	if err := srv.NotifyShutdown(ctx); err != nil {
		return fmt.Errorf("notifying shutdown: %w", err)
	}

	fmt.Println("Successfully notified control plane of planned shutdown")
	return nil
}
