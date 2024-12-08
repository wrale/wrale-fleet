package group

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
)

// memoryStore provides an in-memory implementation of Store interface for testing
type memoryStore struct {
	mu          sync.RWMutex
	groups      map[string]*Group              // key: tenantID:groupID
	memberships map[string]map[string]struct{} // key: tenantID:groupID -> map[deviceID]struct{}
	deviceStore device.Store
}

func newMemoryStore(deviceStore device.Store) Store {
	return &memoryStore{
		groups:      make(map[string]*Group),
		memberships: make(map[string]map[string]struct{}),
		deviceStore: deviceStore,
	}
}

// key generates the map key for a group
func (s *memoryStore) key(tenantID, groupID string) string {
	return fmt.Sprintf("%s:%s", tenantID, groupID)
}

func (s *memoryStore) Create(ctx context.Context, g *Group) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("validate group: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(g.TenantID, g.ID)
	if _, exists := s.groups[key]; exists {
		return ErrGroupExists
	}

	// Use DeepCopy to ensure complete isolation
	s.groups[key] = g.DeepCopy()
	s.memberships[key] = make(map[string]struct{})

	return nil
}

func (s *memoryStore) Get(ctx context.Context, tenantID, groupID string) (*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, ErrGroupNotFound
	}

	// Return a deep copy to prevent modifications through the returned reference
	return g.DeepCopy(), nil
}

func (s *memoryStore) Update(ctx context.Context, g *Group) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("validate group: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(g.TenantID, g.ID)
	if _, exists := s.groups[key]; !exists {
		return ErrGroupNotFound
	}

	// Store a deep copy to ensure complete isolation
	s.groups[key] = g.DeepCopy()
	return nil
}

func (s *memoryStore) Delete(ctx context.Context, tenantID, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return ErrGroupNotFound
	}

	if len(g.Ancestry.Children) > 0 {
		return E("Store.Delete", ErrCodeInvalidOperation,
			"cannot delete group with existing children", nil)
	}

	if g.ParentID != "" {
		parentKey := s.key(tenantID, g.ParentID)
		if parent, exists := s.groups[parentKey]; exists {
			// Create a copy, modify it, and store it back
			parentCopy := parent.DeepCopy()
			parentCopy.RemoveChild(groupID)
			s.groups[parentKey] = parentCopy
		}
	}

	delete(s.groups, key)
	delete(s.memberships, key)
	return nil
}

func (s *memoryStore) List(ctx context.Context, opts ListOptions) ([]*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Group

	for _, g := range s.groups {
		if !s.matchesFilter(g, opts) {
			continue
		}

		// Append deep copy to results
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

func (s *memoryStore) AddDevice(ctx context.Context, tenantID, groupID string, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return ErrGroupNotFound
	}

	if g.Type != TypeStatic {
		return E("Store.AddDevice", ErrCodeInvalidOperation,
			"cannot manually add device to dynamic group", nil)
	}

	if g.TenantID != d.TenantID {
		return E("Store.AddDevice", ErrCodeInvalidOperation,
			"device tenant does not match group tenant", nil)
	}

	members := s.memberships[key]
	members[d.ID] = struct{}{}

	// Update group with new device count
	groupCopy := g.DeepCopy()
	groupCopy.DeviceCount = len(members)
	s.groups[key] = groupCopy

	return nil
}

func (s *memoryStore) RemoveDevice(ctx context.Context, tenantID, groupID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return ErrGroupNotFound
	}

	if g.Type != TypeStatic {
		return E("Store.RemoveDevice", ErrCodeInvalidOperation,
			"cannot manually remove device from dynamic group", nil)
	}

	members := s.memberships[key]
	delete(members, deviceID)

	// Update group with new device count
	groupCopy := g.DeepCopy()
	groupCopy.DeviceCount = len(members)
	s.groups[key] = groupCopy

	return nil
}

