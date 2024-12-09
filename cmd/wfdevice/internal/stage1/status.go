package stage1

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

func newStatusCmd(cfg *options.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent status and health",
		Long: `Display current status and health information for the device agent.

The status command shows:
- Agent running state
- Connection status
- Basic health metrics
- Resource utilization
- Recent operations
- Current configuration`,
		Example: `  # Show agent status
  wfdevice status

  # Show detailed status with debug information
  wfdevice status --log-level debug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), cfg)
		},
	}

	return cmd, nil
}

func runStatus(ctx context.Context, cfg *options.Config) error {
	// Get handle to running server
	srv, err := options.GetRunningServer()
	if err != nil {
		return fmt.Errorf("getting server handle: %w", err)
	}

	// Get status information
	status, err := srv.Status(ctx)
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}

	// Display status information
	return printStatus(status)
}

func printStatus(status interface{}) error {
	// TODO: Implement pretty printing of status information
	fmt.Printf("%+v\n", status)
	return nil
}
