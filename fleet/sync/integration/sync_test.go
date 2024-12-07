package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/sync/manager"
	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

func setupTestSyncManager() *manager.SyncManager {
	config := manager.Config{
		StoragePath: "/tmp/test-sync",
		RetryInterval: time.Second,
		MaxRetries: 3,
	}
	syncManager, err := manager.New(config)
	if err != nil {
		panic(err) // In test setup, panic is acceptable
	}
	return syncManager
}

func TestConfigSync(t *testing.T) {
	syncManager := setupTestSyncManager()
	
	config := &types.ConfigData{
		Version:   "1.0",
		Config:    map[string]interface{}{"key": "value"},
		ValidFrom: time.Now(),
	}

	err := syncManager.UpdateConfig(config)
	assert.NoError(t, err)

	devices := []types.DeviceID{"device-1", "device-2"}
	err = syncManager.DistributeConfig(config, devices)
	assert.NoError(t, err)

	deviceConfig, err := syncManager.GetDeviceConfig("device-1")
	assert.NoError(t, err)
	assert.Equal(t, config.Version, deviceConfig.Version)
}

func TestConfigSyncFailure(t *testing.T) {
	syncManager := setupTestSyncManager()
	
	// Test with invalid config
	config := &types.ConfigData{
		Version:   "1.0",
		Config:    nil,
		ValidFrom: time.Now(),
	}

	err := syncManager.UpdateConfig(config)
	assert.Error(t, err)
}

func TestDeviceConfigSync(t *testing.T) {
	syncManager := setupTestSyncManager()
	
	config := &types.ConfigData{
		Version:   "1.0",
		Config:    map[string]interface{}{"setting": "test"},
		ValidFrom: time.Now(),
	}

	// Test single device config update
	err := syncManager.UpdateDeviceConfig("test-device", config)
	assert.NoError(t, err)

	retrieved, err := syncManager.GetDeviceConfig("test-device")
	assert.NoError(t, err)
	assert.Equal(t, config.Config, retrieved.Config)
}