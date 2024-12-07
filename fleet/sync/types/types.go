package types

import (
	"time"
	
	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// StateVersion represents a version number for state
type StateVersion int64

// VersionedState represents a versioned device state
type VersionedState struct {
	Version     StateVersion        `json:"version"`
	Timestamp   int64              `json:"timestamp"`
	State       types.DeviceState  `json:"state"`
	UpdatedAt   time.Time          `json:"updated_at"`
	UpdatedBy   string             `json:"updated_by"`
	ValidatedBy []string           `json:"validated_by,omitempty"`
}