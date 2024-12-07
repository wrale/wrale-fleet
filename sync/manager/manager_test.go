package manager_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err, "failed to create sync manager")
	assert.NotNil(t, syncManager, "sync manager should not be nil")

	return syncManager
}

func TestSyncManager_Creation(t *testing.T) {
	sm := setupTestSyncManager(t)
	assert.NotNil(t, sm, "sync manager should be created successfully")
}

func TestSyncManager_Configuration(t *testing.T) {
	tests := []struct {
		name        string
		config      manager.Config
		expectError bool
	}{
		{
			name:        "empty config",
			config:      manager.Config{},
			expectError: true,
		},
		{
			name: "minimal valid config",
			config: manager.Config{
				StoragePath: "/tmp/test",
			},
			expectError: false,
		},
		{
			name: "full valid config",
			config: manager.Config{
				StoragePath:   "/tmp/test",
				MaxRetries:    5,
				Timeout:       10 * time.Second,
				RetryInterval: time.Second,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := manager.New(tt.config)

			if tt.expectError {
				assert.Error(t, err, "should error with invalid config")
				assert.Nil(t, sm, "sync manager should be nil with invalid config")
			} else {
				assert.NoError(t, err, "should not error with valid config")
				assert.NotNil(t, sm, "sync manager should not be nil with valid config")
			}
		})
	}
}