// Package device provides test utilities for the device package
package device

import (
	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
)

// NewTestStore creates a new Store implementation for testing purposes.
func NewTestStore() device.Store {
	return memory.New()
}

// NewMemoryStore is a legacy alias for NewTestStore for backward compatibility.
// New tests should use NewTestStore instead.
func NewMemoryStore() device.Store {
	return NewTestStore()
}
