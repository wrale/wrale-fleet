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

// mainWithInit is the main program logic, optionally signaling initialization.
// The initDone channel is used for testing to coordinate program startup.
func mainWithInit(initDone chan<- struct{}) {
	// Create default config and parse command-line flags
	cfg := options.New()
	flag.StringVar(&cfg.Port, "port", cfg.Port, "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Data directory path")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level (debug, info, warn, error)")
	flag.Parse()

	// Initialize logger for command-line operations
	log, err := logger.New(logger.Config{Level: cfg.LogLevel})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(log); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync warning: %v\n", err)
		}
	}()

	// Initialize server with configuration
	srv, err := options.NewServer(cfg)
	if err != nil {
		log.Fatal("failed to initialize server", zap.Error(err))
	}

	// Signal successful initialization if in test mode
	if initDone != nil {
		close(initDone)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	log.Info("starting wfcentral server",
		zap.String("port", cfg.Port),
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

		// Trigger graceful shutdown
		cancel()

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
		os.Exit(1)
	}

	log.Info("shutdown completed successfully")
}
