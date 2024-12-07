// Package coordinator implements the core coordination logic for the fleet brain
package coordinator

import (
	"context"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateManager handles the fleet-wide state management
type StateManager struct {
	deviceStates map[types.DeviceID]types.DeviceState
	mu           sync.RWMutex
}

// NewStateManager creates a new state manager instance
func NewStateManager() *StateManager {
	return &StateManager{
		deviceStates: make(map[types.DeviceID]types.DeviceState),
	}
}

// GetDeviceState retrieves the current state of a device
func (sm *StateManager) GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if state, exists := sm.deviceStates[deviceID]; exists {
		return &state, nil
	}
	return nil, nil
}

// UpdateDeviceState updates the state of a device
func (sm *StateManager) UpdateDeviceState(ctx context.Context, state types.DeviceState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.LastUpdated = time.Now()
	sm.deviceStates[state.ID] = state
	return nil
}

// ListDevices returns all known devices and their states
func (sm *StateManager) ListDevices(ctx context.Context) ([]types.DeviceState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	devices := make([]types.DeviceState, 0, len(sm.deviceStates))
	for _, state := range sm.deviceStates {
		devices = append(devices, state)
	}
	return devices, nil
}

// RemoveDevice removes a device from state tracking
func (sm *StateManager) RemoveDevice(ctx context.Context, deviceID types.DeviceID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.deviceStates, deviceID)
	return nil
}