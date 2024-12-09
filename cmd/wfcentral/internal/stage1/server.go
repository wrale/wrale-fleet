package stage1

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/logger"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/options"
	"go.uber.org/zap"
)

const (
	shutdownTimeout = 5 * time.Second
)

// newStartCmd creates the start command
func newStartCmd(cfg *options.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the control plane server",
		Long: `Start the wfcentral server and begin serving device management requests.

The server requires both a main API port for device management and a separate
management port for health and readiness endpoints. The management port must
be explicitly configured for security reasons.`,
		Example: `  # Start server with default settings
  wfcentral start --management-port 8601

  # Start with custom ports and data directory
  wfcentral start --port 8700 --management-port 8701 --data-dir /data/wfcentral

  # Start with full health endpoint exposure
  wfcentral start --management-port 8601 --health-exposure full`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.Port, "port", cfg.Port,
		"main API port for device management")
	cmd.Flags().StringVar(&cfg.ManagementPort, "management-port", cfg.ManagementPort,
		"management API port for health and readiness endpoints")
	cmd.Flags().StringVar(&cfg.DataDir, "data-dir", cfg.DataDir,
		"data directory path")
	cmd.Flags().StringVar(&cfg.HealthExposure, "health-exposure", cfg.HealthExposure,
		"level of information exposed in health endpoints (minimal, standard, full)")

	if err := cmd.MarkFlagRequired("management-port"); err != nil {
		return nil, fmt.Errorf("marking management-port flag as required: %w", err)
	}

	return cmd, nil
}

// newStopCmd creates the stop command
func newStopCmd(cfg *options.Config) (*cobra.Command, error) {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the control plane server gracefully",
		Long: `Initiate graceful shutdown of the wfcentral server.

The stop command sends a shutdown signal to the server and waits for all
current operations to complete or timeout. If the server does not shut down
within the timeout period, it will be forcefully terminated.`,
		Example: `  # Stop the server
  wfcentral stop

  # Stop and wait for specific port
  wfcentral stop --port 8700`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopServer(cmd.Context(), cfg)
		},
	}, nil
}

// newStatusCmd creates the status command
func newStatusCmd(cfg *options.Config) (*cobra.Command, error) {
	return &cobra.Command{
		Use:   "status",
		Short: "Show server status and health",
		Long: `Display detailed status information about the running server.

The status command connects to the server's management port to retrieve
health and status information. The amount of information displayed depends
on the server's configured health exposure level.`,
		Example: `  # Check server status
  wfcentral status

  # Check status for server on custom port
  wfcentral status --port 8700`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkServerStatus(cmd.Context(), cfg)
		},
	}, nil
}

// startServer implements the start command functionality
func startServer(ctx context.Context, cfg *options.Config) error {
	// Initialize logger for command-line operations
	log, err := logger.New(logger.Config{
		Level:    cfg.LogLevel,
		FilePath: cfg.LogFile,
	})
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer func() {
		if err := logger.Sync(log); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync warning: %v\n", err)
		}
	}()

	// Initialize server with configuration
	srv, err := options.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("initializing server: %w", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	log.Info("starting wfcentral server",
		zap.String("api_port", cfg.Port),
		zap.String("management_port", cfg.ManagementPort),
		zap.String("health_exposure", cfg.HealthExposure),
		zap.String("data_dir", cfg.DataDir),
		zap.String("log_level", cfg.LogLevel),
	)

	// Handle shutdown signal in a separate goroutine
	go func() {
		sig := <-sigChan
		log.Info("received shutdown signal", zap.String("signal", sig.String()))

		// Create context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Wait for shutdown to complete or timeout
		select {
		case <-shutdownCtx.Done():
			log.Warn("shutdown timed out", zap.Duration("timeout", shutdownTimeout))
		case <-ctx.Done():
			log.Info("shutdown completed")
		}
	}()

	// Run server until shutdown
	if err := srv.Start(ctx); err != nil {
		log.Error("server error", zap.Error(err))
		return err
	}

	log.Info("shutdown completed successfully")
	return nil
}

// stopServer implements the stop command functionality
func stopServer(ctx context.Context, cfg *options.Config) error {
	// Implementation would connect to the management API and initiate shutdown
	return fmt.Errorf("not implemented")
}

// checkServerStatus implements the status command functionality
func checkServerStatus(ctx context.Context, cfg *options.Config) error {
	// Implementation would connect to the management API and retrieve status
	return fmt.Errorf("not implemented")
}
