package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/types"
)

// HealthStatus represents a device's health state
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Inventory manages device registration and tracking.
type Inventory struct {
	devices map[types.DeviceID]*DeviceInfo
	mu      sync.RWMutex
}

// Verify interface implementations
var (
	_ types.StateManager  = (*Inventory)(nil)
	_ types.DeviceManager = (*Inventory)(nil)
)

// DeviceInfo extends DeviceState with additional inventory information
type DeviceInfo struct {
	State       types.DeviceState
	Health      HealthStatus
	LastContact time.Time
	RegisteredAt time.Time
}

// NewInventory creates a new device inventory
func NewInventory() *Inventory {
	return &Inventory{
		devices: make(map[types.DeviceID]*DeviceInfo),
	}
}

// GetDeviceState implements StateManager.GetDeviceState
func (i *Inventory) GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
	return i.GetDevice(ctx, deviceID)
}

// UpdateDeviceState implements StateManager.UpdateDeviceState
func (i *Inventory) UpdateDeviceState(ctx context.Context, state types.DeviceState) error {
	return i.UpdateState(ctx, state)
}

// AddDevice implements StateManager.AddDevice
func (i *Inventory) AddDevice(ctx context.Context, state types.DeviceState) error {
	return i.RegisterDevice(ctx, state)
}

// RemoveDevice implements StateManager.RemoveDevice
func (i *Inventory) RemoveDevice(ctx context.Context, deviceID types.DeviceID) error {
	return i.UnregisterDevice(ctx, deviceID)
}

// GetDevice implements DeviceManager.GetDevice
func (i *Inventory) GetDevice(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if info, exists := i.devices[deviceID]; exists {
		return &info.State, nil
	}
	return nil, fmt.Errorf("device not found: %s", deviceID)
}

// ListDevices implements both StateManager.ListDevices and DeviceManager.ListDevices
func (i *Inventory) ListDevices(ctx context.Context) ([]types.DeviceState, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	devices := make([]types.DeviceState, 0, len(i.devices))
	for _, info := range i.devices {
		devices = append(devices, info.State)
	}
	return devices, nil
}

// GetDevicesInZone implements DeviceManager.GetDevicesInZone
func (i *Inventory) GetDevicesInZone(ctx context.Context, zone string) ([]types.DeviceState, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	devices := make([]types.DeviceState, 0)
	for _, info := range i.devices {
		if info.State.Location.Zone == zone {
			devices = append(devices, info.State)
		}
	}
	return devices, nil
}

// RegisterDevice adds a new device to the inventory (deprecated: use AddDevice)
func (i *Inventory) RegisterDevice(ctx context.Context, state types.DeviceState) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	info := &DeviceInfo{
		State:        state,
		Health:       HealthStatusUnknown,
		LastContact:  now,
		RegisteredAt: now,
	}

	i.devices[state.ID] = info
	return nil
}

// UnregisterDevice removes a device from the inventory (deprecated: use RemoveDevice)
func (i *Inventory) UnregisterDevice(ctx context.Context, deviceID types.DeviceID) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.devices[deviceID]; !exists {
		return fmt.Errorf("device not found: %s", deviceID)
	}

	delete(i.devices, deviceID)
	return nil
}

// UpdateState updates a device's state and health information
func (i *Inventory) UpdateState(ctx context.Context, state types.DeviceState) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	info, exists := i.devices[state.ID]
	if !exists {
		return fmt.Errorf("device not found: %s", state.ID)
	}

	info.State = state
	info.LastContact = time.Now()
	info.Health = determineHealth(state)

	return nil
}

// GetHealthReport returns health status for all devices
func (i *Inventory) GetHealthReport(ctx context.Context) map[HealthStatus]int {
	i.mu.RLock()
	defer i.mu.RUnlock()

	report := map[HealthStatus]int{
		HealthStatusHealthy:   0,
		HealthStatusDegraded:  0,
		HealthStatusUnhealthy: 0,
		HealthStatusUnknown:   0,
	}

	for _, info := range i.devices {
		report[info.Health]++
	}

	return report
}

// determineHealth evaluates device health based on metrics
func determineHealth(state types.DeviceState) HealthStatus {
	if state.Metrics.CPULoad > 90 {
		return HealthStatusDegraded
	}
	if state.Metrics.MemoryUsage > 95 {
		return HealthStatusDegraded
	}
	if state.Metrics.Temperature > 80 {
		return HealthStatusUnhealthy
	}
	if state.Metrics.Throttled {
		return HealthStatusDegraded
	}
	
	return HealthStatusHealthy
}