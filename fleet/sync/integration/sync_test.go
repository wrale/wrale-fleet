package integration

import (
	"testing"
	"time"
	
	"github.com/wrale/wrale-fleet/fleet/sync/types"
	"github.com/stretchr/testify/assert"
)

func TestConfigSync(t *testing.T) {
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