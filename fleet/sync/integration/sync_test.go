package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
	"github.com/wrale/wrale-fleet/fleet/sync/manager"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestSyncIntegration(t *testing.T) {
	// Setup test manager and dependencies
	store := newTestStore()
	resolver := newTestResolver()
	config := newTestConfig()
	syncManager := manager.NewManager(store, resolver, config)

	// Test device state sync
	deviceID := types.DeviceID("test-device")
	state := &synctypes.VersionedState{
		Version: "v1",
		State: types.DeviceState{
			ID:       deviceID,
			Status:   "active",
			LastSeen: time.Now(),
		},
		UpdatedAt: time.Now(),
		UpdatedBy: "test",
	}

	// Update state
	err := syncManager.UpdateState(deviceID, state)
	assert.NoError(t, err)

	// Verify state was saved
	savedState, err := syncManager.GetState(deviceID)
	assert.NoError(t, err)
	assert.Equal(t, state.Version, savedState.Version)
	assert.Equal(t, state.State.ID, savedState.State.ID)

	// Test config sync
	config := &synctypes.ConfigData{
		Version: "v1",
		Config: map[string]interface{}{
			"key": "value",
		},
		ValidFrom: time.Now(),
	}

	// Update config
	err = syncManager.UpdateConfig(config)
	assert.NoError(t, err)

	// Distribute config
	devices := []types.DeviceID{deviceID}
	err = syncManager.DistributeConfig(config, devices)
	assert.NoError(t, err)

	// Verify device config
	deviceConfig, err := syncManager.GetDeviceConfig(deviceID)
	assert.NoError(t, err)
	assert.Equal(t, config.Version, deviceConfig.Version)

	// Test state validation
	err = syncManager.ValidateState(state.Version)
	assert.NoError(t, err)

	// Test invalid state validation
	invalidState := &synctypes.VersionedState{
		Version: "invalid",
		State: types.DeviceState{
			ID:       deviceID,
			Status:   "unknown",
			LastSeen: time.Now(),
		},
	}
	err = syncManager.ValidateState(invalidState.Version)
	assert.Error(t, err)

	// Test consensus tracking
	validator := "test-validator"
	err = syncManager.AddValidation(state.Version, validator)
	assert.NoError(t, err)

	consensus, err := syncManager.GetConsensus(state.Version)
	assert.NoError(t, err)
	assert.Contains(t, consensus.Validators, validator)
}

// Test helpers
type testStore struct {
	states map[synctypes.StateVersion]*synctypes.VersionedState
}

func newTestStore() *testStore {
	return &testStore{
		states: make(map[synctypes.StateVersion]*synctypes.VersionedState),
	}
}

func (s *testStore) GetState(version synctypes.StateVersion) (*synctypes.VersionedState, error) {
	state, exists := s.states[version]
	if !exists {
		return nil, fmt.Errorf("state not found")
	}
	return state, nil
}

func (s *testStore) SaveState(state *synctypes.VersionedState) error {
	s.states[state.Version] = state
	return nil
}

func (s *testStore) ListVersions() ([]synctypes.StateVersion, error) {
	versions := make([]synctypes.StateVersion, 0)
	for v := range s.states {
		versions = append(versions, v)
	}
	return versions, nil
}

func (s *testStore) TrackChange(change *synctypes.StateChange) error {
	return nil
}

func (s *testStore) GetChanges(since time.Time) ([]*synctypes.StateChange, error) {
	return nil, nil
}

type testResolver struct{}

func newTestResolver() *testResolver {
	return &testResolver{}
}

func (r *testResolver) DetectConflicts(states []*synctypes.VersionedState) ([]*synctypes.StateChange, error) {
	return nil, nil
}

func (r *testResolver) ResolveConflicts(changes []*synctypes.StateChange) (*synctypes.VersionedState, error) {
	return nil, nil
}

func (r *testResolver) ValidateResolution(state *synctypes.VersionedState) error {
	if state.Version == "invalid" {
		return fmt.Errorf("invalid state")
	}
	return nil
}

type testConfig struct {
	configs map[string]*synctypes.ConfigData
}

func newTestConfig() *testConfig {
	return &testConfig{
		configs: make(map[string]*synctypes.ConfigData),
	}
}

func (c *testConfig) GetConfig(version string) (*synctypes.ConfigData, error) {
	config, exists := c.configs[version]
	if !exists {
		return nil, fmt.Errorf("config not found")
	}
	return config, nil
}

func (c *testConfig) UpdateConfig(config *synctypes.ConfigData) error {
	c.configs[config.Version] = config
	return nil
}

func (c *testConfig) ListConfigs() ([]*synctypes.ConfigData, error) {
	configs := make([]*synctypes.ConfigData, 0)
	for _, config := range c.configs {
		configs = append(configs, config)
	}
	return configs, nil
}

func (c *testConfig) DistributeConfig(config *synctypes.ConfigData, devices []types.DeviceID) error {
	return nil
}

func (c *testConfig) GetDeviceConfig(deviceID types.DeviceID) (*synctypes.ConfigData, error) {
	// Return latest config for testing
	for _, config := range c.configs {
		return config, nil
	}
	return nil, fmt.Errorf("no config found")
}
