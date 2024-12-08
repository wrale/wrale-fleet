package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// DemoManager handles continuous demonstration of fleet management capabilities.
// It maintains demo state and performs periodic updates to showcase features.
type DemoManager struct {
	service *device.Service
	logger  *zap.Logger

	// Lifecycle management
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context

	// Demo state
	demoDeviceMu sync.RWMutex
	demoDevice   *device.Device

	updateInterval time.Duration
}

// NewDemoManager creates a new demo manager instance configured for continuous operation
func NewDemoManager(service *device.Service, logger *zap.Logger) *DemoManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DemoManager{
		service:        service,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
		updateInterval: time.Minute,
	}
}

// Start begins the continuous demo operations
func (dm *DemoManager) Start() error {
	// Initialize demo environment
	if err := dm.initialize(); err != nil {
		return fmt.Errorf("demo initialization failed: %w", err)
	}

	dm.wg.Add(1)
	go dm.runUpdateLoop()

	return nil
}

// Stop gracefully shuts down demo operations
func (dm *DemoManager) Stop() error {
	dm.cancel()
	dm.wg.Wait()
	return nil
}

// initialize sets up the initial demo environment
func (dm *DemoManager) initialize() error {
	const tenantID = "demo-tenant"

	// Register demo device if it doesn't exist
	dev, err := dm.service.Register(dm.ctx, tenantID, "Demo Raspberry Pi")
	if err != nil {
		return fmt.Errorf("device registration: %w", err)
	}

	dm.logger.Info("registered demo device",
		zap.String("device_id", dev.ID),
		zap.String("tenant_id", dev.TenantID),
		zap.String("name", dev.Name))

	// Store device reference
	dm.demoDeviceMu.Lock()
	dm.demoDevice = dev
	dm.demoDeviceMu.Unlock()

	// Perform initial setup
	if err := dm.setupDevice(dev); err != nil {
		return fmt.Errorf("device setup: %w", err)
	}

	return nil
}

// runUpdateLoop continuously updates the demo environment
func (dm *DemoManager) runUpdateLoop() {
	defer dm.wg.Done()

	ticker := time.NewTicker(dm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			dm.logger.Info("demo update loop stopping")
			return
		case <-ticker.C:
			if err := dm.performUpdate(); err != nil {
				dm.logger.Error("demo update failed", zap.Error(err))
				// Continue running despite errors
			}
		}
	}
}

// setupDevice performs initial device configuration
func (dm *DemoManager) setupDevice(dev *device.Device) error {
	// Add production environment tag
	if err := dev.AddTag("environment", "production"); err != nil {
		return fmt.Errorf("adding environment tag: %w", err)
	}

	// Add datacenter location
	if err := dev.AddTag("location", "datacenter-1"); err != nil {
		return fmt.Errorf("adding location tag: %w", err)
	}

	// Update device with tags
	if err := dm.service.Update(dm.ctx, dev); err != nil {
		return fmt.Errorf("updating device tags: %w", err)
	}

	// Set device status to online
	if err := dm.service.UpdateStatus(dm.ctx, dev.TenantID, dev.ID, device.StatusOnline); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}

	// Configure offline capabilities
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

	return dm.service.Update(dm.ctx, dev)
}

// performUpdate demonstrates ongoing management capabilities
func (dm *DemoManager) performUpdate() error {
	dm.demoDeviceMu.RLock()
	dev := dm.demoDevice
	dm.demoDeviceMu.RUnlock()

	if dev == nil {
		return fmt.Errorf("demo device not initialized")
	}

	// Update offline sync time
	if caps := dev.GetOfflineCapabilities(); caps != nil {
		caps.LastSyncTime = time.Now()
		if err := dev.UpdateOfflineCapabilities(caps); err != nil {
			return fmt.Errorf("updating sync time: %w", err)
		}
	}

	// Refresh device status
	if err := dm.service.UpdateStatus(dm.ctx, dev.TenantID, dev.ID, device.StatusOnline); err != nil {
		return fmt.Errorf("refreshing status: %w", err)
	}

	// Update device record
	if err := dm.service.Update(dm.ctx, dev); err != nil {
		return fmt.Errorf("updating device: %w", err)
	}

	dm.logger.Info("demo update completed",
		zap.String("device_id", dev.ID),
		zap.String("tenant_id", dev.TenantID),
		zap.Time("updated_at", time.Now()))

	return nil
}
