package types

import (
	"time"
)

// DeviceID represents a unique identifier for a device
type DeviceID string

// VersionedState represents a state with version information
type VersionedState struct {
	Version    string                 `json:"version"`
	State      map[string]interface{} `json:"state"`
	Timestamp  time.Time              `json:"timestamp"`
	UpdatedAt  time.Time              `json:"updated_at"`
	UpdatedBy  string                 `json:"updated_by"`
	Source     string                 `json:"source"`
}

// StateChange represents a change in state
type StateChange struct {
	DeviceID      DeviceID       `json:"device_id"`
	OldState      *VersionedState `json:"old_state,omitempty"`
	NewState      VersionedState  `json:"new_state"`
	ChangeType    string         `json:"change_type"`
	Timestamp     time.Time       `json:"timestamp"`
	ConflictState bool           `json:"conflict_state"`
	Changes       []string       `json:"changes,omitempty"`
}