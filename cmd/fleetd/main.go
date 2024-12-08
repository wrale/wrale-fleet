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
)

func main() {
	// Initialize logger
	logger, err := setupLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
	}()

	// Create device store and service
	store := memory.New()
	service := device.NewService(store, logger)

	// Handle shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create error channel for main goroutine
	errChan := make(chan error, 1)

	// Run demo in separate goroutine
	go func() {
		if err := runDemo(ctx, service, logger); err != nil {
			logger.Error("demo failed", zap.Error(err))
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// Wait for either signal or demo completion
	var shutdownErr error
	select {
	case sig := <-sigChan:
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()

		// Create shutdown timeout context
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		// Wait for demo to finish or timeout
		select {
		case err := <-errChan:
			shutdownErr = err
		case <-shutdownCtx.Done():
			logger.Warn("shutdown timed out")
			shutdownErr = shutdownCtx.Err()
		}

	case err := <-errChan:
		shutdownErr = err
	}

	// Begin graceful shutdown
	logger.Info("shutting down")

	// Allow a moment for final cleanup operations
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
	defer cleanupCancel()

	// Wait for cleanup or timeout
	select {
	case <-time.After(500 * time.Millisecond):
		logger.Info("cleanup completed")
	case <-cleanupCtx.Done():
		logger.Warn("cleanup timed out")
	}

	// Exit with appropriate status
	if shutdownErr != nil {
		logger.Error("shutdown completed with error", zap.Error(shutdownErr))
		os.Exit(1)
	}

	logger.Info("shutdown completed successfully")
}
