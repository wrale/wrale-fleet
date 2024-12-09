package health

import (
	"context"
)

// Store defines the interface for health status persistence
type Store interface {
	// UpdateComponentStatus updates the health status for a specific component
	UpdateComponentStatus(ctx context.Context, component string, status *HealthStatus) error

	// GetComponentStatus retrieves the health status for a specific component
	GetComponentStatus(ctx context.Context, component string) (*HealthStatus, error)

	// ListComponentStatuses retrieves health status for all components
	ListComponentStatuses(ctx context.Context) (map[string]*HealthStatus, error)

	// GetReadyStatus retrieves the current ready status
	GetReadyStatus(ctx context.Context) (bool, error)

	// SetReadyStatus updates the ready status
	SetReadyStatus(ctx context.Context, ready bool) error

	// RegisterComponent registers a new component for health monitoring
	RegisterComponent(ctx context.Context, component string, info ComponentInfo) error

	// UnregisterComponent removes a component from health monitoring
	UnregisterComponent(ctx context.Context, component string) error
}

// StoreOption defines functional options for configuring stores
type StoreOption func(*StoreOptions) error

// StoreOptions contains configuration options for health stores
type StoreOptions struct {
	RetentionPeriod string
}

// WithRetentionPeriod sets how long to retain historical health data
func WithRetentionPeriod(period string) StoreOption {
	return func(o *StoreOptions) error {
		o.RetentionPeriod = period
		return nil
	}
}
