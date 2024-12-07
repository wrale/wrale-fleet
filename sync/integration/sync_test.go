package integration

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/sync/manager"
	"github.com/stretchr/testify/assert"
)

func setupTestSyncManager() *manager.SyncManager {
	config := manager.Config{
		StoragePath:   "/tmp/test-sync",
		MaxRetries:    3,
		Timeout:       5 * time.Second,
		RetryInterval: time.Second,
	}

	syncManager, err := manager.New(config)
	if err != nil {
		panic(err)
	}
	return syncManager
}

func TestSyncManager(t *testing.T) {
	sm := setupTestSyncManager()
	assert.NotNil(t, sm)
}