package logging

import (
	"context"
	"time"
)

// Store defines the interface for event storage implementations
type Store interface {
	// Store stores a new event
	Store(ctx context.Context, event *Event) error

	// Get retrieves an event by ID
	Get(ctx context.Context, tenantID, eventID string) (*Event, error)

	// List retrieves events matching the given options
	List(ctx context.Context, opts ListOptions) ([]*Event, error)

	// Delete removes an event
	Delete(ctx context.Context, tenantID, eventID string) error

	// DeleteBefore removes events older than the given time
	DeleteBefore(ctx context.Context, tenantID string, before time.Time) error

	// Query performs a structured query on events
	Query(ctx context.Context, query QueryOptions) ([]*Event, error)

	// BatchStore stores multiple events in a single operation
	BatchStore(ctx context.Context, events []*Event) error

	// Sync ensures all events are persisted
	Sync(ctx context.Context) error
}
