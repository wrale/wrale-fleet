package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
)

// Store implements an in-memory group store
type Store struct {
	mu          sync.RWMutex
	groups      map[string]map[string]*group.Group // tenant -> id -> group
	memberships map[string]map[string]struct{}     // groupKey -> deviceID -> struct{}
	deviceStore device.Store                       // Device store for membership queries
}

// New creates a new memory store instance
func New(deviceStore device.Store) *Store {
	return &Store{
		groups:      make(map[string]map[string]*group.Group),
		memberships: make(map[string]map[string]struct{}),
		deviceStore: deviceStore,
	}
}

// groupKey generates a unique key for group operations
func (s *Store) groupKey(tenantID, groupID string) string {
	return fmt.Sprintf("%s:%s", tenantID, groupID)
}

// Create adds a new group
func (s *Store) Create(ctx context.Context, g *group.Group) error {
	if err := g.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize tenant map if needed
	if _, exists := s.groups[g.TenantID]; !exists {
		s.groups[g.TenantID] = make(map[string]*group.Group)
	}

	// Check for duplicate
	if _, exists := s.groups[g.TenantID][g.ID]; exists {
		return group.ErrGroupExists
	}

	// Store copy of group
	s.groups[g.TenantID][g.ID] = g.DeepCopy()

	// Initialize membership tracking
	key := s.groupKey(g.TenantID, g.ID)
	s.memberships[key] = make(map[string]struct{})

	return nil
}

// Get retrieves a group by ID and tenant
func (s *Store) Get(ctx context.Context, tenantID, id string) (*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	g, exists := tenantGroups[id]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	return g.DeepCopy(), nil
}

// Update modifies an existing group
func (s *Store) Update(ctx context.Context, g *group.Group) error {
	if err := g.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tenantGroups, exists := s.groups[g.TenantID]
	if !exists {
		return group.ErrGroupNotFound
	}

	if _, exists := tenantGroups[g.ID]; !exists {
		return group.ErrGroupNotFound
	}

	s.groups[g.TenantID][g.ID] = g.DeepCopy()
	return nil
}

// Delete removes a group
func (s *Store) Delete(ctx context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return group.ErrGroupNotFound
	}

	if _, exists := tenantGroups[id]; !exists {
		return group.ErrGroupNotFound
	}

	// Clean up memberships
	key := s.groupKey(tenantID, id)
	delete(s.memberships, key)

	// Delete group
	delete(s.groups[tenantID], id)
	return nil
}

// List returns groups matching the query criteria
func (s *Store) List(ctx context.Context, tenantID string, opts group.ListOptions) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return []*group.Group{}, nil
	}

	var result []*group.Group
	for _, g := range tenantGroups {
		if matchesListOptions(g, opts) {
			result = append(result, g.DeepCopy())
		}
	}

	return result, nil
}

// Clear removes all groups (used for testing)
func (s *Store) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups = make(map[string]map[string]*group.Group)
	s.memberships = make(map[string]map[string]struct{})
	return nil
}

// matchesListOptions checks if a group matches the list criteria
func matchesListOptions(g *group.Group, opts group.ListOptions) bool {
	// Parent ID filter
	if opts.ParentID != "" && g.ParentID != opts.ParentID {
		return false
	}

	// Group type filter
	if opts.Type != "" && g.Type != opts.Type {
		return false
	}

	// Tags filter
	if len(opts.Tags) > 0 {
		for k, v := range opts.Tags {
			if tv, ok := g.Properties.Metadata[k]; !ok || tv != v {
				return false
			}
		}
	}

	return true
}
