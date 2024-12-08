package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
)

// Store provides an in-memory implementation of group.Store interface.
// It is primarily used for testing and demonstration purposes.
type Store struct {
	mu          sync.RWMutex
	groups      map[string]*group.Group        // key: tenantID:groupID
	memberships map[string]map[string]struct{} // key: tenantID:groupID -> map[deviceID]struct{}
	deviceStore device.Store
}

// New creates a new in-memory group store
func New(deviceStore device.Store) *Store {
	return &Store{
		groups:      make(map[string]*group.Group),
		memberships: make(map[string]map[string]struct{}),
		deviceStore: deviceStore,
	}
}

// key generates the map key for a group
func (s *Store) key(tenantID, groupID string) string {
	return fmt.Sprintf("%s:%s", tenantID, groupID)
}

// Create stores a new group
func (s *Store) Create(ctx context.Context, g *group.Group) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("validate group: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(g.TenantID, g.ID)
	if _, exists := s.groups[key]; exists {
		return group.ErrGroupExists
	}

	// Store a copy to prevent external modifications
	copy := *g
	s.groups[key] = &copy

	// Initialize empty membership set for the group
	s.memberships[key] = make(map[string]struct{})

	return nil
}

// Get retrieves a group by ID
func (s *Store) Get(ctx context.Context, tenantID, groupID string) (*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	// Return a copy to prevent external modifications
	copy := *g
	return &copy, nil
}

// Update modifies an existing group
func (s *Store) Update(ctx context.Context, g *group.Group) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("validate group: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(g.TenantID, g.ID)
	if _, exists := s.groups[key]; !exists {
		return group.ErrGroupNotFound
	}

	// Store a copy to prevent external modifications
	copy := *g
	s.groups[key] = &copy

	return nil
}

// Delete removes a group
func (s *Store) Delete(ctx context.Context, tenantID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	if _, exists := s.groups[key]; !exists {
		return group.ErrGroupNotFound
	}

	delete(s.groups, key)
	delete(s.memberships, key)
	return nil
}

// List retrieves groups matching the given options
func (s *Store) List(ctx context.Context, opts group.ListOptions) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*group.Group

	for _, g := range s.groups {
		if !s.matchesFilter(g, opts) {
			continue
		}

		// Add a copy to prevent external modifications
		copy := *g
		result = append(result, &copy)
	}

	// Apply pagination if specified
	if opts.Limit > 0 {
		start := opts.Offset
		if start > len(result) {
			start = len(result)
		}
		end := start + opts.Limit
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, nil
}
