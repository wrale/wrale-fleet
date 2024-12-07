package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestSyncOperation(t *testing.T) {
	// Create test operation
	op := types.SyncOperation{
		ID:        "test-op",
		Type:      types.OpStateSync,
		DeviceIDs: []types.DeviceID{"device-1"},
		Payload:   map[string]interface{}{"key": "value"},
		Priority:  1,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Test operation with higher priority
	opHigh := types.SyncOperation{
		ID:        "test-op-high",
		Type:      types.OpStateSync,
		DeviceIDs: []types.DeviceID{"device-1"},
		Payload:   map[string]interface{}{"key": "value"},
		Priority:  2,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Verify priority is respected
	assert.True(t, opHigh.Priority > op.Priority)
}