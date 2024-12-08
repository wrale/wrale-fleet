package group

import "context"

// Store defines the interface for group storage operations
type Store interface {
	// Create creates a new group
	Create(ctx context.Context, group *Group) error

	// Get retrieves a group by ID and tenant
	Get(ctx context.Context, tenantID, id string) (*Group, error)

	// Update updates an existing group
	Update(ctx context.Context, group *Group) error

	// Delete removes a group
	Delete(ctx context.Context, tenantID, id string) error

	// List returns groups matching the query criteria
	List(ctx context.Context, tenantID string, opts ListOptions) ([]*Group, error)

	// Clear removes all groups (used for testing)
	Clear(ctx context.Context) error
}

// ListOptions defines criteria for listing groups
type ListOptions struct {
	ParentID     string            // Filter by parent ID
	Type         Type              // Filter by group type
	Tags         map[string]string // Filter by tags
	IncludeEmpty bool              // Include groups with no devices
	Offset       int               // Pagination offset
	Limit        int               // Pagination limit
}
