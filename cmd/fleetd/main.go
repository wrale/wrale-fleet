package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()

	// Create device store and service
	store := memory.New()
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

	// Demonstrate core device management features
	if err := runDemo(ctx, service, logger); err != nil {
		logger.Fatal("demo failed", zap.Error(err))
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Sync logger before final message - ignore sync errors as they're expected on some platforms
	_ = logger.Sync()

	// Log final shutdown message
	logger.Info("shutting down")
}

func runDemo(ctx context.Context, service *device.Service, logger *zap.Logger) error {
	const tenantID = "demo-tenant"

	// 1. Device Registration
	demoDevice, err := service.Register(ctx, tenantID, "Demo Raspberry Pi")
	if err != nil {
		return err
	}

	logger.Info("registered demo device",
		zap.String("device_id", demoDevice.ID),
		zap.String("tenant_id", demoDevice.TenantID),
		zap.String("name", demoDevice.Name),
	)

	// 2. Add device metadata
	if err := demoDevice.AddTag("environment", "production"); err != nil {
		return err
	}
	if err := demoDevice.AddTag("location", "datacenter-1"); err != nil {
		return err
	}

	if err := service.Update(ctx, demoDevice); err != nil {
		return err
	}

	logger.Info("updated device tags",
		zap.String("device_id", demoDevice.ID),
		zap.Any("tags", demoDevice.Tags),
	)

	// 3. Update device status
	if err := service.UpdateStatus(ctx, tenantID, demoDevice.ID, device.StatusOnline); err != nil {
		return err
	}

	// 4. Configure device
	config := map[string]interface{}{
		"monitoring_interval": "30s",
		"log_level":          "info",
		"features": map[string]bool{
			"metrics_enabled":  true,
			"tracing_enabled":  false,
			"alerting_enabled": true,
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := demoDevice.SetConfig(configJSON, "admin"); err != nil {
		return err
	}

	if err := service.Update(ctx, demoDevice); err != nil {
		return err
	}

	logger.Info("applied device configuration",
		zap.String("device_id", demoDevice.ID),
		zap.String("config_hash", demoDevice.LastConfigHash),
	)

	// 5. Validate configuration
	if err := demoDevice.ValidateConfig(); err != nil {
		return err
	}

	if err := service.Update(ctx, demoDevice); err != nil {
		return err
	}

	// 6. Update network information
	networkInfo := &device.NetworkInfo{
		IPAddress:  "192.168.1.100",
		MACAddress: "00:11:22:33:44:55",
		Hostname:   "demo-device-1",
		Port:       9100,
	}

	if err := demoDevice.UpdateDiscoveryInfo(device.DiscoveryManual, networkInfo); err != nil {
		return err
	}

	if err := service.Update(ctx, demoDevice); err != nil {
		return err
	}

	logger.Info("updated device network info",
		zap.String("device_id", demoDevice.ID),
		zap.String("ip", networkInfo.IPAddress),
		zap.String("hostname", networkInfo.Hostname),
	)

	// 7. Demonstrate offline capabilities
	offlineCapabilities := &device.OfflineCapabilities{
		SupportsAirgap: true,
		SyncInterval:   time.Hour,
		LastSyncTime:   time.Now(),
		OfflineOperations: []string{
			"status_update",
			"metric_collection",
			"log_collection",
		},
		LocalBufferSize: 1024 * 1024 * 100, // 100MB
	}

	if err := demoDevice.UpdateOfflineCapabilities(offlineCapabilities); err != nil {
		return err
	}

	if err := service.Update(ctx, demoDevice); err != nil {
		return err
	}

	logger.Info("configured offline capabilities",
		zap.String("device_id", demoDevice.ID),
		zap.Bool("airgap_supported", offlineCapabilities.SupportsAirgap),
		zap.Duration("sync_interval", offlineCapabilities.SyncInterval),
	)

	// 8. List devices
	devices, err := service.List(ctx, device.ListOptions{
		TenantID: tenantID,
		Tags: map[string]string{
			"environment": "production",
		},
	})

	if err != nil {
		return err
	}

	logger.Info("listed devices",
		zap.Int("count", len(devices)),
		zap.String("tenant_id", tenantID),
	)

	return nil
}