package logging

import (
	"context"
	"encoding/json"
	"time"
)

// AuditAction represents the type of audit event
type AuditAction string

const (
	// AuditActionCreate represents resource creation
	AuditActionCreate AuditAction = "create"
	// AuditActionUpdate represents resource modification
	AuditActionUpdate AuditAction = "update"
	// AuditActionDelete represents resource deletion
	AuditActionDelete AuditAction = "delete"
	// AuditActionAccess represents resource access
	AuditActionAccess AuditAction = "access"
	// AuditActionAuth represents authentication events
	AuditActionAuth AuditAction = "auth"
	// AuditActionConfig represents configuration changes
	AuditActionConfig AuditAction = "config"
)

// AuditMetadata provides structured data for audit events
type AuditMetadata struct {
	// Action describes what was done
	Action AuditAction `json:"action"`

	// ResourceType identifies the type of resource affected
	ResourceType string `json:"resource_type"`

	// ResourceID identifies the specific resource
	ResourceID string `json:"resource_id"`

	// PreviousState captures the state before changes (if applicable)
	PreviousState json.RawMessage `json:"previous_state,omitempty"`

	// NewState captures the state after changes (if applicable)
	NewState json.RawMessage `json:"new_state,omitempty"`

	// Changes describes specific modifications made
	Changes map[string]interface{} `json:"changes,omitempty"`

	// Outcome indicates success or failure
	Outcome string `json:"outcome"`

	// Reason provides justification for the action
	Reason string `json:"reason,omitempty"`
}

// SecurityEvent represents a security-relevant event requiring special handling
type SecurityEvent struct {
	// Action describes the security event
	Action string

	// Severity indicates the security impact
	Severity Level

	// Status indicates success or failure
	Status string

	// UserAgent identifies the client (if applicable)
	UserAgent string

	// IPAddress records the source IP (if applicable)
	IPAddress string

	// Location provides geographic context (if available)
	Location string

	// Signatures lists any security signatures triggered
	Signatures []string

	// PolicyViolations lists any security policies violated
	PolicyViolations []string

	// RiskScore provides a normalized risk assessment
	RiskScore float64

	// Mitigations describes any automatic responses taken
	Mitigations []string
}

// CreateAuditEvent creates a new audit trail event
func (s *Service) CreateAuditEvent(ctx context.Context, tenantID string, metadata AuditMetadata, opts ...EventOption) error {
	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	// Create base event
	event := New(tenantID, EventAudit, LevelInfo, buildAuditMessage(metadata))

	// Add audit-specific fields
	event.WithMetadata(metadataJSON)

	// Add timestamp for audit trail
	event.Timestamp = time.Now().UTC()

	// Apply any additional options
	for _, opt := range opts {
		if err := opt(event); err != nil {
			return err
		}
	}

	// Store the event
	return s.store.Store(ctx, event)
}

// CreateSecurityEvent creates a new security event with appropriate metadata
func (s *Service) CreateSecurityEvent(ctx context.Context, tenantID string, secEvent SecurityEvent, opts ...EventOption) error {
	// Convert security event to JSON
	metadataJSON, err := json.Marshal(secEvent)
	if err != nil {
		return err
	}

	// Create base event with appropriate severity
	event := New(tenantID, EventSecurity, secEvent.Severity, buildSecurityMessage(secEvent))

	// Add security-specific fields
	event.WithMetadata(metadataJSON)

	// Apply any additional options
	for _, opt := range opts {
		if err := opt(event); err != nil {
			return err
		}
	}

	// Store the event with high priority
	return s.store.Store(ctx, event)
}

// buildAuditMessage creates a human-readable audit message
func buildAuditMessage(metadata AuditMetadata) string {
	return string(metadata.Action) + " " + metadata.ResourceType + " " + metadata.ResourceID
}

// buildSecurityMessage creates a human-readable security event message
func buildSecurityMessage(event SecurityEvent) string {
	return event.Action + " (" + event.Status + ")"
}
