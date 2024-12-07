package types

// DeviceID represents a unique identifier for a device
type DeviceID string

// VersionedState represents a state with version information
type VersionedState struct {
	Version int64
	State   map[string]interface{}
}

// StateChange represents a change in state
type StateChange struct {
	DeviceID      DeviceID
	OldState      VersionedState
	NewState      VersionedState
	ChangeType    string
	Timestamp     int64
	ConflictState bool
}
