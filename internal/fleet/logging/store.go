package logging

import (
	"context"
	"time"
)

// Store defines the interface for event persistence
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

	// Sync ensures all events are persisted (important for airgapped operations)
	Sync(ctx context.Context) error
}

// ListOptions defines parameters for listing events
type ListOptions struct {
	// TenantID filters events by tenant
	TenantID string

	// Type filters events by type
	Type EventType

	// Level filters events by minimum severity
	Level Level

	// StartTime filters events after this time
	StartTime *time.Time

	// EndTime filters events before this time
	EndTime *time.Time

	// Source filters events by source
	Source string

	// Tags filters events by tags (all must match)
	Tags map[string]string

	// ComponentID filters events by component
	ComponentID string

	// DeviceID filters events by device
	DeviceID string

	// Offset for pagination
	Offset int

	// Limit maximum number of results
	Limit int
}

// QueryOptions provides advanced event querying capabilities
type QueryOptions struct {
	// TenantID is required for multi-tenant isolation
	TenantID string

	// Types filters by multiple event types
	Types []EventType

	// Levels filters by multiple severity levels
	Levels []Level

	// TimeRange specifies the time window for events
	TimeRange *TimeRange

	// Sources filters by multiple sources
	Sources []string

	// TagQuery allows complex tag matching
	TagQuery *TagQuery

	// ContextQuery allows filtering on event context fields
	ContextQuery *ContextQuery

	// RetentionPolicies filters by retention policies
	RetentionPolicies []string

	// OrderBy specifies sort order
	OrderBy string

	// OrderDirection specifies sort direction (asc/desc)
	OrderDirection string

	// Offset for pagination
	Offset int

	// Limit maximum number of results
	Limit int
}

// TimeRange specifies a time window for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// TagQuery enables complex tag matching
type TagQuery struct {
	// Must contains tags that must all match
	Must map[string]string

	// Should contains tags where at least one must match
	Should map[string]string

	// MustNot contains tags that must not match
	MustNot map[string]string
}

// ContextQuery enables filtering on event context
type ContextQuery struct {
	ComponentIDs []string
	DeviceIDs    []string
	UserIDs      []string
	RequestIDs   []string
	MinStage     int
	MaxStage     int
}
