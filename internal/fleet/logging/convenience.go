package logging

import "github.com/wrale/wrale-fleet/internal/fleet/logging/store/factory"

// NewMemoryStore creates a new in-memory store implementation.
// This is a convenience wrapper around the factory method, primarily
// intended for testing purposes.
func NewMemoryStore() Store {
	return factory.NewMemoryStore()
}
