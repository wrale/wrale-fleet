package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
)

// Store provides an in-memory implementation of group.Store interface
type Store struct {
	mu          sync.RWMutex
	groups      map[string]*group.Group        // key: tenantID:groupID
	memberships map[string]map[string]struct{} // key: tenantID:groupID -> map[deviceID]struct{}
	deviceStore device.Store
}

// New creates a new in-memory group store
func New(deviceStore device.Store) group.Store {
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

// Create implements group.Store
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

	// Use DeepCopy to ensure complete isolation
	s.groups[key] = g.DeepCopy()
	s.memberships[key] = make(map[string]struct{})

	return nil
}

// Get implements group.Store
func (s *Store) Get(ctx context.Context, tenantID, groupID string) (*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	return g.DeepCopy(), nil
}

// Update implements group.Store
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

	s.groups[key] = g.DeepCopy()
	return nil
}

// Delete implements group.Store
func (s *Store) Delete(ctx context.Context, tenantID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return group.ErrGroupNotFound
	}

	if len(g.Ancestry.Children) > 0 {
		return group.E("Store.Delete", group.ErrCodeInvalidOperation,
			"cannot delete group with existing children", nil)
	}

	if g.ParentID != "" {
		parentKey := s.key(tenantID, g.ParentID)
		if parent, exists := s.groups[parentKey]; exists {
			parentCopy := parent.DeepCopy()
			parentCopy.RemoveChild(groupID)
			s.groups[parentKey] = parentCopy
		}
	}

	delete(s.groups, key)
	delete(s.memberships, key)
	return nil
}

// List implements group.Store
func (s *Store) List(ctx context.Context, opts group.ListOptions) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*group.Group

	for _, g := range s.groups {
		if !s.matchesFilter(g, opts) {
			continue
		}
		result = append(result, g.DeepCopy())
	}

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
