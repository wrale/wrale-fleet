// Package memory provides an in-memory implementation of the health store interface.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/wrale-fleet/internal/fleet/health"
)

// Store provides an in-memory implementation of the health.Store interface.
// This implementation is primarily intended for testing and development.
type Store struct {
	mu         sync.RWMutex
	components map[string]health.ComponentInfo
	statuses   map[string]*health.HealthStatus
	ready      bool
}

// New creates a new in-memory health store.
func New(opts ...health.StoreOption) *Store {
	s := &Store{
		components: make(map[string]health.ComponentInfo),
		statuses:   make(map[string]*health.HealthStatus),
	}

	// Apply options - since this is a memory implementation,
	// we ignore errors as options primarily affect persistence
	options := &health.StoreOptions{}
	for _, opt := range opts {
		_ = opt(options)
	}

	return s
}

// UpdateComponentStatus updates the health status for a specific component
func (s *Store) UpdateComponentStatus(ctx context.Context, component string, status *health.HealthStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.components[component]; !exists {
		return fmt.Errorf("component %s not registered", component)
	}

	s.statuses[component] = status
	return nil
}

// GetComponentStatus retrieves the health status for a specific component
func (s *Store) GetComponentStatus(ctx context.Context, component string) (*health.HealthStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statuses[component]
	if !exists {
		return nil, fmt.Errorf("no status found for component %s", component)
	}

	return status, nil
}

// ListComponentStatuses retrieves health status for all components
func (s *Store) ListComponentStatuses(ctx context.Context) (map[string]*health.HealthStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy to prevent map mutations
	statuses := make(map[string]*health.HealthStatus, len(s.statuses))
	for k, v := range s.statuses {
		copied := *v
		statuses[k] = &copied
	}

	return statuses, nil
}

// GetReadyStatus retrieves the current ready status
func (s *Store) GetReadyStatus(ctx context.Context) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ready, nil
}

// SetReadyStatus updates the ready status
func (s *Store) SetReadyStatus(ctx context.Context, ready bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ready = ready
	return nil
}

// RegisterComponent registers a new component for health monitoring
func (s *Store) RegisterComponent(ctx context.Context, component string, info health.ComponentInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.components[component]; exists {
		return fmt.Errorf("component %s already registered", component)
	}

	s.components[component] = info
	return nil
}

// UnregisterComponent removes a component from health monitoring
func (s *Store) UnregisterComponent(ctx context.Context, component string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.components, component)
	delete(s.statuses, component)
	return nil
}
