package group

import (
	"context"
	"strings"

	"github.com/wrale/fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// Service provides group management operations
type Service struct {
	store       Store
	deviceStore device.Store
	logger      *zap.Logger
}

// NewService creates a new group management service
func NewService(store Store, deviceStore device.Store, logger *zap.Logger) *Service {
	return &Service{
		store:       store,
		deviceStore: deviceStore,
		logger:      logger,
	}
}

// Create creates a new device group
func (s *Service) Create(ctx context.Context, tenantID, name string, groupType Type) (*Group, error) {
	const op = "group.Service.Create"

	group := New(tenantID, name, groupType)
	if err := group.Validate(); err != nil {
		return nil, E(op, ErrCodeInvalidInput, "invalid group data", err)
	}

	if err := s.store.Create(ctx, group); err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to create group", err)
	}

	s.logger.Info("created new group",
		zap.String("group_id", group.ID),
		zap.String("tenant_id", group.TenantID),
		zap.String("name", group.Name),
		zap.String("type", string(group.Type)),
	)

	return group, nil
}

// Get retrieves a group by ID
func (s *Service) Get(ctx context.Context, tenantID, groupID string) (*Group, error) {
	const op = "group.Service.Get"

	group, err := s.store.Get(ctx, tenantID, groupID)
	if err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to get group", err)
	}
	return group, nil
}

// Update updates an existing group
func (s *Service) Update(ctx context.Context, group *Group) error {
	const op = "group.Service.Update"

	if err := group.Validate(); err != nil {
		return E(op, ErrCodeInvalidInput, "invalid group data", err)
	}

	// If parent ID is changing, verify no cycles would be created
	if group.ParentID != "" {
		if err := s.validateHierarchy(ctx, group.TenantID, group.ID, group.ParentID); err != nil {
			return E(op, ErrCodeInvalidOperation, "invalid group hierarchy", err)
		}
	}

	if err := s.store.Update(ctx, group); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	s.logger.Info("updated group",
		zap.String("group_id", group.ID),
		zap.String("tenant_id", group.TenantID),
		zap.Time("updated_at", group.UpdatedAt),
	)

	return nil
}

// Delete removes a group and its child groups
func (s *Service) Delete(ctx context.Context, tenantID, groupID string) error {
	const op = "group.Service.Delete"

	// Get all child groups
	children, err := s.List(ctx, ListOptions{
		TenantID: tenantID,
		ParentID: groupID,
	})
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to list child groups", err)
	}

	// Delete children first
	for _, child := range children {
		if err := s.Delete(ctx, tenantID, child.ID); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to delete child group", err)
		}
	}

	// Delete the group itself
	if err := s.store.Delete(ctx, tenantID, groupID); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to delete group", err)
	}

	s.logger.Info("deleted group",
		zap.String("group_id", groupID),
		zap.String("tenant_id", tenantID),
		zap.Int("child_groups_deleted", len(children)),
	)

	return nil
}

// List retrieves groups matching the given criteria
func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Group, error) {
	const op = "group.Service.List"
	groups, err := s.store.List(ctx, opts)
	if err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}
	return groups, nil
}

// AddDevice adds a device to a static group
func (s *Service) AddDevice(ctx context.Context, tenantID, groupID string, device *device.Device) error {
	const op = "group.Service.AddDevice"

	group, err := s.store.Get(ctx, tenantID, groupID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get group", err)
	}

	if group.Type != TypeStatic {
		return E(op, ErrCodeInvalidOperation, "cannot manually add device to dynamic group", nil)
	}

	if err := s.store.AddDevice(ctx, tenantID, groupID, device); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to add device to group", err)
	}

	s.logger.Info("added device to group",
		zap.String("group_id", groupID),
		zap.String("device_id", device.ID),
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// RemoveDevice removes a device from a static group
func (s *Service) RemoveDevice(ctx context.Context, tenantID, groupID, deviceID string) error {
	const op = "group.Service.RemoveDevice"

	group, err := s.store.Get(ctx, tenantID, groupID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get group", err)
	}

	if group.Type != TypeStatic {
		return E(op, ErrCodeInvalidOperation, "cannot manually remove device from dynamic group", nil)
	}

	if err := s.store.RemoveDevice(ctx, tenantID, groupID, deviceID); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to remove device from group", err)
	}

	s.logger.Info("removed device from group",
		zap.String("group_id", groupID),
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// ListDevices lists all devices in a group
func (s *Service) ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error) {
	const op = "group.Service.ListDevices"
	
	devices, err := s.store.ListDevices(ctx, tenantID, groupID)
	if err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to list devices in group", err)
	}
	return devices, nil
}

// validateHierarchy ensures no cycles would be created in the group hierarchy
func (s *Service) validateHierarchy(ctx context.Context, tenantID, groupID, newParentID string) error {
	const op = "group.Service.validateHierarchy"

	// Check that the new parent exists
	parent, err := s.store.Get(ctx, tenantID, newParentID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get parent group", err)
	}

	// Prevent self-referential cycles
	if groupID == newParentID {
		return E(op, ErrCodeInvalidOperation, "group cannot be its own parent", ErrCyclicDependency)
	}

	// Check that the new parent isn't a descendant of the group
	parentPath := parent.Path
	if strings.Contains(parentPath, "/"+groupID+"/") {
		return E(op, ErrCodeInvalidOperation, "cyclic dependency detected", ErrCyclicDependency)
	}

	return nil
}
