package memory

import (
	"context"
	"fmt"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
)

// AddDevice implements group.Store
func (s *Store) AddDevice(ctx context.Context, tenantID, groupID string, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the group
	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return group.E("Store.AddDevice", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return group.E("Store.AddDevice", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	if g.Type != group.TypeStatic {
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation,
			"cannot manually add device to dynamic group", nil)
	}

	if g.TenantID != d.TenantID {
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation,
			"device tenant does not match group tenant", nil)
	}

	key := s.groupKey(tenantID, groupID)
	if s.memberships[key] == nil {
		s.memberships[key] = make(map[string]struct{})
	}
	s.memberships[key][d.ID] = struct{}{}

	// Update group with new device count
	groupCopy := g.DeepCopy()
	groupCopy.DeviceCount = len(s.memberships[key])
	s.groups[tenantID][groupID] = groupCopy

	return nil
}

// RemoveDevice implements group.Store
func (s *Store) RemoveDevice(ctx context.Context, tenantID, groupID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the group
	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return group.E("Store.RemoveDevice", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return group.E("Store.RemoveDevice", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	if g.Type != group.TypeStatic {
		return group.E("Store.RemoveDevice", group.ErrCodeInvalidOperation,
			"cannot manually remove device from dynamic group", nil)
	}

	key := s.groupKey(tenantID, groupID)
	if s.memberships[key] != nil {
		delete(s.memberships[key], deviceID)

		// Update group with new device count
		groupCopy := g.DeepCopy()
		groupCopy.DeviceCount = len(s.memberships[key])
		s.groups[tenantID][groupID] = groupCopy
	}

	return nil
}

// ListDevices implements group.Store
func (s *Store) ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the group
	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil, group.E("Store.ListDevices", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return nil, group.E("Store.ListDevices", group.ErrCodeGroupNotFound,
			"group not found", nil)
	}

	var devices []*device.Device

	if g.Type == group.TypeStatic {
		key := s.groupKey(tenantID, groupID)
		members := s.memberships[key]
		if members == nil {
			return []*device.Device{}, nil
		}

		for deviceID := range members {
			device, err := s.deviceStore.Get(ctx, tenantID, deviceID)
			if err != nil {
				// Log error but continue - the device might have been deleted
				continue
			}
			devices = append(devices, device)
		}
	} else {
		// For dynamic groups, evaluate the query
		var err error
		devices, err = s.evaluateDynamicGroupMembers(ctx, g)
		if err != nil {
			return nil, fmt.Errorf("evaluate dynamic members: %w", err)
		}
	}

	return devices, nil
}

// evaluateDynamicGroupMembers handles device listing for dynamic groups
func (s *Store) evaluateDynamicGroupMembers(ctx context.Context, g *group.Group) ([]*device.Device, error) {
	if g.Query == nil {
		return []*device.Device{}, nil
	}

	// Prepare device list options from group query
	opts := device.ListOptions{
		TenantID: g.TenantID,
		Tags:     g.Query.Tags,
	}

	devices, err := s.deviceStore.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	// Update the group's device count
	s.mu.Lock()
	if _, exists := s.groups[g.TenantID][g.ID]; exists {
		groupCopy := g.DeepCopy()
		groupCopy.DeviceCount = len(devices)
		s.groups[g.TenantID][g.ID] = groupCopy
	}
	s.mu.Unlock()

	return devices, nil
}
