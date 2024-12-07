package integration

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/sync/manager"
)

func setupTestSyncManager() (*manager.SyncManager, error) {
	config := manager.Config{
		StoragePath:   "/tmp/test-sync",
		MaxRetries:    3,
		Timeout:       5 * time.Second,
		RetryInterval: time.Second,
	}
	return manager.New(config)
}

func TestSyncManager(t *testing.T) {
	syncManager, err := setupTestSyncManager()
	if err != nil {
		t.Fatalf("failed to create sync manager: %v", err)
	}
	if syncManager == nil {
		t.Fatal("sync manager is nil")
	}
}