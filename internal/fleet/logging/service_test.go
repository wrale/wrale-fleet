package logging_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	loggingtest "github.com/wrale/wrale-fleet/internal/fleet/logging/testing"
)

func TestService_Log(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := loggingtest.NewTestService(t)

	tests := []struct {
		name      string
		tenantID  string
		eventType logging.EventType
		level     logging.Level
		message   string
		opts      []logging.EventOption
		wantErr   bool
	}{
		{
			name:      "basic event",
			tenantID:  "tenant1",
			eventType: logging.EventSystem,
			level:     logging.LevelInfo,
			message:   "test event",
			wantErr:   false,
		},
		{
			name:      "with context",
			tenantID:  "tenant1",
			eventType: logging.EventOperational,
			level:     logging.LevelInfo,
			message:   "context test",
			opts: []logging.EventOption{
				logging.WithEventContext(logging.EventContext{
					ComponentID: "test-component",
					DeviceID:    "test-device",
				}),
			},
			wantErr: false,
		},
		{
			name:      "with metadata",
			tenantID:  "tenant1",
			eventType: logging.EventAudit,
			level:     logging.LevelInfo,
			message:   "metadata test",
			opts: []logging.EventOption{
				logging.WithEventMetadata(map[string]string{
					"key": "value",
				}),
			},
			wantErr: false,
		},
		{
			name:      "missing tenant",
			tenantID:  "",
			eventType: logging.EventSystem,
			level:     logging.LevelInfo,
			message:   "invalid event",
			wantErr:   true,
		},
		{
			name:      "missing message",
			tenantID:  "tenant1",
			eventType: logging.EventSystem,
			level:     logging.LevelInfo,
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
			query := logging.QueryOptions{
				TenantID: tt.tenantID,
				Types:    []logging.EventType{tt.eventType},
				TimeRange: &logging.TimeRange{
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
	service := loggingtest.NewTestService(t)

	ctx := context.Background()
	events := []*logging.Event{
		logging.New("tenant1", logging.EventSystem, logging.LevelInfo, "event1"),
		logging.New("tenant1", logging.EventOperational, logging.LevelWarn, "event2"),
		logging.New("tenant2", logging.EventAudit, logging.LevelError, "event3"),
	}

	err := service.BatchLog(ctx, events)
	require.NoError(t, err)

	// Verify events for tenant1
	query := logging.QueryOptions{
		TenantID: "tenant1",
		TimeRange: &logging.TimeRange{
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
	service := loggingtest.NewTestService(t)

	ctx := context.Background()

	// Create old and new events
	oldEvent := logging.New("tenant1", logging.EventOperational, logging.LevelInfo, "old event")
	oldEvent.Timestamp = time.Now().Add(-2 * time.Hour)
	newEvent := logging.New("tenant1", logging.EventOperational, logging.LevelInfo, "new event")

	err := service.BatchLog(ctx, []*logging.Event{oldEvent, newEvent})
	require.NoError(t, err)

	// Run retention with 1-hour policy
	err = service.ApplyRetentionPolicy(ctx, "tenant1", time.Hour)
	require.NoError(t, err)

	// Verify only new event remains
	query := logging.QueryOptions{
		TenantID: "tenant1",
		Types:    []logging.EventType{logging.EventOperational},
	}
	results, err := service.Query(ctx, query)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "new event", results[0].Message)
}

func TestService_QueryFilters(t *testing.T) {
	service := loggingtest.NewTestService(t)

	ctx := context.Background()

	// Create events with different characteristics
	events := []*logging.Event{
		logging.New("tenant1", logging.EventSystem, logging.LevelInfo, "event1").
			WithTag("env", "prod"),
		logging.New("tenant1", logging.EventSecurity, logging.LevelError, "event2").
			WithTag("env", "prod").
			WithSource("server1"),
		logging.New("tenant1", logging.EventSystem, logging.LevelWarn, "event3").
			WithTag("env", "dev"),
	}

	err := service.BatchLog(ctx, events)
	require.NoError(t, err)

	tests := []struct {
		name      string
		query     logging.QueryOptions
		wantCount int
	}{
		{
			name: "filter by type",
			query: logging.QueryOptions{
				TenantID: "tenant1",
				Types:    []logging.EventType{logging.EventSystem},
			},
			wantCount: 2,
		},
		{
			name: "filter by level",
			query: logging.QueryOptions{
				TenantID: "tenant1",
				Levels:   []logging.Level{logging.LevelError},
			},
			wantCount: 1,
		},
		{
			name: "filter by source",
			query: logging.QueryOptions{
				TenantID: "tenant1",
				Sources:  []string{"server1"},
			},
			wantCount: 1,
		},
		{
			name: "filter by tag",
			query: logging.QueryOptions{
				TenantID: "tenant1",
				TagQuery: &logging.TagQuery{
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
			requiredStage: logging.MaxStage + 1,
			wantAllowed:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			stagedLogger := logging.WithStage(logger, tt.stage)

			allowed := logging.StageCheck(stagedLogger, tt.requiredStage, tt.operation)
			assert.Equal(t, tt.wantAllowed, allowed)

			stage := logging.GetStage(stagedLogger)
			assert.Equal(t, tt.stage, stage)
		})
	}
}
