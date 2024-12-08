package group

import (
	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group/store/memory"
)

// NewTestStore creates a new Store implementation for testing purposes.
// This is meant to be used only in tests and provides an in-memory
// implementation with proper initialization of all required components.
func NewTestStore(deviceStore device.Store) Store {
	if deviceStore == nil {
		deviceStore = device.NewTestStore()
	}
	return memory.New(deviceStore)
}

// NewMemoryStore is a legacy alias for NewTestStore to maintain
// backward compatibility with existing tests. New tests should use
// NewTestStore instead.
func NewMemoryStore() Store {
	return NewTestStore(nil)
}