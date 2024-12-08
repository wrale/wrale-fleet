package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
)

func TestNew(t *testing.T) {
	store := New()
	// Test store initialization by creating and retrieving a device
	dev := &device.Device{
		ID:       "test-init",
		TenantID: "tenant-init",
		Name:     "Test Init Device",
	}
	ctx := context.Background()
	err := store.Create(ctx, dev)
	require.NoError(t, err)
	
	retrieved, err := store.Get(ctx, dev.TenantID, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, dev.ID, retrieved.ID)
}

// Rest of the file stays the same...
