package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/store/memory"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create device store and service
	store := memory.NewDeviceStore()
	service := device.NewService(store, logger)

	// Create context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Register a demo device
	demoDevice, err := service.Register(ctx, "demo-tenant", "Demo Raspberry Pi")
	if err != nil {
		logger.Fatal("failed to register demo device", zap.Error(err))
	}

	logger.Info("registered demo device",
		zap.String("device_id", demoDevice.ID),
		zap.String("tenant_id", demoDevice.TenantID),
		zap.String("name", demoDevice.Name),
	)

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("shutting down")
}
