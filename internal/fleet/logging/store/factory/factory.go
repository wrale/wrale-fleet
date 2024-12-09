// Package factory provides creation functions for logging store implementations.
package factory

import (
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
)

// NewMemoryStore creates a new in-memory logging store.
// This is primarily used for testing and development purposes.
func NewMemoryStore() logging.Store {
	return memory.New()
}
