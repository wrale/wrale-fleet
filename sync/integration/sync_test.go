package integration

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/sync/manager"
)

func setupTestSyncManager() (*manager.SyncManager, error) {
	config := manager.Config{
		StoragePath:   "/tmp/test-sync",
		RetryInterval: time.Second,
		MaxRetries:    3,
		Timeout:       5 * time.Second,
	}

	return manager.New(config)
}

func TestSyncManager(t *testing.T) {
	syncMgr, err := setupTestSyncManager()
	if err != nil {
		t.Fatalf("Failed to create sync manager: %v", err)
	}

	t.Run("Basic Operations", func(t *testing.T) {
		if syncMgr == nil {
			t.Fatal("Sync manager should not be nil")
		}
	})
}