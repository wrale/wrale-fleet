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
	// Parse command-line flags
	cfg := options.New()
	flag.StringVar(&cfg.Port, "port", "8080", "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", "/var/lib/wfcentral", "Data directory path")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Logging level (debug, info, warn, error)")
	flag.Parse()

	// Initialize logger
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

	// Initialize server with Stage 1 capabilities
	srv, err := options.NewServer(
		options.WithPort(cfg.Port),
		options.WithDataDir(cfg.DataDir),
	)
	if err != nil {
		log.Fatal("failed to initialize server", err)
		os.Exit(1)
	}

	// Signal successful initialization if in test mode
	if initDone != nil {
		close(initDone)
	}

	// Create root context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info("received shutdown signal", "signal", sig.String())
		// Create context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()
		cancel()
	}()

	// Start server
	log.Info("starting wfcentral server",
		"port", cfg.Port,
		"data_dir", cfg.DataDir,
		"log_level", cfg.LogLevel,
	)

	if err := srv.Run(ctx); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	log.Info("shutdown completed successfully")
}
