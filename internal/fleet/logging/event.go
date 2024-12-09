// Package logging provides core logging functionality for the fleet management system.
package logging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Level represents the severity level of a log event
type Level string

const (
	// LevelDebug is for detailed troubleshooting
	LevelDebug Level = "debug"
	// LevelInfo is for general operational events
	LevelInfo Level = "info"
	// LevelWarn is for important but non-critical events
	LevelWarn Level = "warn"
	// LevelError is for error conditions
	LevelError Level = "error"
)

// EventType categorizes the type of log event
type EventType string

const (
	// EventSystem represents system-level events
	EventSystem EventType = "system"
	// EventSecurity represents security-related events
	EventSecurity EventType = "security"
	// EventAudit represents audit trail events
	EventAudit EventType = "audit"
	// EventCompliance represents compliance-related events
	EventCompliance EventType = "compliance"
	// EventOperational represents operational events
	EventOperational EventType = "operational"
)

// EventContext provides additional context for log events
type EventContext struct {
	// ComponentID identifies the system component that generated the event
	ComponentID string `json:"component_id,omitempty"`
	// DeviceID identifies a related device, if applicable
	DeviceID string `json:"device_id,omitempty"`
	// UserID identifies the user who triggered the event, if applicable
	UserID string `json:"user_id,omitempty"`
	// RequestID identifies the related request for tracing
	RequestID string `json:"request_id,omitempty"`
	// Stage indicates the capability stage requirement (1-6)
	Stage int `json:"stage,omitempty"`
}

// Event represents a log event in the system
type Event struct {
	// ID uniquely identifies this event
	ID string `json:"id"`

	// TenantID ensures proper multi-tenant isolation
	TenantID string `json:"tenant_id"`

	// Type categorizes the event
	Type EventType `json:"type"`

	// Level indicates event severity
	Level Level `json:"level"`

	// Message is the human-readable event description
	Message string `json:"message"`

	// Timestamp records when the event occurred
	Timestamp time.Time `json:"timestamp"`

	// Context provides additional event details
	Context EventContext `json:"context,omitempty"`

	// Metadata stores arbitrary structured data about the event
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// Source indicates where the event originated (e.g., hostname, IP)
	Source string `json:"source,omitempty"`

	// Tags allow for flexible event categorization
	Tags map[string]string `json:"tags,omitempty"`

	// Hash is a unique content hash for deduplication
	Hash string `json:"hash,omitempty"`

	// RetentionPolicy specifies how long to retain this event
	RetentionPolicy string `json:"retention_policy,omitempty"`
}

// New creates a new Event with a unique ID and current timestamp
func New(tenantID string, eventType EventType, level Level, message string) *Event {
	return &Event{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Type:      eventType,
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Tags:      make(map[string]string),
	}
}

// WithContext adds context information to the event
func (e *Event) WithContext(ctx EventContext) *Event {
	e.Context = ctx
	return e
}

// WithMetadata adds structured metadata to the event
func (e *Event) WithMetadata(metadata interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	e.Metadata = data
	return nil
}

// WithTag adds a tag to the event
func (e *Event) WithTag(key, value string) *Event {
	if e.Tags == nil {
		e.Tags = make(map[string]string)
	}
	e.Tags[key] = value
	return e
}

// WithSource sets the event source
func (e *Event) WithSource(source string) *Event {
	e.Source = source
	return e
}

// WithRetention sets the event retention policy
func (e *Event) WithRetention(policy string) *Event {
	e.RetentionPolicy = policy
	return e
}

// Validate checks if the event data is valid
func (e *Event) Validate() error {
	if e.TenantID == "" {
		return ErrMissingTenant
	}
	if e.Message == "" {
		return ErrMissingMessage
	}
	return nil
}
