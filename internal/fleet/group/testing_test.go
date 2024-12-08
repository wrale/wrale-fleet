package group

import (
	"github.com/wrale/fleet/internal/fleet/device"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group/store/memory"
)

// newTestStore creates a new Store implementation for testing with a device store.
// If no device store is provided, a new memory device store is created.
func newTestStore(deviceStore device.Store) Store {
	if deviceStore == nil {
		deviceStore = devmem.New()
	}
	return memory.New(deviceStore)
}

// NewMemoryStore provides a memory-based store implementation for testing.
func NewMemoryStore() Store {
	return newTestStore(nil)
}
