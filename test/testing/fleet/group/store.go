// Package group provides test utilities for the group package
package group

import (
	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
	memstore "github.com/wrale/fleet/internal/fleet/group/store/memory"
	testdevice "github.com/wrale/fleet/test/testing/fleet/device"
)

// NewTestStore creates a new Store implementation for testing with a device store.
// If no device store is provided, a new test device store is created.
func NewTestStore(deviceStore device.Store) group.Store {
	if deviceStore == nil {
		deviceStore = testdevice.NewTestStore()
	}
	return memstore.New(deviceStore)
}

// NewMemoryStore is a legacy alias for NewTestStore for backward compatibility.
// New tests should use NewTestStore instead.
func NewMemoryStore() group.Store {
	return NewTestStore(nil)
}
