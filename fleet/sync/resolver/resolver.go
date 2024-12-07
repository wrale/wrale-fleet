package resolver

import (
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

// Resolver handles state conflicts and versioning
type Resolver struct {
	mu sync.RWMutex
	timeout time.Duration
}

// NewResolver creates a new resolver instance
func NewResolver(timeoutSeconds int) *Resolver {
	return &Resolver{
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// ResolveStateConflict resolves conflicts between states
func (r *Resolver) ResolveStateConflict(states []types.VersionedState) (*types.VersionedState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(states) == 0 {
		return nil, nil
	}

	// Find the most recent valid state
	var latest *types.VersionedState
	for i := range states {
		state := &states[i]
		if latest == nil || state.Version > latest.Version {
			latest = state
		}
	}

	return latest, nil
}

// ValidateState checks if a state is valid
func (r *Resolver) ValidateState(state *types.VersionedState) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if state == nil {
		return false
	}

	// Check if state is not too old
	if time.Since(time.Unix(state.Timestamp, 0)) > r.timeout {
		return false
	}

	return true
}
