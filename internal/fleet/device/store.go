package device

import (
	"context"
)

// Store defines the interface for device persistence
type Store interface {
	// Create stores a new device
	Create(ctx context.Context, device *Device) error

	// Get retrieves a device by ID
	Get(ctx context.Context, tenantID, deviceID string) (*Device, error)

	// Update modifies an existing device
	Update(ctx context.Context, device *Device) error

	// Delete removes a device
	Delete(ctx context.Context, tenantID, deviceID string) error

	// List retrieves devices matching the given options
	List(ctx context.Context, opts ListOptions) ([]*Device, error)
}

// ListOptions defines parameters for listing devices
type ListOptions struct {
	TenantID string
	Tags     map[string]string
	Status   Status
	Offset   int
	Limit    int
}
