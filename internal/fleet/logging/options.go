package logging

import "time"

// ListOptions provides filtering and pagination options for listing events
type ListOptions struct {
	// TenantID ensures proper multi-tenant isolation
	TenantID string

	// Type filters events by type
	Type EventType

	// Level filters events by severity level
	Level Level

	// Source filters events by source
	Source string

	// StartTime filters events after this time
	StartTime *time.Time

	// EndTime filters events before this time
	EndTime *time.Time

	// ComponentID filters events by component
	ComponentID string

	// DeviceID filters events by device
	DeviceID string

	// Tags filters events by tag key-value pairs
	Tags map[string]string

	// Offset is the number of items to skip
	Offset int

	// Limit is the maximum number of items to return
	Limit int
}

// QueryOptions provides advanced query capabilities for event search
type QueryOptions struct {
	// TenantID ensures proper multi-tenant isolation
	TenantID string

	// Types filters events by multiple types
	Types []EventType

	// Levels filters events by multiple severity levels
	Levels []Level

	// Sources filters events by multiple sources
	Sources []string

	// TimeRange specifies a time window for events
	TimeRange *TimeRange

	// TagQuery provides complex tag matching
	TagQuery *TagQuery

	// ContextQuery provides context-based filtering
	ContextQuery *ContextQuery

	// Offset is the number of items to skip
	Offset int

	// Limit is the maximum number of items to return
	Limit int

	// OrderBy specifies the field to sort by
	OrderBy string

	// OrderDirection specifies sort order ("asc" or "desc")
	OrderDirection string
}

// TimeRange represents a time window for event queries
type TimeRange struct {
	// Start is the beginning of the time range
	Start time.Time

	// End is the end of the time range
	End time.Time
}

// TagQuery provides complex tag matching capabilities
type TagQuery struct {
	// Must contains tags that must all match
	Must map[string]string

	// Should contains tags where at least one must match
	Should map[string]string

	// MustNot contains tags that must not match
	MustNot map[string]string
}

// ContextQuery provides context-based filtering options
type ContextQuery struct {
	// ComponentIDs filters by multiple component IDs
	ComponentIDs []string

	// DeviceIDs filters by multiple device IDs
	DeviceIDs []string

	// MinStage filters events by minimum stage requirement
	MinStage int

	// MaxStage filters events by maximum stage requirement
	MaxStage int
}
