package logging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestService_Log(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := NewService(store, logger)
	require.NoError(t, err)

	tests := []struct {
		name      string
		tenantID  string
		eventType EventType
		level     Level
		message   string
		opts      []EventOption
		wantErr   bool
	}{
		{
			name:      "basic event",
			tenantID:  "tenant1",
			eventType: EventSystem,
			level:     LevelInfo,
			message:   "test event",
			wantErr:   false,
		},
		{
			name:      "with context",
			tenantID:  "tenant1",
			eventType: EventOperational,
			level:     LevelInfo,
			message:   "context test",
			opts: []EventOption{
				WithEventContext(EventContext{
					ComponentID: "test-component",
					DeviceID:    "test-device",
				}),
			},
			wantErr: false,
		},
		{
			name:      "with metadata",
			tenantID:  "tenant1",
			eventType: EventAudit,
			level:     LevelInfo,
			message:   "metadata test",
			opts: []EventOption{
				WithEventMetadata(map[string]string{
					"key": "value",
				}),
			},
			wantErr: false,
		},
		{
			name:      "missing tenant",
			tenantID:  "",
			eventType: EventSystem,
			level:     LevelInfo,
			message:   "invalid event",
			wantErr:   true,
		},
		{
			name:      "missing message",
			tenantID:  "tenant1",
			eventType: EventSystem,
			level:     LevelInfo,
			message:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.Log(ctx, tt.tenantID, tt.eventType, tt.level, tt.message, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Query to verify event was stored
			query := QueryOptions{
				TenantID: tt.tenantID,
				Types:    []EventType{tt.eventType},
				TimeRange: &TimeRange{
					Start: time.Now().Add(-time.Minute),
					End:   time.Now().Add(time.Minute),
				},
			}
			events, err := service.Query(ctx, query)
			require.NoError(t, err)
			require.Len(t, events, 1)

			event := events[0]
			assert.Equal(t, tt.tenantID, event.TenantID)
			assert.Equal(t, tt.eventType, event.Type)
			assert.Equal(t, tt.level, event.Level)
			assert.Equal(t, tt.message, event.Message)
		})
	}
}

func TestService_BatchLog(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := NewService(store, logger)
	require.NoError(t, err)

	ctx := context.Background()
	events := []*Event{
		New("tenant1", EventSystem, LevelInfo, "event1"),
		New("tenant1", EventOperational, LevelWarn, "event2"),
		New("tenant2", EventAudit, LevelError, "event3"),
	}

	err = service.BatchLog(ctx, events)
	require.NoError(t, err)

	// Verify events for tenant1
	query := QueryOptions{
		TenantID: "tenant1",
		TimeRange: &TimeRange{
			Start: time.Now().Add(-time.Minute),
			End:   time.Now().Add(time.Minute),
		},
	}
	results, err := service.Query(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Verify events for tenant2
	query.TenantID = "tenant2"
	results, err = service.Query(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestService_AuditEvents(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := NewService(store, logger)
	require.NoError(t, err)

	ctx := context.Background()
	metadata := AuditMetadata{
		Action:       AuditActionCreate,
		ResourceType: "device",
		ResourceID:   "dev1",
		Outcome:      "success",
		Changes: map[string]interface{}{
			"name": "new device",
		},
	}

	err = service.CreateAuditEvent(ctx, "tenant1", metadata)
	require.NoError(t, err)

	// Verify audit event
	query := QueryOptions{
		TenantID: "tenant1",
		Types:    []EventType{EventAudit},
	}
	events, err := service.Query(ctx, query)
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, EventAudit, event.Type)
	assert.NotEmpty(t, event.Metadata)
}

func TestService_SecurityEvents(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := NewService(store, logger)
	require.NoError(t, err)

	ctx := context.Background()
	secEvent := SecurityEvent{
		Action:           "unauthorized_access",
		Severity:         LevelError,
		Status:           "blocked",
		IPAddress:        "192.168.1.1",
		PolicyViolations: []string{"invalid_cert"},
		RiskScore:        0.8,
	}

	err = service.CreateSecurityEvent(ctx, "tenant1", secEvent)
	require.NoError(t, err)

	// Verify security event
	query := QueryOptions{
		TenantID: "tenant1",
		Types:    []EventType{EventSecurity},
		Levels:   []Level{LevelError},
	}
	events, err := service.Query(ctx, query)
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, EventSecurity, event.Type)
	assert.Equal(t, LevelError, event.Level)
	assert.NotEmpty(t, event.Metadata)
}

func TestService_Retention(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := NewService(store, logger,
		WithRetentionPolicy(EventOperational, time.Hour))
	require.NoError(t, err)

	ctx := context.Background()

	// Create old and new events
	oldEvent := New("tenant1", EventOperational, LevelInfo, "old event")
	oldEvent.Timestamp = time.Now().Add(-2 * time.Hour)
	newEvent := New("tenant1", EventOperational, LevelInfo, "new event")

	err = service.BatchLog(ctx, []*Event{oldEvent, newEvent})
	require.NoError(t, err)

	// Run retention
	err = service.Retention(ctx, "tenant1")
	require.NoError(t, err)

	// Verify only new event remains
	query := QueryOptions{
		TenantID: "tenant1",
		Types:    []EventType{EventOperational},
	}
	events, err := service.Query(ctx, query)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "new event", events[0].Message)
}
