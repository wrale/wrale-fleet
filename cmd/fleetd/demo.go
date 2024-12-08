package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wrale/fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// runDemo demonstrates the core device management features
// It properly handles context cancellation throughout its operations
func runDemo(ctx context.Context, service *device.Service, logger *zap.Logger) error {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("demo startup: %w", err)
	}

	const tenantID = "demo-tenant"

	// Device registration
	demoDevice, err := service.Register(ctx, tenantID, "Demo Raspberry Pi")
	if err != nil {
		return fmt.Errorf("device registration: %w", err)
	}

	logger.Info("registered demo device",
		zap.String("device_id", demoDevice.ID),
		zap.String("tenant_id", demoDevice.TenantID),
		zap.String("name", demoDevice.Name),
	)

	// Check context after each major operation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("post-registration: %w", err)
	}

	// Update metadata
	if err := updateDeviceMetadata(ctx, service, demoDevice, logger); err != nil {
		return err
	}

	// Update status
	if err := updateDeviceStatus(ctx, service, demoDevice, logger); err != nil {
		return err
	}

	// Apply configuration
	if err := applyDeviceConfiguration(ctx, service, demoDevice, logger); err != nil {
		return err
	}

	// Update network information
	if err := updateNetworkInfo(ctx, service, demoDevice, logger); err != nil {
		return err
	}

	// Configure offline capabilities
	if err := configureOfflineCapabilities(ctx, service, demoDevice, logger); err != nil {
		return err
	}

	// List devices
	return listDevices(ctx, service, tenantID, logger)
}

// Helper functions to break down the demo into manageable chunks
func updateDeviceMetadata(ctx context.Context, service *device.Service, dev *device.Device, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-metadata: %w", err)
	}

	if err := dev.AddTag("environment", "production"); err != nil {
		return fmt.Errorf("adding environment tag: %w", err)
	}
	if err := dev.AddTag("location", "datacenter-1"); err != nil {
		return fmt.Errorf("adding location tag: %w", err)
	}

	if err := service.Update(ctx, dev); err != nil {
		return fmt.Errorf("updating device tags: %w", err)
	}

	logger.Info("updated device tags",
		zap.String("device_id", dev.ID),
		zap.Any("tags", dev.Tags),
	)

	return nil
}

func updateDeviceStatus(ctx context.Context, service *device.Service, dev *device.Device, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-status: %w", err)
	}

	if err := service.UpdateStatus(ctx, dev.TenantID, dev.ID, device.StatusOnline); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}

	return nil
}

func applyDeviceConfiguration(ctx context.Context, service *device.Service, dev *device.Device, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-config: %w", err)
	}

	config := map[string]interface{}{
		"monitoring_interval": "30s",
		"log_level":           "info",
		"features": map[string]bool{
			"metrics_enabled":  true,
			"tracing_enabled":  false,
			"alerting_enabled": true,
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := dev.SetConfig(configJSON, "admin"); err != nil {
		return fmt.Errorf("setting config: %w", err)
	}

	if err := service.Update(ctx, dev); err != nil {
		return fmt.Errorf("updating device config: %w", err)
	}

	logger.Info("applied device configuration",
		zap.String("device_id", dev.ID),
		zap.String("config_hash", dev.LastConfigHash),
	)

	if err := dev.ValidateConfig(); err != nil {
		return fmt.Errorf("validating config: %w", err)
	}

	return service.Update(ctx, dev)
}

func updateNetworkInfo(ctx context.Context, service *device.Service, dev *device.Device, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-network: %w", err)
	}

	networkInfo := &device.NetworkInfo{
		IPAddress:  "192.168.1.100",
		MACAddress: "00:11:22:33:44:55",
		Hostname:   "demo-device-1",
		Port:       9100,
	}

	if err := dev.UpdateDiscoveryInfo(device.DiscoveryManual, networkInfo); err != nil {
		return fmt.Errorf("updating discovery info: %w", err)
	}

	if err := service.Update(ctx, dev); err != nil {
		return fmt.Errorf("updating network info: %w", err)
	}

	logger.Info("updated device network info",
		zap.String("device_id", dev.ID),
		zap.String("ip", networkInfo.IPAddress),
		zap.String("hostname", networkInfo.Hostname),
	)

	return nil
}

func configureOfflineCapabilities(ctx context.Context, service *device.Service, dev *device.Device, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-offline-config: %w", err)
	}

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

	if err := dev.UpdateOfflineCapabilities(offlineCapabilities); err != nil {
		return fmt.Errorf("updating offline capabilities: %w", err)
	}

	if err := service.Update(ctx, dev); err != nil {
		return fmt.Errorf("saving offline capabilities: %w", err)
	}

	logger.Info("configured offline capabilities",
		zap.String("device_id", dev.ID),
		zap.Bool("airgap_supported", offlineCapabilities.SupportsAirgap),
		zap.Duration("sync_interval", offlineCapabilities.SyncInterval),
	)

	return nil
}

func listDevices(ctx context.Context, service *device.Service, tenantID string, logger *zap.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pre-list: %w", err)
	}

	devices, err := service.List(ctx, device.ListOptions{
		TenantID: tenantID,
		Tags: map[string]string{
			"environment": "production",
		},
	})

	if err != nil {
		return fmt.Errorf("listing devices: %w", err)
	}

	logger.Info("listed devices",
		zap.Int("count", len(devices)),
		zap.String("tenant_id", tenantID),
	)

	return nil
}
