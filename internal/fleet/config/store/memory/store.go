package memory

import (
	"fmt"
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

// clearStore resets the store to its initial state.
// This is primarily used for testing to ensure a clean state between test cases.
func (s *Store) clearStore() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.templates = make(map[string]*config.Template)
	s.versions = make(map[string][]*config.Version)
	s.deployments = make(map[string]*config.Deployment)
}

// validateInput checks that required string fields are not empty
func (s *Store) validateInput(op string, fields map[string]string) error {
	for field, value := range fields {
		if value == "" {
			return config.NewError(op, config.ErrInvalidInput, fmt.Sprintf("%s is required", field))
		}
	}
	return nil
}

// applyPagination calculates the correct slice bounds based on offset and limit
func (s *Store) applyPagination(total int, opts config.ListOptions) (start, end int) {
	// If no limit specified, return all items
	if opts.Limit <= 0 {
		return 0, total
	}

	// Calculate start index
	start = opts.Offset
	if start >= total {
		return 0, 0
	}

	// Calculate end index
	end = start + opts.Limit
	if end > total {
		end = total
	}

	return start, end
}

// templateKey generates a composite key for template storage
func (s *Store) templateKey(tenantID, templateID string) string {
	return fmt.Sprintf("%s/%s", tenantID, templateID)
}

// deploymentKey generates a composite key for deployment storage
func (s *Store) deploymentKey(tenantID, deploymentID string) string {
	return fmt.Sprintf("%s/%s", tenantID, deploymentID)
}
