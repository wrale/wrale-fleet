package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/wrale-fleet/sync/manager"
)

func setupTestSyncManager(t *testing.T) *manager.SyncManager {
	t.Helper()
	
	config := manager.Config{
		StoragePath:   "/tmp/test-sync",
		MaxRetries:    3,
		Timeout:       5 * time.Second,
		RetryInterval: time.Second,
	}

	syncManager, err := manager.New(config)
	require.NoError(t, err, "failed to create sync manager")
	require.NotNil(t, syncManager, "sync manager should not be nil")
	
	return syncManager
}

func TestSyncManager(t *testing.T) {
	sm := setupTestSyncManager(t)
	assert.NotNil(t, sm, "sync manager should be created successfully")
}

func TestSyncManagerWithEmptyConfig(t *testing.T) {
	config := manager.Config{}
	sm, err := manager.New(config)
	
	assert.Error(t, err, "should error with empty config")
	assert.Nil(t, sm, "sync manager should be nil with invalid config")
}