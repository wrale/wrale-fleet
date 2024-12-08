package memory

import (
	"context"
	"fmt"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group"
)

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
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation,
			"cannot manually add device to dynamic group", nil)
	}

	if g.TenantID != d.TenantID {
		return group.E("Store.AddDevice", group.ErrCodeInvalidOperation,
			"device tenant does not match group tenant", nil)
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
		return group.E("Store.RemoveDevice", group.ErrCodeInvalidOperation,
			"cannot manually remove device from dynamic group", nil)
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

	// Update device count while holding lock
	s.mu.Lock()
	if grp, exists := s.groups[s.key(g.TenantID, g.ID)]; exists {
		grp.DeviceCount = len(devices)
	}
	s.mu.Unlock()

	return devices, nil
}
