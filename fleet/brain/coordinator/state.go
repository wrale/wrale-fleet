// Package coordinator implements the core coordination logic for the fleet brain
package coordinator

import (
	"context"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateManager defines the interface for managing device state
type StateManager interface {
	GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error)
	UpdateDeviceState(ctx context.Context, state types.DeviceState) error
	ListDevices(ctx context.Context) ([]types.DeviceState, error)
	RemoveDevice(ctx context.Context, deviceID types.DeviceID) error
	AddDevice(ctx context.Context, state types.DeviceState) error
}

// DefaultStateManager provides basic state management implementation
type DefaultStateManager struct {
	deviceStates map[types.DeviceID]types.DeviceState
	mu           sync.RWMutex
}

// NewStateManager creates a new state manager instance
func NewStateManager() *DefaultStateManager {
	return &DefaultStateManager{
		deviceStates: make(map[types.DeviceID]types.DeviceState),
	}
}

// GetDeviceState retrieves the current state of a device
func (sm *DefaultStateManager) GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if state, exists := sm.deviceStates[deviceID]; exists {
		return &state, nil
	}
	return nil, nil
}

// UpdateDeviceState updates the state of a device
func (sm *DefaultStateManager) UpdateDeviceState(ctx context.Context, state types.DeviceState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.LastUpdated = time.Now()
	sm.deviceStates[state.ID] = state
	return nil
}

// ListDevices returns all known devices and their states
func (sm *DefaultStateManager) ListDevices(ctx context.Context) ([]types.DeviceState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	devices := make([]types.DeviceState, 0, len(sm.deviceStates))
	for _, state := range sm.deviceStates {
		devices = append(devices, state)
	}
	return devices, nil
}

// RemoveDevice removes a device from state tracking
func (sm *DefaultStateManager) RemoveDevice(ctx context.Context, deviceID types.DeviceID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.deviceStates, deviceID)
	return nil
}

// AddDevice adds a new device to state tracking
func (sm *DefaultStateManager) AddDevice(ctx context.Context, state types.DeviceState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state.LastUpdated = time.Now()
	sm.deviceStates[state.ID] = state
	return nil
}