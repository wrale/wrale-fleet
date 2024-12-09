package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
)

func createTestEvent(tenantID string, eventType logging.EventType, level logging.Level, message string) *logging.Event {
	return logging.New(tenantID, eventType, level, message)
}

func TestStore_Store(t *testing.T) {
	store := New()
	ctx := context.Background()

	event := createTestEvent("tenant1", logging.EventSystem, logging.LevelInfo, "test event")
	require.NoError(t, store.Store(ctx, event))

	// Verify event was stored
	tenant, exists := store.events[event.TenantID]
	require.True(t, exists)
	require.Contains(t, tenant, event.ID)
	assert.Equal(t, event, tenant[event.ID])
}

func TestStore_BatchStore(t *testing.T) {
	store := New()
	ctx := context.Background()

	events := []*logging.Event{
		createTestEvent("tenant1", logging.EventSystem, logging.LevelInfo, "event1"),
		createTestEvent("tenant1", logging.EventOperational, logging.LevelWarn, "event2"),
		createTestEvent("tenant2", logging.EventAudit, logging.LevelError, "event3"),
	}

	require.NoError(t, store.BatchStore(ctx, events))

	// Verify tenant1 events
	tenant1, exists := store.events["tenant1"]
	require.True(t, exists)
	assert.Len(t, tenant1, 2)

	// Verify tenant2 events
	tenant2, exists := store.events["tenant2"]
	require.True(t, exists)
	assert.Len(t, tenant2, 1)
}

func TestStore_DeleteBefore(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create events with different timestamps
	oldEvent := createTestEvent("tenant1", logging.EventSystem, logging.LevelInfo, "old event")
	oldEvent.Timestamp = time.Now().Add(-2 * time.Hour)

	newEvent := createTestEvent("tenant1", logging.EventSystem, logging.LevelInfo, "new event")

	require.NoError(t, store.BatchStore(ctx, []*logging.Event{oldEvent, newEvent}))

	// Delete old events
	require.NoError(t, store.DeleteBefore(ctx, "tenant1", time.Now().Add(-time.Hour)))

	// Verify only new event remains
	tenant, exists := store.events["tenant1"]
	require.True(t, exists)
	assert.Len(t, tenant, 1)
	_, exists = tenant[oldEvent.ID]
	assert.False(t, exists)
	_, exists = tenant[newEvent.ID]
	assert.True(t, exists)
}

func TestStore_Query(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create test events
	events := []*logging.Event{
		createTestEvent("tenant1", logging.EventSystem, logging.LevelInfo, "event1").
			WithTag("env", "prod"),
		createTestEvent("tenant1", logging.EventSecurity, logging.LevelError, "event2").
			WithTag("env", "prod").
			WithSource("server1"),
		createTestEvent("tenant1", logging.EventSystem, logging.LevelWarn, "event3").
			WithTag("env", "dev"),
	}

	require.NoError(t, store.BatchStore(ctx, events))

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
			results, err := store.Query(ctx, tt.query)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestStore_Sync(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Sync should always succeed for memory store
	assert.NoError(t, store.Sync(ctx))
}
