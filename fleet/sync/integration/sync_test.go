package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	"github.com/wrale/wrale-fleet/fleet/sync/config"
	"github.com/wrale/wrale-fleet/fleet/sync/manager"
	"github.com/wrale/wrale-fleet/fleet/sync/resolver"
	"github.com/wrale/wrale-fleet/fleet/sync/store"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestSyncIntegration(t *testing.T) {
	// Create temp directory for testing
	tempDir, err := os.MkdirTemp("", "wrale-sync-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize components
	stateStore, err := store.NewFileStore(filepath.Join(tempDir, "states"))
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	conflictResolver := resolver.NewResolver(100)
	configManager := config.NewManager()
	syncManager := manager.NewManager(stateStore, conflictResolver, configManager)

	t.Run("End-to-End State Sync", func(t *testing.T) {
		// Create initial device state
		deviceID := types.DeviceID("test-device-1")
		state1 := &synctypes.VersionedState{
			Version: "v1",
			State: types.DeviceState{
				ID:     deviceID,
				Status: "active",
				Metrics: types.DeviceMetrics{
					Temperature: 45.0,
					PowerUsage:  400.0,
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "test",
		}

		// Update state through sync manager
		err := syncManager.UpdateState(deviceID, state1)
		if err != nil {
			t.Errorf("Failed to update initial state: %v", err)
		}

		// Create conflicting update
		state2 := &synctypes.VersionedState{
			Version: "v2",
			State: types.DeviceState{
				ID:     deviceID,
				Status: "standby",
				Metrics: types.DeviceMetrics{
					Temperature: 40.0,
					PowerUsage:  200.0,
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "test",
		}

		// Update with conflict
		err = syncManager.UpdateState(deviceID, state2)
		if err != nil {
			t.Errorf("Failed to update conflicting state: %v", err)
		}

		// Get final state
		finalState, err := syncManager.GetState(deviceID)
		if err != nil {
			t.Errorf("Failed to get final state: %v", err)
		}

		// Verify conflict resolution
		if finalState.Version == state1.Version {
			t.Error("State was not updated after conflict")
		}
		if finalState.State.Status != "standby" {
			t.Errorf("Expected status 'standby', got '%s'", finalState.State.Status)
		}
	})

	t.Run("Config Distribution", func(t *testing.T) {
		// Create test config
		config := &synctypes.ConfigData{
			Config: map[string]interface{}{
				"update_interval": 30,
				"max_retries":     5,
			},
			ValidFrom: time.Now(),
		}

		// Update config through sync manager
		err := syncManager.UpdateConfig(config)
		if err != nil {
			t.Errorf("Failed to update config: %v", err)
		}

		// Distribute to devices
		devices := []types.DeviceID{
			"device-1",
			"device-2",
		}

		err = syncManager.DistributeConfig(config.Version, devices)
		if err != nil {
			t.Errorf("Failed to distribute config: %v", err)
		}

		// Verify distribution
		for _, deviceID := range devices {
			deviceConfig, err := syncManager.GetDeviceConfig(deviceID)
			if err != nil {
				t.Errorf("Failed to get device config: %v", err)
			}
			if deviceConfig.Version != config.Version {
				t.Errorf("Expected config version %s, got %s",
					config.Version, deviceConfig.Version)
			}
		}
	})

	t.Run("State Validation", func(t *testing.T) {
		deviceID := types.DeviceID("test-device-2")

		// Create valid state
		validState := &synctypes.VersionedState{
			Version: "v1",
			State: types.DeviceState{
				ID:     deviceID,
				Status: "active",
				Metrics: types.DeviceMetrics{
					Temperature: 45.0,
					CPULoad:     50.0,
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "test",
		}

		// Should succeed
		err := syncManager.ValidateState(validState)
		if err != nil {
			t.Errorf("Failed to validate valid state: %v", err)
		}

		// Create invalid state
		invalidState := &synctypes.VersionedState{
			Version: "v2",
			State: types.DeviceState{
				ID:     deviceID,
				Status: "active",
				Metrics: types.DeviceMetrics{
					Temperature: 101.0, // Invalid temperature
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "test",
		}

		// Should fail
		err = syncManager.ValidateState(invalidState)
		if err == nil {
			t.Error("Expected validation error for invalid state")
		}
	})

	t.Run("Operation Tracking", func(t *testing.T) {
		// Create test operation
		op := &synctypes.SyncOperation{
			ID:       "test-op-1",
			Type:     synctypes.OpStateSync,
			Priority: 1,
		}

		// Create operation
		err := syncManager.CreateOperation(op)
		if err != nil {
			t.Errorf("Failed to create operation: %v", err)
		}

		// Get operation
		retrieved, err := syncManager.GetOperation(op.ID)
		if err != nil {
			t.Errorf("Failed to get operation: %v", err)
		}
		if retrieved.ID != op.ID {
			t.Error("Operation ID mismatch")
		}

		// List operations
		ops, err := syncManager.ListOperations()
		if err != nil {
			t.Errorf("Failed to list operations: %v", err)
		}
		if len(ops) == 0 {
			t.Error("No operations found")
		}
	})

	t.Run("Error Recovery", func(t *testing.T) {
		// Simulate operation failure
		deviceID := types.DeviceID("test-device-3")
		state := &synctypes.VersionedState{
			Version: "v1",
			State: types.DeviceState{
				ID:     deviceID,
				Status: "error",
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "test",
		}

		// Should handle error gracefully
		err := syncManager.UpdateState(deviceID, state)
		if err != nil {
			t.Errorf("Failed to handle error state: %v", err)
		}

		// State should be recoverable
		recovered, err := syncManager.GetState(deviceID)
		if err != nil {
			t.Errorf("Failed to recover state: %v", err)
		}
		if recovered.Version != state.Version {
			t.Error("Failed to recover correct state version")
		}
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		deviceID := types.DeviceID("test-device-4")
		done := make(chan bool)
		errors := make(chan error, 2)

		// Concurrent state updates
		for i := 0; i < 2; i++ {
			go func(i int) {
				state := &synctypes.VersionedState{
					Version: fmt.Sprintf("v%d", i+1),
					State: types.DeviceState{
						ID:     deviceID,
						Status: fmt.Sprintf("status-%d", i),
					},
					UpdatedAt: time.Now(),
					UpdatedBy: fmt.Sprintf("test-%d", i),
				}

				err := syncManager.UpdateState(deviceID, state)
				if err != nil {
					errors <- err
					return
				}
				done <- true
			}(i)
		}

		// Wait for completion
		for i := 0; i < 2; i++ {
			select {
			case err := <-errors:
				t.Errorf("Concurrent operation failed: %v", err)
			case <-done:
				continue
			case <-time.After(time.Second * 5):
				t.Error("Concurrent operation timed out")
			}
		}

		// Verify final state is consistent
		finalState, err := syncManager.GetState(deviceID)
		if err != nil {
			t.Errorf("Failed to get final state: %v", err)
		}
		if finalState == nil {
			t.Error("No final state found")
		}
	})
}
