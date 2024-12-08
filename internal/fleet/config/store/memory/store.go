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

// New creates a new in-memory configuration store with initialized maps
func New() *Store {
	return &Store{
		templates:   make(map[string]*config.Template),
		versions:    make(map[string][]*config.Version),
		deployments: make(map[string]*config.Deployment),
	}
}
