package group

import (
	"context"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/group/store/memory"
)

// Store defines the interface for group persistence
type Store interface {
	// Core CRUD Operations
	Create(ctx context.Context, group *Group) error
	Get(ctx context.Context, tenantID, groupID string) (*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, tenantID, groupID string) error
	List(ctx context.Context, opts ListOptions) ([]*Group, error)

	// Device Management Operations
	AddDevice(ctx context.Context, tenantID, groupID string, device *device.Device) error
	RemoveDevice(ctx context.Context, tenantID, groupID string, deviceID string) error
	ListDevices(ctx context.Context, tenantID, groupID string) ([]*device.Device, error)

	// Hierarchy Operations
	GetAncestors(ctx context.Context, tenantID, groupID string) ([]*Group, error)
	GetDescendants(ctx context.Context, tenantID, groupID string) ([]*Group, error)
	GetChildren(ctx context.Context, tenantID, groupID string) ([]*Group, error)
	ValidateHierarchy(ctx context.Context, tenantID string) error
}

// NewMemoryStore creates a new in-memory implementation of the Store interface.
// This is primarily used for testing and demonstration purposes.
func NewMemoryStore(deviceStore device.Store) Store {
	return memory.New(deviceStore)
}

// ListOptions defines parameters for listing groups
type ListOptions struct {
	TenantID  string
	ParentID  string            // Filter by parent group
	Type      Type              // Filter by group type
	Tags      map[string]string // Filter by metadata tags
	Depth     int               // Filter by hierarchy depth (-1 for all)
	Offset    int
	Limit     int
	SortBy    string // Field to sort by
	SortOrder string // "asc" or "desc"
}

// QueryOptions defines advanced query parameters for group operations
type QueryOptions struct {
	IncludeChildren    bool              // Include child groups in operations
	IncludeDescendants bool              // Include all descendant groups in operations
	IncludeDevices     bool              // Include device information
	FilterTags         map[string]string // Additional tag-based filtering
	MaxDepth           int               // Maximum depth for hierarchy operations (-1 for unlimited)
}

// BatchOperation represents a batch update operation for groups
type BatchOperation struct {
	GroupIDs []string          // Groups to update
	Updates  map[string]string // Key-value pairs of updates to apply
	Options  QueryOptions      // Operation options
}