package device

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/wrale-fleet/fleet/types"
)

// RackConfig defines physical rack constraints
type RackConfig struct {
	MaxUnits    int
	PowerLimit  float64
	CoolingZone string
}

// TopologyManager handles physical device placement and relationships
type TopologyManager struct {
	inventory   *Inventory
	rackConfigs map[string]RackConfig
	mu          sync.RWMutex
}

// NewTopologyManager creates a new topology manager
func NewTopologyManager(inventory *Inventory) *TopologyManager {
	return &TopologyManager{
		inventory:   inventory,
		rackConfigs: make(map[string]RackConfig),
	}
}

// RegisterRack adds or updates rack configuration
func (t *TopologyManager) RegisterRack(ctx context.Context, rackID string, config RackConfig) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.rackConfigs[rackID] = config
	return nil
}

// UnregisterRack removes a rack configuration
func (t *TopologyManager) UnregisterRack(ctx context.Context, rackID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.rackConfigs, rackID)
	return nil
}

// GetLocation retrieves a device's physical location
func (t *TopologyManager) GetLocation(ctx context.Context, deviceID types.DeviceID) (*types.PhysicalLocation, error) {
	state, err := t.inventory.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return &state.Location, nil
}

// UpdateLocation updates a device's physical location
func (t *TopologyManager) UpdateLocation(ctx context.Context, deviceID types.DeviceID, location types.PhysicalLocation) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Verify rack exists
	if _, exists := t.rackConfigs[location.Rack]; !exists {
		return fmt.Errorf("rack not found: %s", location.Rack)
	}

	// Get current device state
	state, err := t.inventory.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	// Update location
	state.Location = location
	return t.inventory.UpdateState(ctx, *state)
}

// GetDevicesInZone returns all devices in a cooling zone
func (t *TopologyManager) GetDevicesInZone(ctx context.Context, zone string) ([]types.DeviceState, error) {
	devices, err := t.inventory.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	zoneDevices := make([]types.DeviceState, 0)
	for _, device := range devices {
		if t.getDeviceZone(device.Location.Rack) == zone {
			zoneDevices = append(zoneDevices, device)
		}
	}
	return zoneDevices, nil
}

// GetDevicesInRack returns all devices in a rack
func (t *TopologyManager) GetDevicesInRack(ctx context.Context, rack string) ([]types.DeviceState, error) {
	devices, err := t.inventory.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	rackDevices := make([]types.DeviceState, 0)
	for _, device := range devices {
		if device.Location.Rack == rack {
			rackDevices = append(rackDevices, device)
		}
	}
	return rackDevices, nil
}

// ValidateLocation checks if a location assignment is valid
func (t *TopologyManager) ValidateLocation(ctx context.Context, location types.PhysicalLocation) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check rack exists
	config, exists := t.rackConfigs[location.Rack]
	if !exists {
		return fmt.Errorf("rack not found: %s", location.Rack)
	}

	// Check position within bounds
	if location.Position <= 0 || location.Position > config.MaxUnits {
		return fmt.Errorf("invalid rack position: %d (max: %d)", location.Position, config.MaxUnits)
	}

	return nil
}

// getDeviceZone returns the cooling zone for a rack
func (t *TopologyManager) getDeviceZone(rack string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if config, exists := t.rackConfigs[rack]; exists {
		return config.CoolingZone
	}
	return ""
}

// GetRackPowerUsage calculates total power usage for a rack
func (t *TopologyManager) GetRackPowerUsage(ctx context.Context, rack string) (float64, error) {
	devices, err := t.GetDevicesInRack(ctx, rack)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, device := range devices {
		total += device.Metrics.PowerUsage
	}
	return total, nil
}