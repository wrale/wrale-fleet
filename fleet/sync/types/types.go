// Package types defines core types for fleet synchronization
package types

import (
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateVersion represents a version of device state
type StateVersion string

// Operation represents a sync operation type
type Operation string

const (
	OpStateSync    Operation = "state_sync"
	OpConfigSync   Operation = "config_sync"
	OpPolicySync   Operation = "policy_sync"
	OpResourceSync Operation = "resource_sync"
)

// VersionedState represents a versioned device state
type VersionedState struct {
	Version     StateVersion
	State       types.DeviceState
	UpdatedAt   time.Time
	UpdatedBy   string
	ValidatedBy []string
}

// StateChange represents a change in device state
type StateChange struct {
	PrevVersion StateVersion
	NewVersion  StateVersion
	Changes     map[string]interface{}
	Timestamp   time.Time
	Source      string
}

// SyncOperation represents a synchronization operation
type SyncOperation struct {
	ID          string
	Type        Operation
	DeviceIDs   []types.DeviceID
	Payload     interface{}
	Priority    int
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
	Error       error
}

// ConsensusStatus tracks state consensus
type ConsensusStatus struct {
	Version       StateVersion
	Validators    []string
	Confirmations int
	Threshold     int
	ReachedAt     *time.Time
}

// SyncError represents a synchronization error
type SyncError struct {
	Operation  *SyncOperation
	Error      error
	Device     types.DeviceID
	Timestamp  time.Time
	RetryCount int
}

// ConfigData represents configuration data
type ConfigData struct {
	Version   string
	Config    map[string]interface{}
	ValidFrom time.Time
	ValidTo   *time.Time
	Signature string
}

// SyncManager handles state synchronization
type SyncManager interface {
	// State operations
	GetState(deviceID types.DeviceID) (*VersionedState, error)
	UpdateState(deviceID types.DeviceID, state *VersionedState) error
	ValidateState(version StateVersion) error

	// Operation management
	CreateOperation(op *SyncOperation) error
	GetOperation(id string) (*SyncOperation, error)
	ListOperations() ([]*SyncOperation, error)

	// Consensus operations
	GetConsensus(version StateVersion) (*ConsensusStatus, error)
	AddValidation(version StateVersion, validator string) error
}

// StateStore handles state persistence
type StateStore interface {
	// State management
	GetState(version StateVersion) (*VersionedState, error)
	SaveState(state *VersionedState) error
	ListVersions() ([]StateVersion, error)

	// Change tracking
	TrackChange(change *StateChange) error
	GetChanges(since time.Time) ([]*StateChange, error)
}

// ConflictResolver handles state conflicts
type ConflictResolver interface {
	// Conflict detection
	DetectConflicts(states []*VersionedState) ([]*StateChange, error)

	// Resolution
	ResolveConflicts(changes []*StateChange) (*VersionedState, error)
	ValidateResolution(state *VersionedState) error
}

// ConfigManager handles configuration distribution
type ConfigManager interface {
	// Config management
	GetConfig(version string) (*ConfigData, error)
	UpdateConfig(config *ConfigData) error
	ListConfigs() ([]*ConfigData, error)

	// Distribution
	DistributeConfig(config *ConfigData, devices []types.DeviceID) error
	GetDeviceConfig(deviceID types.DeviceID) (*ConfigData, error)
}
