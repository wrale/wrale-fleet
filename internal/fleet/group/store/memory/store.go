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
	mu            sync.RWMutex
	groups        map[string]*group.Group        // key: tenantID:groupID
	memberships   map[string]map[string]struct{} // key: tenantID:groupID -> map[deviceID]struct{}
	deviceStore   device.Store
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

// AddDevice adds a device to a static group
func (s *Store) AddDevice(ctx context.Context, tenantID, groupID string, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return group.ErrGroupNotFound
	}

	if g.Type != group.TypeStatic {
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation, "cannot manually add device to dynamic group", nil)
	}

	if g.TenantID != d.TenantID {
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation, "device tenant does not match group tenant", nil)
	}

	members := s.memberships[key]
	members[d.ID] = struct{}{}
	g.DeviceCount = len(members)
	return nil
}

// RemoveDevice removes a device from a static group
func (s *Store) RemoveDevice(ctx context.Context, tenantID, groupID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return group.ErrGroupNotFound
	}

	if g.Type != group.TypeStatic {
		return group.E("Store.RemoveDevice", group.ErrCodeInvalidOperation, "cannot manually remove device from dynamic group", nil)
	}

	members := s.memberships[key]
	delete(members, deviceID)
	g.DeviceCount = len(members)
	return nil
}

// ListDevices lists all devices in a group
func (s *Store) ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error) {
	s.mu.RLock()
	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		s.mu.RUnlock()
		return nil, group.ErrGroupNotFound
	}

	var deviceIDs []string
	if g.Type == group.TypeStatic {
		// For static groups, use membership map
		members := s.memberships[key]
		for deviceID := range members {
			deviceIDs = append(deviceIDs, deviceID)
		}
	} else {
		// For dynamic groups, evaluate query
		// Release lock before querying devices
		s.mu.RUnlock()
		return s.evaluateDynamicGroupMembers(ctx, g)
	}
	s.mu.RUnlock()

	// Fetch actual device objects
	var devices []*device.Device
	for _, deviceID := range deviceIDs {
		device, err := s.deviceStore.Get(ctx, tenantID, deviceID)
		if err != nil {
			// Skip devices that may have been deleted
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// matchesFilter checks if a group matches the filter criteria
func (s *Store) matchesFilter(g *group.Group, opts group.ListOptions) bool {
	if opts.TenantID != "" && g.TenantID != opts.TenantID {
		return false
	}

	if opts.ParentID != "" && g.ParentID != opts.ParentID {
		return false
	}

	if opts.Type != "" && g.Type != opts.Type {
		return false
	}

	// Check if all required tags are present with matching values
	for key, value := range opts.Tags {
		if g.Properties.Metadata[key] != value {
			return false
		}
	}

	return true
}

// evaluateDynamicGroupMembers evaluates the group's query to find matching devices
func (s *Store) evaluateDynamicGroupMembers(ctx context.Context, g *group.Group) ([]*device.Device, error) {
	if g.Query == nil {
		return nil, nil
	}

	// Convert group query to device list options
	opts := device.ListOptions{
		TenantID: g.TenantID,
		Status:   g.Query.Status,
		Tags:     g.Query.Tags,
	}

	devices, err := s.deviceStore.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	// Update device count
	s.mu.Lock()
	g.DeviceCount = len(devices)
	s.mu.Unlock()

	return devices, nil
}
