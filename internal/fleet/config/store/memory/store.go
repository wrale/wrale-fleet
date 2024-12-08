package memory

import (
	"sync"

	"github.com/wrale/fleet/internal/fleet/config"
)

// Store implements an in-memory configuration store with thread-safe operations.
// It ensures proper validation, consistent ordering, and safe concurrent access.
type Store struct {
	mu          sync.RWMutex
	templates   map[string]*config.Template   // key: tenantID/templateID
	versions    map[string][]*config.Version  // key: tenantID/templateID
	deployments map[string]*config.Deployment // key: tenantID/deploymentID
}

// New creates a new in-memory configuration store with initialized maps.
func New() *Store {
	return &Store{
		templates:   make(map[string]*config.Template),
		versions:    make(map[string][]*config.Version),
		deployments: make(map[string]*config.Deployment),
	}
}

// clearStore resets the store to its initial state.
// This is primarily used for testing to ensure a clean state between test cases.
func (s *Store) clearStore() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.templates = make(map[string]*config.Template)
	s.versions = make(map[string][]*config.Version)
	s.deployments = make(map[string]*config.Deployment)
}
