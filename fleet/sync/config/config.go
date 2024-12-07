package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

// Manager handles configuration management and distribution
type Manager struct {
	mu sync.RWMutex

	// Track configurations
	configs map[string]*synctypes.ConfigData

	// Track device configs
	deviceConfigs map[types.DeviceID]string // DeviceID -> ConfigVersion

	// Track config distribution status
	distribution map[string]map[types.DeviceID]bool // ConfigVersion -> {DeviceID -> Distributed}
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{
		configs:       make(map[string]*synctypes.ConfigData),
		deviceConfigs: make(map[types.DeviceID]string),
		distribution:  make(map[string]map[types.DeviceID]bool),
	}
}

// GetConfig retrieves a configuration by version
func (m *Manager) GetConfig(version string) (*synctypes.ConfigData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[version]
	if !exists {
		return nil, fmt.Errorf("config version not found: %s", version)
	}

	return config, nil
}

// UpdateConfig stores a new configuration version
func (m *Manager) UpdateConfig(config *synctypes.ConfigData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate config
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Generate version if not provided
	if config.Version == "" {
		config.Version = generateConfigVersion(config)
	}

	// Set validity period if not specified
	if config.ValidFrom.IsZero() {
		config.ValidFrom = time.Now()
	}

	// Store config
	m.configs[config.Version] = config

	// Initialize distribution tracking
	m.distribution[config.Version] = make(map[types.DeviceID]bool)

	return nil
}

// ListConfigs returns all available configurations
func (m *Manager) ListConfigs() ([]*synctypes.ConfigData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]*synctypes.ConfigData, 0, len(m.configs))
	for _, config := range m.configs {
		configs = append(configs, config)
	}

	return configs, nil
}

// DistributeConfig distributes a configuration to devices
func (m *Manager) DistributeConfig(config *synctypes.ConfigData, devices []types.DeviceID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify config exists
	if _, exists := m.configs[config.Version]; !exists {
		return fmt.Errorf("config version not found: %s", config.Version)
	}

	// Track distribution
	for _, deviceID := range devices {
		// Update device's current config
		m.deviceConfigs[deviceID] = config.Version

		// Mark as distributed
		if _, exists := m.distribution[config.Version]; !exists {
			m.distribution[config.Version] = make(map[types.DeviceID]bool)
		}
		m.distribution[config.Version][deviceID] = true
	}

	return nil
}

// GetDeviceConfig gets current config for a device
func (m *Manager) GetDeviceConfig(deviceID types.DeviceID) (*synctypes.ConfigData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get device's current config version
	version, exists := m.deviceConfigs[deviceID]
	if !exists {
		return nil, fmt.Errorf("no config found for device: %s", deviceID)
	}

	// Get the config
	config, exists := m.configs[version]
	if !exists {
		return nil, fmt.Errorf("config version not found: %s", version)
	}

	return config, nil
}

// GetDistributionStatus gets config distribution status
func (m *Manager) GetDistributionStatus(version string) (map[types.DeviceID]bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.distribution[version]
	if !exists {
		return nil, fmt.Errorf("config version not found: %s", version)
	}

	// Return copy of status map
	result := make(map[types.DeviceID]bool)
	for id, distributed := range status {
		result[id] = distributed
	}

	return result, nil
}

// IsConfigValid checks if a config is currently valid
func (m *Manager) IsConfigValid(version string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	config, exists := m.configs[version]
	if !exists {
		return false, fmt.Errorf("config version not found: %s", version)
	}

	now := time.Now()
	if now.Before(config.ValidFrom) {
		return false, nil
	}
	if config.ValidTo != nil && now.After(*config.ValidTo) {
		return false, nil
	}

	return true, nil
}

// validateConfig validates configuration data
func validateConfig(config *synctypes.ConfigData) error {
	if config == nil {
		return fmt.Errorf("nil config")
	}

	if config.Config == nil {
		return fmt.Errorf("empty config data")
	}

	// For v1.0, we do basic validation
	// Can be expanded based on specific needs
	return nil
}

// generateConfigVersion generates a version string for a config
func generateConfigVersion(config *synctypes.ConfigData) string {
	hasher := sha256.New()

	// Add config data to hash
	if data, err := json.Marshal(config.Config); err == nil {
		hasher.Write(data)
	}

	// Add timestamp to hash
	hasher.Write([]byte(time.Now().String()))

	// Generate version string
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("config-%s", hash[:8])
}
