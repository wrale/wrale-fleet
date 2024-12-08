// Package main implements the wfdevice command, which provides local device management
// capabilities for the Wrale Fleet Management Platform.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/wrale/wrale-fleet/cmd/wfdevice/logger"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
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
	// Parse command-line flags
	cfg := options.New()
	flag.StringVar(&cfg.Port, "port", "9090", "Agent port")
	flag.StringVar(&cfg.DataDir, "data-dir", "/var/lib/wfdevice", "Data directory path")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Logging level (debug, info, warn, error)")
	flag.StringVar(&cfg.Name, "name", "", "Device name")
	flag.StringVar(&cfg.ControlPlane, "control-plane", "", "Control plane address")

	// Handle tags as comma-separated key=value pairs
	var tags string
	flag.StringVar(&tags, "tags", "", "Device tags (key=value,key2=value2)")
	flag.Parse()

	// Parse tags if provided
	if tags != "" {
		pairs := strings.Split(tags, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				cfg.Tags[kv[0]] = kv[1]
			}
		}
	}

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
		options.WithName(cfg.Name),
		options.WithControlPlane(cfg.ControlPlane),
		options.WithTags(cfg.Tags),
	)
	if err != nil {
		log.Fatal("failed to initialize server", zap.Error(err))
		os.Exit(1)
	}

	// Signal successful initialization if in test mode
	if initDone != nil {
		close(initDone)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	log.Info("starting wfdevice agent",
		zap.String("name", cfg.Name),
		zap.String("port", cfg.Port),
		zap.String("data_dir", cfg.DataDir),
		zap.String("log_level", cfg.LogLevel),
		zap.String("control_plane", cfg.ControlPlane),
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

	if err := srv.Run(ctx); err != nil {
		log.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("shutdown completed successfully")
}
