package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
)

// Store provides an in-memory implementation of device.Store interface.
// It is primarily used for testing and demonstration purposes.
type Store struct {
	mu      sync.RWMutex
	devices map[string]*device.Device // key: tenantID:deviceID
}

// New creates a new in-memory device store
func New() device.Store {
	return &Store{
		devices: make(map[string]*device.Device),
	}
}

// rest of the file stays the same...
