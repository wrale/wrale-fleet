package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap"
)

const (
	shutdownTimeout = 5 * time.Second
	cleanupTimeout  = time.Second
)

func main() {
	// Initialize logger with enhanced error handling
	logger, err := setupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := safeSync(logger); err != nil {
			// Logger sync errors are expected on some platforms, don't fail
			fmt.Fprintf(os.Stderr, "logger sync warning: %v\n", err)
		}
	}()

	// Create device store and service
	store := memory.New()
	service := device.NewService(store, logger)

	// Create and start demo manager
	demoManager := NewDemoManager(service, logger)
	if err := demoManager.Start(); err != nil {
		logger.Error("failed to start demo manager", zap.Error(err))
		os.Exit(1)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Begin graceful shutdown
	logger.Info("initiating shutdown sequence")

	// Stop demo manager
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := demoManager.Stop(); err != nil {
		logger.Error("failed to stop demo manager", zap.Error(err))
		os.Exit(1)
	}

	// Clean up resources
	if closer, ok := store.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.Error("failed to close store", zap.Error(err))
			os.Exit(1)
		}
	}

	logger.Info("shutdown completed successfully")
}
