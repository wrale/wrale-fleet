package logging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
)

func TestService_Log(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.NewTestStore()
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
	store := memory.NewTestStore()
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

func TestService_RetentionPolicy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.NewTestStore()

	// Create service with short retention for testing
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
	results, err := service.Query(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "new event", results[0].Message)
}

func TestService_QueryFilters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.NewTestStore()
	service, err := NewService(store, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Create events with different characteristics
	events := []*Event{
		New("tenant1", EventSystem, LevelInfo, "event1").
			WithTag("env", "prod"),
		New("tenant1", EventSecurity, LevelError, "event2").
			WithTag("env", "prod").
			WithSource("server1"),
		New("tenant1", EventSystem, LevelWarn, "event3").
			WithTag("env", "dev"),
	}

	err = service.BatchLog(ctx, events)
	require.NoError(t, err)

	tests := []struct {
		name      string
		query     QueryOptions
		wantCount int
	}{
		{
			name: "filter by type",
			query: QueryOptions{
				TenantID: "tenant1",
				Types:    []EventType{EventSystem},
			},
			wantCount: 2,
		},
		{
			name: "filter by level",
			query: QueryOptions{
				TenantID: "tenant1",
				Levels:   []Level{LevelError},
			},
			wantCount: 1,
		},
		{
			name: "filter by source",
			query: QueryOptions{
				TenantID: "tenant1",
				Sources:  []string{"server1"},
			},
			wantCount: 1,
		},
		{
			name: "filter by tag",
			query: QueryOptions{
				TenantID: "tenant1",
				TagQuery: &TagQuery{
					Must: map[string]string{"env": "prod"},
				},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.Query(ctx, tt.query)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestStage_WithLogger(t *testing.T) {
	tests := []struct {
		name          string
		stage         int
		operation     string
		requiredStage int
		wantAllowed   bool
	}{
		{
			name:          "stage 1 operation always allowed",
			stage:         1,
			operation:     "basic_op",
			requiredStage: 1,
			wantAllowed:   true,
		},
		{
			name:          "higher stage operation not allowed",
			stage:         1,
			operation:     "advanced_op",
			requiredStage: 2,
			wantAllowed:   false,
		},
		{
			name:          "operation at current stage allowed",
			stage:         3,
			operation:     "current_op",
			requiredStage: 3,
			wantAllowed:   true,
		},
		{
			name:          "invalid stage not allowed",
			stage:         1,
			operation:     "invalid_op",
			requiredStage: MaxStage + 1,
			wantAllowed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			stagedLogger := WithStage(logger, tt.stage)

			allowed := StageCheck(stagedLogger, tt.requiredStage, tt.operation)
			assert.Equal(t, tt.wantAllowed, allowed)

			stage := GetStage(stagedLogger)
			assert.Equal(t, tt.stage, stage)
		})
	}
}
