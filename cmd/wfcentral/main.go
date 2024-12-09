// Package main implements the wfcentral command, which provides the central control plane
// for managing global fleets of devices in the Wrale Fleet Management Platform.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wrale/wrale-fleet/cmd/wfcentral/logger"
	"github.com/wrale/wrale-fleet/cmd/wfcentral/options"
	"go.uber.org/zap"
)

const (
	// shutdownTimeout is the maximum time allowed for graceful shutdown
	shutdownTimeout = 5 * time.Second
)

func main() {
	// Regular entry point runs without init signal
	mainWithInit(nil)
}

// Command represents a subcommand with its own flags
type Command struct {
	fs *flag.FlagSet
	fn func(ctx context.Context, cfg *options.Config) error
}

// mainWithInit is the main program logic, optionally signaling initialization.
// The initDone channel is used for testing to coordinate program startup.
func mainWithInit(initDone chan<- struct{}) {
	// Create default config for command-line operations
	cfg := options.New()

	// Define commands
	commands := map[string]Command{
		"start": {
			fs: flag.NewFlagSet("start", flag.ExitOnError),
			fn: startServer,
		},
	}

	// Register flags for start command
	startCmd := commands["start"]
	startCmd.fs.StringVar(&cfg.Port, "port", cfg.Port, "Main API port for device management")
	startCmd.fs.StringVar(&cfg.ManagementPort, "management-port", cfg.ManagementPort, "Management API port for health and readiness endpoints")
	startCmd.fs.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Data directory path")
	startCmd.fs.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level (debug, info, warn, error)")
	startCmd.fs.StringVar(&cfg.HealthExposure, "health-exposure", cfg.HealthExposure,
		"Level of information exposed in health endpoints (minimal, standard, full)")
	startCmd.fs.StringVar(&cfg.LogFile, "log-file", "", "Log file path (defaults to stdout)")

	// No arguments shows usage
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n\nCommands:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  start    Start the control plane server\n")
		os.Exit(1)
	}

	// Get the command and verify it exists
	cmdName := os.Args[1]
	cmd, exists := commands[cmdName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
		os.Exit(1)
	}

	// Parse command-specific flags
	if err := cmd.fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Create context for program lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the command
	if err := cmd.fn(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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

	// Signal successful initialization if in test mode
	if initDone != nil {
		close(initDone)
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
