package types

import (
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateVersion represents a version number for state
type StateVersion string

// DeviceID represents a unique device identifier
type DeviceID string

// ConsensusStatus represents the status of state consensus
type ConsensusStatus string

const (
	// ConsensusAchieved indicates all nodes agree on state
	ConsensusAchieved ConsensusStatus = "achieved"
	// ConsensusPending indicates consensus is being negotiated
	ConsensusPending ConsensusStatus = "pending"
	// ConsensusConflict indicates there are conflicts to resolve
	ConsensusConflict ConsensusStatus = "conflict"
)

// SyncOperation represents a sync operation type
type SyncOperation string

const (
	// SyncPush indicates pushing state to devices
	SyncPush SyncOperation = "push"
	// SyncPull indicates pulling state from devices
	SyncPull SyncOperation = "pull"
	// SyncMerge indicates merging conflicting states
	SyncMerge SyncOperation = "merge"
)

// StateChange represents a change in device state
type StateChange struct {
	DeviceID  DeviceID         `json:"device_id"`
	OldState  types.DeviceState `json:"old_state,omitempty"`
	NewState  types.DeviceState `json:"new_state"`
	Timestamp time.Time        `json:"timestamp"`
	Source    string          `json:"source"`
}

// ConfigData represents device configuration data
type ConfigData struct {
	Version     string                 `json:"version"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
	Policies    map[string]interface{} `json:"policies"`
	Constraints map[string]interface{} `json:"constraints"`
}

// StateStore defines interface for state storage and retrieval
type StateStore interface {
	GetState(version StateVersion) (types.DeviceState, error)
	SetState(state types.DeviceState) error
	GetHistory(limit int) ([]StateChange, error)
	GetVersion() StateVersion
}

// ConflictResolver defines interface for resolving state conflicts
type ConflictResolver interface {
	DetectConflicts(states []types.DeviceState) bool
	ResolveConflicts(states []types.DeviceState) (types.DeviceState, error)
	ValidateState(state types.DeviceState) error
}

// ConfigManager defines interface for configuration management
type ConfigManager interface {
	GetConfig(deviceID DeviceID) (*ConfigData, error)
	UpdateConfig(config *ConfigData) error
	ValidateConfig(config *ConfigData) error
	DistributeConfig(config *ConfigData, devices []DeviceID) error
}