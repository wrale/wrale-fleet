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
	// Regular entry point runs without init signal
	mainWithInit(nil)
}

// mainWithInit is the main program logic, optionally signaling initialization.
// The initDone channel is used for testing to coordinate program startup.
func mainWithInit(initDone chan<- struct{}) {
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

	// Signal successful initialization if in test mode
	if initDone != nil {
		close(initDone)
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	// Begin graceful shutdown
	logger.Info("initiating shutdown sequence")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// Stop demo manager with timeout
	if err := demoManager.Stop(shutdownCtx); err != nil {
		logger.Error("failed to stop demo manager", zap.Error(err))
		os.Exit(1)
	}

	// Clean up resources
	if closer, ok := store.(interface{ Close() error }); ok {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), cleanupTimeout)
		defer cleanupCancel()

		// Create done channel for cleanup
		done := make(chan struct{})
		go func() {
			if err := closer.Close(); err != nil {
				logger.Error("failed to close store", zap.Error(err))
				os.Exit(1)
			}
			close(done)
		}()

		// Wait for cleanup or timeout
		select {
		case <-done:
			logger.Info("store cleanup completed")
		case <-cleanupCtx.Done():
			logger.Error("store cleanup timed out")
			os.Exit(1)
		}
	}

	logger.Info("shutdown completed successfully")
}
