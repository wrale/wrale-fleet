package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// DemoManager handles continuous demonstration of fleet management capabilities
// with multi-tenant isolation.
type DemoManager struct {
	service *device.Service
	logger  *zap.Logger

	// Lifecycle management
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context

	// Demo state with multi-tenant support
	tenantsMu sync.RWMutex
	tenants   map[string]*TenantDemo

	updateInterval time.Duration
}

// TenantDemo represents a tenant's demo environment and state,
// tracking devices and operations for isolation demonstrations.
type TenantDemo struct {
	TenantID     string
	Devices      map[string]*device.Device // deviceID -> device
	devicesMu    sync.RWMutex
	LastUpdateAt time.Time
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
		tenants:        make(map[string]*TenantDemo),
	}
}

// Start begins the continuous demo operations
func (dm *DemoManager) Start() error {
	// Initialize demo environments for multiple tenants
	if err := dm.initialize(); err != nil {
		return fmt.Errorf("demo initialization failed: %w", err)
	}

	dm.wg.Add(1)
	go dm.runUpdateLoop()

	return nil
}

// Stop gracefully shuts down demo operations with timeout from the provided context
func (dm *DemoManager) Stop(ctx context.Context) error {
	dm.cancel()

	done := make(chan struct{})
	go func() {
		dm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}

// initialize sets up the initial demo environments for multiple tenants
func (dm *DemoManager) initialize() error {
	// Create demo tenants with different characteristics
	tenants := []string{"tenant-production", "tenant-staging", "tenant-dev"}
	deviceCounts := []int{3, 2, 1} // Different device counts per tenant

	for i, tenantID := range tenants {
		td := &TenantDemo{
			TenantID: tenantID,
			Devices:  make(map[string]*device.Device),
		}

		// Create tenant-specific context for registration and setup
		ctx := device.ContextWithTenant(dm.ctx, tenantID)

		// Create devices for this tenant
		for j := 0; j < deviceCounts[i]; j++ {
			deviceName := fmt.Sprintf("Demo Device %d", j+1)
			dev, err := dm.service.Register(ctx, tenantID, deviceName)
			if err != nil {
				return fmt.Errorf("failed to register device for tenant %s: %w", tenantID, err)
			}

			// Setup device configuration using tenant context
			if err := dm.setupDevice(ctx, dev); err != nil {
				return fmt.Errorf("device setup failed for tenant %s: %w", tenantID, err)
			}

			td.devicesMu.Lock()
			td.Devices[dev.ID] = dev
			td.devicesMu.Unlock()

			// Log successful device creation
			dm.logger.Info("registered tenant device",
				zap.String("tenant_id", tenantID),
				zap.String("device_id", dev.ID),
				zap.String("name", deviceName))
		}

		dm.tenantsMu.Lock()
		dm.tenants[tenantID] = td
		dm.tenantsMu.Unlock()

		dm.logger.Info("initialized tenant environment",
			zap.String("tenant_id", tenantID),
			zap.Int("device_count", deviceCounts[i]))
	}

	// Initial demonstration of tenant isolation
	dm.demonstrateIsolation()

	return nil
}

// runUpdateLoop continuously updates the demo environment
func (dm *DemoManager) runUpdateLoop() {
	defer dm.wg.Done()

	ticker := time.NewTicker(dm.updateInterval)
	defer ticker.Stop()

	updateCount := 0

	for {
		select {
		case <-dm.ctx.Done():
			dm.logger.Info("demo update loop stopping")
			return
		case <-ticker.C:
			dm.tenantsMu.RLock()
			for _, td := range dm.tenants {
				if err := dm.updateTenant(td); err != nil {
					dm.logger.Error("tenant update failed",
						zap.String("tenant_id", td.TenantID),
						zap.Error(err))
					// Continue with other tenants despite errors
				}
			}
			dm.tenantsMu.RUnlock()

			// Periodically demonstrate isolation
			updateCount++
			if updateCount%5 == 0 {
				dm.demonstrateIsolation()
			}
		}
	}
}

// setupDevice performs initial device configuration
func (dm *DemoManager) setupDevice(ctx context.Context, dev *device.Device) error {
	// Set environment tag based on tenant type
	environment := "unknown"
	switch dev.TenantID {
	case "tenant-production":
		environment = "production"
	case "tenant-staging":
		environment = "staging"
	case "tenant-dev":
		environment = "development"
	}

	if err := dev.AddTag("environment", environment); err != nil {
		return fmt.Errorf("adding environment tag: %w", err)
	}

	if err := dev.AddTag("location", "datacenter-1"); err != nil {
		return fmt.Errorf("adding location tag: %w", err)
	}

	if err := dm.service.Update(ctx, dev); err != nil {
		return fmt.Errorf("updating device tags: %w", err)
	}

	if err := dm.service.UpdateStatus(ctx, dev.TenantID, dev.ID, device.StatusOnline); err != nil {
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

	return dm.service.Update(ctx, dev)
}

// updateTenant updates all devices for a specific tenant
func (dm *DemoManager) updateTenant(td *TenantDemo) error {
	td.devicesMu.RLock()
	defer td.devicesMu.RUnlock()

	// Create tenant-specific context
	ctx := device.ContextWithTenant(dm.ctx, td.TenantID)

	for _, dev := range td.Devices {
		// Update offline sync time
		if caps := dev.GetOfflineCapabilities(); caps != nil {
			caps.LastSyncTime = time.Now()
			if err := dev.UpdateOfflineCapabilities(caps); err != nil {
				return fmt.Errorf("updating sync time: %w", err)
			}
		}

		// Refresh device status
		if err := dm.service.UpdateStatus(ctx, td.TenantID, dev.ID, device.StatusOnline); err != nil {
			return fmt.Errorf("refreshing status: %w", err)
		}

		// Update device record
		if err := dm.service.Update(ctx, dev); err != nil {
			return fmt.Errorf("updating device: %w", err)
		}

		dm.logger.Info("device update completed",
			zap.String("tenant_id", td.TenantID),
			zap.String("device_id", dev.ID),
			zap.Time("updated_at", time.Now()))
	}

	td.LastUpdateAt = time.Now()
	return nil
}

// demonstrateIsolation showcases tenant isolation by attempting cross-tenant operations
func (dm *DemoManager) demonstrateIsolation() {
	dm.tenantsMu.RLock()
	defer dm.tenantsMu.RUnlock()

	// Get devices from different tenants for isolation testing
	type testCase struct {
		sourceTenant string
		targetTenant string
		device       *device.Device
	}

	var tests []testCase

	// Build test cases for cross-tenant attempts
	for _, srcTD := range dm.tenants {
		srcTD.devicesMu.RLock()
		for _, targetTD := range dm.tenants {
			if srcTD.TenantID == targetTD.TenantID {
				continue // Skip same tenant
			}

			targetTD.devicesMu.RLock()
			for _, dev := range targetTD.Devices {
				tests = append(tests, testCase{
					sourceTenant: srcTD.TenantID,
					targetTenant: targetTD.TenantID,
					device:       dev,
				})
				break // One device per tenant pair is sufficient
			}
			targetTD.devicesMu.RUnlock()
		}
		srcTD.devicesMu.RUnlock()
	}

	// Execute isolation tests
	for _, tc := range tests {
		// Create context with incorrect tenant
		ctx := device.ContextWithTenant(dm.ctx, tc.sourceTenant)

		// Attempt unauthorized access
		err := dm.service.UpdateStatus(ctx, tc.targetTenant, tc.device.ID, device.StatusOnline)
		if err != nil {
			// Expected error - log successful isolation
			dm.logger.Info("tenant isolation enforced",
				zap.String("source_tenant", tc.sourceTenant),
				zap.String("target_tenant", tc.targetTenant),
				zap.String("device_id", tc.device.ID),
				zap.String("operation", "status_update"),
				zap.Error(err))
		} else {
			// Unexpected success - this should never happen
			dm.logger.Error("tenant isolation failure detected",
				zap.String("source_tenant", tc.sourceTenant),
				zap.String("target_tenant", tc.targetTenant),
				zap.String("device_id", tc.device.ID),
				zap.String("operation", "status_update"))
		}

		// Also try device retrieval
		_, err = dm.service.Get(ctx, tc.targetTenant, tc.device.ID)
		if err != nil {
			dm.logger.Info("tenant isolation enforced",
				zap.String("source_tenant", tc.sourceTenant),
				zap.String("target_tenant", tc.targetTenant),
				zap.String("device_id", tc.device.ID),
				zap.String("operation", "device_get"),
				zap.Error(err))
		}
	}
}
