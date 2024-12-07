package manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

// Manager implements the sync manager functionality
type Manager struct {
	store    synctypes.StateStore
	resolver synctypes.ConflictResolver
	config   synctypes.ConfigManager

	// Track active operations
	operations map[string]*synctypes.SyncOperation
	opLock     sync.RWMutex

	// Track consensus status
	consensus map[synctypes.StateVersion]*synctypes.ConsensusStatus
	consLock  sync.RWMutex
}

// NewManager creates a new sync manager instance
func NewManager(
	store synctypes.StateStore,
	resolver synctypes.ConflictResolver,
	config synctypes.ConfigManager,
) *Manager {
	return &Manager{
		store:      store,
		resolver:   resolver,
		config:     config,
		operations: make(map[string]*synctypes.SyncOperation),
		consensus:  make(map[synctypes.StateVersion]*synctypes.ConsensusStatus),
	}
}

// GetState retrieves versioned device state
func (m *Manager) GetState(deviceID types.DeviceID) (*synctypes.VersionedState, error) {
	versions, err := m.store.ListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	var latest synctypes.StateVersion
	for _, version := range versions {
		state, err := m.store.GetState(version)
		if err != nil {
			continue
		}
		if state.State.ID == deviceID {
			latest = version
		}
	}

	if latest == "" {
		return nil, fmt.Errorf("no state found for device %s", deviceID)
	}

	return m.store.GetState(latest)
}

// UpdateState updates device state with version tracking
func (m *Manager) UpdateState(deviceID types.DeviceID, state *synctypes.VersionedState) error {
	if state.State.ID != deviceID {
		return fmt.Errorf("state device ID mismatch")
	}

	current, err := m.GetState(deviceID)
	if err == nil {
		states := []*synctypes.VersionedState{current, state}
		conflicts, err := m.resolver.DetectConflicts(states)
		if err != nil {
			return fmt.Errorf("failed to detect conflicts: %w", err)
		}
		if len(conflicts) > 0 {
			resolved, err := m.resolver.ResolveConflicts(conflicts)
			if err != nil {
				return fmt.Errorf("failed to resolve conflicts: %w", err)
			}
			state = resolved
		}
	}

	if err := m.store.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	m.consLock.Lock()
	m.consensus[state.Version] = &synctypes.ConsensusStatus{
		Version:    state.Version,
		Validators: make([]string, 0),
		Threshold:  3, // Simple majority for v1.0
	}
	m.consLock.Unlock()

	return nil
}

// ValidateState validates a state version
func (m *Manager) ValidateState(version synctypes.StateVersion) error {
	state, err := m.store.GetState(version)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}

	if err := m.resolver.ValidateResolution(state); err != nil {
		return fmt.Errorf("state validation failed: %w", err)
	}

	return nil
}

// UpdateConfig updates configuration
func (m *Manager) UpdateConfig(config *synctypes.ConfigData) error {
	if err := m.config.UpdateConfig(config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	return nil
}

// DistributeConfig distributes configuration to devices
func (m *Manager) DistributeConfig(config *synctypes.ConfigData, devices []types.DeviceID) error {
	if err := m.config.DistributeConfig(config, devices); err != nil {
		return fmt.Errorf("failed to distribute config: %w", err)
	}
	return nil
}

// GetDeviceConfig gets configuration for a specific device
func (m *Manager) GetDeviceConfig(deviceID types.DeviceID) (*synctypes.ConfigData, error) {
	config, err := m.config.GetDeviceConfig(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device config: %w", err)
	}
	return config, nil
}

// CreateOperation creates a new sync operation
func (m *Manager) CreateOperation(op *synctypes.SyncOperation) error {
	m.opLock.Lock()
	defer m.opLock.Unlock()

	if op.ID == "" || op.Type == "" {
		return fmt.Errorf("invalid operation")
	}

	op.CreatedAt = time.Now()
	op.Status = "pending"
	m.operations[op.ID] = op

	return nil
}

// GetOperation retrieves an operation by ID
func (m *Manager) GetOperation(id string) (*synctypes.SyncOperation, error) {
	m.opLock.RLock()
	defer m.opLock.RUnlock()

	op, exists := m.operations[id]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", id)
	}

	return op, nil
}

// ListOperations returns all sync operations
func (m *Manager) ListOperations() ([]*synctypes.SyncOperation, error) {
	m.opLock.RLock()
	defer m.opLock.RUnlock()

	ops := make([]*synctypes.SyncOperation, 0, len(m.operations))
	for _, op := range m.operations {
		ops = append(ops, op)
	}

	return ops, nil
}

// GetConsensus gets consensus status for a version
func (m *Manager) GetConsensus(version synctypes.StateVersion) (*synctypes.ConsensusStatus, error) {
	m.consLock.RLock()
	defer m.consLock.RUnlock()

	status, exists := m.consensus[version]
	if !exists {
		return nil, fmt.Errorf("no consensus tracking for version %s", version)
	}

	return status, nil
}

// AddValidation adds a validator to consensus tracking
func (m *Manager) AddValidation(version synctypes.StateVersion, validator string) error {
	m.consLock.Lock()
	defer m.consLock.Unlock()

	status, exists := m.consensus[version]
	if !exists {
		return fmt.Errorf("no consensus tracking for version %s", version)
	}

	for _, v := range status.Validators {
		if v == validator {
			return nil
		}
	}

	status.Validators = append(status.Validators, validator)
	status.Confirmations++

	if status.Confirmations >= status.Threshold {
		now := time.Now()
		status.ReachedAt = &now
	}

	return nil
}
