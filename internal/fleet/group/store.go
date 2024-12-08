package group

import (
	"context"

	"github.com/wrale/fleet/internal/fleet/device"
)

// Store defines the interface for group persistence
type Store interface {
	// Create stores a new group
	Create(ctx context.Context, group *Group) error

	// Get retrieves a group by ID
	Get(ctx context.Context, tenantID, groupID string) (*Group, error)

	// Update modifies an existing group
	Update(ctx context.Context, group *Group) error

	// Delete removes a group
	Delete(ctx context.Context, tenantID, groupID string) error

	// List retrieves groups matching the given options
	List(ctx context.Context, opts ListOptions) ([]*Group, error)

	// AddDevice adds a device to a static group
	AddDevice(ctx context.Context, tenantID, groupID string, device *device.Device) error

	// RemoveDevice removes a device from a static group
	RemoveDevice(ctx context.Context, tenantID, groupID string, deviceID string) error

	// ListDevices lists all devices in a group (both static and dynamic)
	ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error)
}

// ListOptions defines parameters for listing groups
type ListOptions struct {
	TenantID string
	ParentID string            // Filter by parent group
	Type     Type              // Filter by group type
	Tags     map[string]string // Filter by metadata tags
	Offset   int
	Limit    int
}