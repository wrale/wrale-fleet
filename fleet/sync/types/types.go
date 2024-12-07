package types

import (
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateVersion represents a version number for state
type StateVersion string

// DeviceID represents a unique device identifier
type DeviceID = types.DeviceID

// ConsensusStatus represents the status of state consensus
type ConsensusStatus struct {
	Version       StateVersion `json:"version"`
	Validators    []string     `json:"validators"`
	Threshold     int          `json:"threshold"`
	Confirmations int          `json:"confirmations"`
	ReachedAt     *time.Time   `json:"reached_at,omitempty"`
}

// SyncOperation represents a sync operation
type SyncOperation struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	// SyncPush indicates pushing state to devices
	SyncPush = "push"
	// SyncPull indicates pulling state from devices
	SyncPull = "pull"
	// SyncMerge indicates merging conflicting states
	SyncMerge = "merge"
)

// VersionedState represents a versioned device state
type VersionedState struct {
	Version   StateVersion      `json:"version"`
	State     types.DeviceState `json:"state"`
	Timestamp time.Time         `json:"timestamp"`
	UpdatedAt time.Time         `json:"updated_at"`
	UpdatedBy string            `json:"updated_by"`
	Source    string            `json:"source"`
}

// StateChange represents a change in device state
type StateChange struct {
	DeviceID    DeviceID           `json:"device_id"`
	PrevVersion StateVersion       `json:"prev_version,omitempty"`
	NewVersion  StateVersion       `json:"new_version"`
	OldState    *types.DeviceState `json:"old_state,omitempty"`
	NewState    types.DeviceState  `json:"new_state"`
	Timestamp   time.Time          `json:"timestamp"`
	Source      string             `json:"source"`
	Changes     []string           `json:"changes,omitempty"`
}

// ConfigData represents device configuration data
type ConfigData struct {
	Version     string                 `json:"version"`
	Config      map[string]interface{} `json:"config"`
	ValidFrom   time.Time              `json:"valid_from"`
	ValidTo     *time.Time             `json:"valid_to,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Settings    map[string]interface{} `json:"settings"`
	Policies    map[string]interface{} `json:"policies"`
	Constraints map[string]interface{} `json:"constraints"`
}

// StateStore defines interface for state storage and retrieval
type StateStore interface {
	GetState(version StateVersion) (*VersionedState, error)
	SetState(deviceID DeviceID, state types.DeviceState) error
	SaveState(state *VersionedState) error
	ListVersions() ([]StateVersion, error)
	GetHistory(limit int) ([]StateChange, error)
	GetVersion() StateVersion
}

// ConflictResolver defines interface for resolving state conflicts
type ConflictResolver interface {
	DetectConflicts(states []*VersionedState) ([]*VersionedState, error)
	ResolveConflicts(states []*VersionedState) (*VersionedState, error)
	ValidateState(state *VersionedState) error
	ValidateResolution(state *VersionedState) error
}

// ConfigManager defines interface for configuration management
type ConfigManager interface {
	GetConfig(deviceID DeviceID) (*ConfigData, error)
	UpdateConfig(config *ConfigData) error
	ValidateConfig(config *ConfigData) error
	DistributeConfig(config *ConfigData, devices []DeviceID) error
	GetDeviceConfig(deviceID DeviceID) (*ConfigData, error)
}
