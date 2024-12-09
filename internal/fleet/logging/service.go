package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Service provides logging operations with multi-tenant isolation.
// It handles event logging, audit trails, and retention policies while
// ensuring strict tenant boundaries are maintained.
type Service struct {
	store           Store
	logger          *zap.Logger
	mu              sync.RWMutex
	bufferSize      int
	retentionPolicy map[EventType]time.Duration
}

// ServiceOption is a functional option for configuring the service
type ServiceOption func(*Service) error

// WithBufferSize sets the event buffer size for batch operations
func WithBufferSize(size int) ServiceOption {
	return func(s *Service) error {
		if size < 0 {
			return fmt.Errorf("buffer size cannot be negative")
		}
		s.bufferSize = size
		return nil
	}
}

// WithRetentionPolicy sets the retention duration for an event type
func WithRetentionPolicy(eventType EventType, duration time.Duration) ServiceOption {
	return func(s *Service) error {
		if duration < 0 {
			return fmt.Errorf("retention duration cannot be negative")
		}
		s.retentionPolicy[eventType] = duration
		return nil
	}
}

// NewService creates a new logging service with the provided store and logger
func NewService(store Store, logger *zap.Logger, opts ...ServiceOption) (*Service, error) {
	if store == nil {
		return nil, ErrStoreNotInitialized
	}

	s := &Service{
		store:           store,
		logger:          logger,
		bufferSize:      1000, // Default buffer size
		retentionPolicy: make(map[EventType]time.Duration),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	// Set default retention policies if not configured
	if len(s.retentionPolicy) == 0 {
		s.retentionPolicy = map[EventType]time.Duration{
			EventSystem:      30 * 24 * time.Hour,  // 30 days
			EventSecurity:    90 * 24 * time.Hour,  // 90 days
			EventAudit:       365 * 24 * time.Hour, // 1 year
			EventCompliance:  730 * 24 * time.Hour, // 2 years
			EventOperational: 7 * 24 * time.Hour,   // 7 days
		}
	}

	return s, nil
}

// Log creates and stores a new log event
func (s *Service) Log(ctx context.Context, tenantID string, eventType EventType, level Level, message string, opts ...EventOption) error {
	event := New(tenantID, eventType, level, message)

	// Apply options
	for _, opt := range opts {
		if err := opt(event); err != nil {
			return fmt.Errorf("invalid event option: %w", err)
		}
	}

	// Set retention policy
	if event.RetentionPolicy == "" {
		if duration, ok := s.retentionPolicy[eventType]; ok {
			event.RetentionPolicy = duration.String()
		}
	}

	// Store the event
	if err := s.store.Store(ctx, event); err != nil {
		return fmt.Errorf("storing event: %w", err)
	}

	// Log to infrastructure logger as well
	s.logToInfrastructure(event)

	return nil
}

// BatchLog stores multiple events efficiently
func (s *Service) BatchLog(ctx context.Context, events []*Event) error {
	// Validate and prepare events
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return fmt.Errorf("validating event: %w", err)
		}

		// Set retention policy if not specified
		if event.RetentionPolicy == "" {
			if duration, ok := s.retentionPolicy[event.Type]; ok {
				event.RetentionPolicy = duration.String()
			}
		}
	}

	// Store events in batch
	if err := s.store.BatchStore(ctx, events); err != nil {
		return fmt.Errorf("batch storing events: %w", err)
	}

	// Log to infrastructure logger
	for _, event := range events {
		s.logToInfrastructure(event)
	}

	return nil
}

// Query performs a structured query on events
func (s *Service) Query(ctx context.Context, query QueryOptions) ([]*Event, error) {
	return s.store.Query(ctx, query)
}

// Retention enforces retention policies by removing expired events
func (s *Service) Retention(ctx context.Context, tenantID string) error {
	for eventType, duration := range s.retentionPolicy {
		cutoff := time.Now().UTC().Add(-duration)
		if err := s.store.DeleteBefore(ctx, tenantID, cutoff); err != nil {
			return fmt.Errorf("enforcing retention for %s: %w", eventType, err)
		}
	}
	return nil
}

// Sync ensures all events are durably stored
func (s *Service) Sync(ctx context.Context) error {
	return s.store.Sync(ctx)
}

// logToInfrastructure logs events to the infrastructure logger
func (s *Service) logToInfrastructure(event *Event) {
	var fields []zap.Field
	fields = append(fields,
		zap.String("event_id", event.ID),
		zap.String("tenant_id", event.TenantID),
		zap.String("type", string(event.Type)),
		zap.Time("timestamp", event.Timestamp),
	)

	if event.Context.ComponentID != "" {
		fields = append(fields, zap.String("component_id", event.Context.ComponentID))
	}
	if event.Context.DeviceID != "" {
		fields = append(fields, zap.String("device_id", event.Context.DeviceID))
	}
	if event.Context.UserID != "" {
		fields = append(fields, zap.String("user_id", event.Context.UserID))
	}
	if event.Context.RequestID != "" {
		fields = append(fields, zap.String("request_id", event.Context.RequestID))
	}
	if event.Source != "" {
		fields = append(fields, zap.String("source", event.Source))
	}

	// Include metadata if present
	if len(event.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(event.Metadata, &metadata); err == nil {
			fields = append(fields, zap.Any("metadata", metadata))
		}
	}

	switch event.Level {
	case LevelError:
		s.logger.Error(event.Message, fields...)
	case LevelWarn:
		s.logger.Warn(event.Message, fields...)
	case LevelInfo:
		s.logger.Info(event.Message, fields...)
	case LevelDebug:
		s.logger.Debug(event.Message, fields...)
	}
}

// EventOption configures an Event
type EventOption func(*Event) error

// WithEventContext adds context to an event
func WithEventContext(ctx EventContext) EventOption {
	return func(e *Event) error {
		e.Context = ctx
		return nil
	}
}

// WithEventMetadata adds metadata to an event
func WithEventMetadata(metadata interface{}) EventOption {
	return func(e *Event) error {
		return e.WithMetadata(metadata)
	}
}

// WithEventTag adds a tag to an event
func WithEventTag(key, value string) EventOption {
	return func(e *Event) error {
		e.WithTag(key, value)
		return nil
	}
}

// WithEventSource sets the event source
func WithEventSource(source string) EventOption {
	return func(e *Event) error {
		e.WithSource(source)
		return nil
	}
}

// WithEventRetention sets the event retention policy
func WithEventRetention(policy string) EventOption {
	return func(e *Event) error {
		e.WithRetention(policy)
		return nil
	}
}