func (s *memoryStore) ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error) {
	s.mu.RLock()
	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		s.mu.RUnlock()
		return nil, ErrGroupNotFound
	}

	var deviceIDs []string
	if g.Type == TypeStatic {
		members := s.memberships[key]
		for deviceID := range members {
			deviceIDs = append(deviceIDs, deviceID)
		}
	} else {
		s.mu.RUnlock()
		return s.evaluateDynamicGroupMembers(ctx, g.DeepCopy())
	}
	s.mu.RUnlock()

	var devices []*device.Device
	for _, deviceID := range deviceIDs {
		device, err := s.deviceStore.Get(ctx, tenantID, deviceID)
		if err != nil {
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (s *memoryStore) evaluateDynamicGroupMembers(ctx context.Context, g *Group) ([]*device.Device, error) {
	if g.Query == nil {
		return nil, nil
	}

	opts := device.ListOptions{
		TenantID: g.TenantID,
		Status:   g.Query.Status,
		Tags:     g.Query.Tags,
	}

	devices, err := s.deviceStore.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	s.mu.Lock()
	if grp, exists := s.groups[s.key(g.TenantID, g.ID)]; exists {
		groupCopy := grp.DeepCopy()
		groupCopy.DeviceCount = len(devices)
		s.groups[s.key(g.TenantID, g.ID)] = groupCopy
	}
	s.mu.Unlock()

	return devices, nil
}

func (s *memoryStore) GetAncestors(ctx context.Context, tenantID, groupID string) ([]*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, ErrGroupNotFound
	}

	ancestors := make([]*Group, 0, g.Ancestry.Depth)
	for _, ancestorID := range g.Ancestry.PathParts[:len(g.Ancestry.PathParts)-1] {
		ancestorKey := s.key(tenantID, ancestorID)
		ancestor, exists := s.groups[ancestorKey]
		if !exists {
			return nil, E("Store.GetAncestors", ErrCodeStoreOperation,
				fmt.Sprintf("ancestor %s not found", ancestorID), nil)
		}
		ancestors = append(ancestors, ancestor.DeepCopy())
	}

	return ancestors, nil
}

func (s *memoryStore) GetDescendants(ctx context.Context, tenantID, groupID string) ([]*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, ErrGroupNotFound
	}

	descendants := make([]*Group, 0)
	for _, childID := range g.Ancestry.Children {
		childKey := s.key(tenantID, childID)
		child, exists := s.groups[childKey]
		if !exists {
			return nil, E("Store.GetDescendants", ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}

		descendants = append(descendants, child.DeepCopy())

		childDescendants, err := s.GetDescendants(ctx, tenantID, childID)
		if err != nil {
			return nil, fmt.Errorf("get child descendants: %w", err)
		}
		descendants = append(descendants, childDescendants...)
	}

	return descendants, nil
}

func (s *memoryStore) GetChildren(ctx context.Context, tenantID, groupID string) ([]*Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, ErrGroupNotFound
	}

	children := make([]*Group, 0, len(g.Ancestry.Children))
	for _, childID := range g.Ancestry.Children {
		childKey := s.key(tenantID, childID)
		child, exists := s.groups[childKey]
		if !exists {
			return nil, E("Store.GetChildren", ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}
		children = append(children, child.DeepCopy())
	}

	return children, nil
}

func (s *memoryStore) ValidateHierarchy(ctx context.Context, tenantID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tenantGroups []*Group
	for _, g := range s.groups {
		if g.TenantID == tenantID {
			tenantGroups = append(tenantGroups, g.DeepCopy())
		}
	}

	groupMap := make(map[string]*Group)
	for _, g := range tenantGroups {
		groupMap[g.ID] = g
	}

	for _, g := range tenantGroups {
		if g.ParentID != "" {
			parent, exists := groupMap[g.ParentID]
			if !exists {
				return E("Store.ValidateHierarchy", ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == g.ID {
					found = true
					break
				}
			}
			if !found {
				return E("Store.ValidateHierarchy", ErrCodeInvalidGroup,
					fmt.Sprintf("group %s not found in parent %s children list", g.ID, parent.ID), nil)
			}
		}

		for _, childID := range g.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E("Store.ValidateHierarchy", ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}
			if child.ParentID != g.ID {
				return E("Store.ValidateHierarchy", ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference parent %s", child.ID, g.ID), nil)
			}
		}
	}

	return nil
}

func (s *memoryStore) matchesFilter(g *Group, opts ListOptions) bool {
	if opts.TenantID != "" && g.TenantID != opts.TenantID {
		return false
	}

	if opts.ParentID != "" && g.ParentID != opts.ParentID {
		return false
	}

	if opts.Type != "" && g.Type != opts.Type {
		return false
	}

	if opts.Depth >= 0 && g.Ancestry.Depth != opts.Depth {
		return false
	}

	if len(opts.Tags) > 0 {
		for k, v := range opts.Tags {
			if gv, ok := g.Properties.Metadata[k]; !ok || gv != v {
				return false
			}
		}
	}

	return true
}
