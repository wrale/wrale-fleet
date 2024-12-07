package config

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestConfigManager(t *testing.T) {
	manager := NewManager()

	t.Run("Basic Config Operations", func(t *testing.T) {
		// Create test config
		config := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"update_interval": 30,
				"max_retries":     5,
			},
			ValidFrom: time.Now(),
		}

		// Update config
		if err := manager.UpdateConfig(config); err != nil {
			t.Errorf("Failed to update config: %v", err)
		}

		// Version should be generated
		if config.Version == "" {
			t.Error("Config version not generated")
		}

		// Get config
		retrieved, err := manager.GetConfig(config.Version)
		if err != nil {
			t.Errorf("Failed to get config: %v", err)
		}
		if retrieved.Version != config.Version {
			t.Error("Config version mismatch")
		}

		// List configs
		configs, err := manager.ListConfigs()
		if err != nil {
			t.Errorf("Failed to list configs: %v", err)
		}
		if len(configs) != 1 {
			t.Errorf("Expected 1 config, got %d", len(configs))
		}
	})

	t.Run("Config Distribution", func(t *testing.T) {
		config := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"key": "value",
			},
			ValidFrom: time.Now(),
		}

		// Store config
		if err := manager.UpdateConfig(config); err != nil {
			t.Errorf("Failed to update config: %v", err)
		}

		// Define test devices
		devices := []types.DeviceID{
			"device-1",
			"device-2",
		}

		// Distribute config
		if err := manager.DistributeConfig(config, devices); err != nil {
			t.Errorf("Failed to distribute config: %v", err)
		}

		// Check distribution status
		status, err := manager.GetDistributionStatus(config.Version)
		if err != nil {
			t.Errorf("Failed to get distribution status: %v", err)
		}

		// Verify all devices received config
		for _, deviceID := range devices {
			if distributed, exists := status[deviceID]; !exists || !distributed {
				t.Errorf("Config not marked as distributed for device %s", deviceID)
			}
		}

		// Check device config
		for _, deviceID := range devices {
			deviceConfig, err := manager.GetDeviceConfig(deviceID)
			if err != nil {
				t.Errorf("Failed to get device config: %v", err)
			}
			if deviceConfig.Version != config.Version {
				t.Errorf("Device config version mismatch for device %s", deviceID)
			}
		}
	})

	t.Run("Config Validation", func(t *testing.T) {
		// Test valid config
		validConfig := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"key": "value",
			},
			ValidFrom: time.Now(),
		}

		if err := manager.UpdateConfig(validConfig); err != nil {
			t.Errorf("Failed to update valid config: %v", err)
		}

		// Test invalid configs
		invalidConfigs := []struct {
			name   string
			config *synctypes.ConfigData
		}{
			{
				name:   "nil config",
				config: nil,
			},
			{
				name: "empty config data",
				config: &synctypes.ConfigData{
					ValidFrom: time.Now(),
				},
			},
		}

		for _, tc := range invalidConfigs {
			t.Run(tc.name, func(t *testing.T) {
				if err := manager.UpdateConfig(tc.config); err == nil {
					t.Error("Expected error for invalid config")
				}
			})
		}
	})

	t.Run("Config Validity Period", func(t *testing.T) {
		now := time.Now()
		futureTime := now.Add(time.Hour)
		pastTime := now.Add(-time.Hour)

		// Test current config
		currentConfig := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"key": "value",
			},
			ValidFrom: pastTime,
		}

		if err := manager.UpdateConfig(currentConfig); err != nil {
			t.Errorf("Failed to update current config: %v", err)
		}

		valid, err := manager.IsConfigValid(currentConfig.Version)
		if err != nil {
			t.Errorf("Failed to check config validity: %v", err)
		}
		if !valid {
			t.Error("Current config should be valid")
		}

		// Test future config
		futureConfig := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"key": "value",
			},
			ValidFrom: futureTime,
		}

		if err := manager.UpdateConfig(futureConfig); err != nil {
			t.Errorf("Failed to update future config: %v", err)
		}

		valid, err = manager.IsConfigValid(futureConfig.Version)
		if err != nil {
			t.Errorf("Failed to check config validity: %v", err)
		}
		if valid {
			t.Error("Future config should not be valid yet")
		}

		// Test expired config
		validTo := pastTime
		expiredConfig := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"key": "value",
			},
			ValidFrom: pastTime.Add(-time.Hour),
			ValidTo:   &validTo,
		}

		if err := manager.UpdateConfig(expiredConfig); err != nil {
			t.Errorf("Failed to update expired config: %v", err)
		}

		valid, err = manager.IsConfigValid(expiredConfig.Version)
		if err != nil {
			t.Errorf("Failed to check config validity: %v", err)
		}
		if valid {
			t.Error("Expired config should not be valid")
		}
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test getting non-existent config
		if _, err := manager.GetConfig("non-existent"); err == nil {
			t.Error("Expected error getting non-existent config")
		}

		// Test getting non-existent device config
		if _, err := manager.GetDeviceConfig("non-existent-device"); err == nil {
			t.Error("Expected error getting non-existent device config")
		}

		// Test distributing non-existent config
		nonExistentConfig := &synctypes.ConfigData{
			Version: "non-existent",
			Config:  map[string]interface{}{},
		}
		if err := manager.DistributeConfig(nonExistentConfig, []types.DeviceID{"device-1"}); err == nil {
			t.Error("Expected error distributing non-existent config")
		}
	})
}
