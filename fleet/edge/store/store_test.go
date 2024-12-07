package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	"github.com/wrale/wrale-fleet/fleet/edge/agent"
)

func TestFileStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wrale-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	t.Run("State Operations", func(t *testing.T) {
		// Initial empty state
		state, err := store.GetState()
		if err != nil {
			t.Errorf("Failed to get initial state: %v", err)
		}
		if state.IsHealthy {
			t.Error("Expected empty initial state to be unhealthy")
		}

		// Update state
		testState := agent.AgentState{
			DeviceState: types.DeviceState{
				ID:     "test-device",
				Status: "active",
				Metrics: types.DeviceMetrics{
					Temperature: 45.0,
					PowerUsage:  400.0,
				},
			},
			LastSync:  time.Now(),
			IsHealthy: true,
			Mode:      agent.ModeNormal,
		}

		if err := store.UpdateState(testState); err != nil {
			t.Errorf("Failed to update state: %v", err)
		}

		// Read back state
		savedState, err := store.GetState()
		if err != nil {
			t.Errorf("Failed to read back state: %v", err)
		}

		if savedState.DeviceState.ID != testState.DeviceState.ID {
			t.Errorf("State mismatch - Expected ID: %s, Got: %s",
				testState.DeviceState.ID, savedState.DeviceState.ID)
		}
	})

	t.Run("Command History", func(t *testing.T) {
		// Test command results
		results := []agent.CommandResult{
			{
				CommandID:   "cmd-1",
				Success:     true,
				CompletedAt: time.Now().Add(-time.Hour),
			},
			{
				CommandID:   "cmd-2",
				Success:     false,
				Error:       fmt.Errorf("test error"),
				CompletedAt: time.Now(),
			},
		}

		// Add results
		for _, result := range results {
			if err := store.AddCommandResult(result); err != nil {
				t.Errorf("Failed to add command result: %v", err)
			}
		}

		// Read back history
		history, err := store.GetCommandHistory()
		if err != nil {
			t.Errorf("Failed to get command history: %v", err)
		}

		if len(history) != len(results) {
			t.Errorf("Expected %d results, got %d", len(results), len(history))
		}
	})

	t.Run("Configuration", func(t *testing.T) {
		// Initial empty config
		config, err := store.GetConfig()
		if err != nil {
			t.Errorf("Failed to get initial config: %v", err)
		}
		if len(config) != 0 {
			t.Error("Expected empty initial config")
		}

		// Update config
		testConfig := map[string]interface{}{
			"update_interval": 30,
			"brain_endpoint":  "http://brain:8080",
			"metal_endpoint":  "http://localhost:9090",
		}

		if err := store.UpdateConfig(testConfig); err != nil {
			t.Errorf("Failed to update config: %v", err)
		}

		// Read back config
		savedConfig, err := store.GetConfig()
		if err != nil {
			t.Errorf("Failed to read back config: %v", err)
		}

		if savedConfig["update_interval"] != testConfig["update_interval"] {
			t.Error("Config mismatch in update_interval")
		}
	})

	t.Run("Command History Cleanup", func(t *testing.T) {
		// Add commands with different ages
		now := time.Now()
		results := []agent.CommandResult{
			{
				CommandID:   "old-1",
				CompletedAt: now.Add(-48 * time.Hour),
			},
			{
				CommandID:   "old-2",
				CompletedAt: now.Add(-25 * time.Hour),
			},
			{
				CommandID:   "new-1",
				CompletedAt: now.Add(-1 * time.Hour),
			},
		}

		for _, result := range results {
			if err := store.AddCommandResult(result); err != nil {
				t.Errorf("Failed to add command result: %v", err)
			}
		}

		// Clean up commands older than 24 hours
		if err := store.Cleanup(24 * time.Hour); err != nil {
			t.Errorf("Failed to cleanup command history: %v", err)
		}

		// Verify cleanup
		history, err := store.GetCommandHistory()
		if err != nil {
			t.Errorf("Failed to get command history after cleanup: %v", err)
		}

		if len(history) != 1 {
			t.Errorf("Expected 1 command after cleanup, got %d", len(history))
		}

		if len(history) > 0 && history[0].CommandID != "new-1" {
			t.Error("Expected only new command to remain after cleanup")
		}
	})

	t.Run("Store Persistence", func(t *testing.T) {
		// Create new store instance with same directory
		newStore, err := NewFileStore(tempDir)
		if err != nil {
			t.Fatalf("Failed to create new store: %v", err)
		}

		// Verify state persists
		state, err := newStore.GetState()
		if err != nil {
			t.Errorf("Failed to get state from new store: %v", err)
		}
		if state.DeviceState.ID != "test-device" {
			t.Error("Failed to persist state across store instances")
		}

		// Verify config persists
		config, err := newStore.GetConfig()
		if err != nil {
			t.Errorf("Failed to get config from new store: %v", err)
		}
		if config["brain_endpoint"] != "http://brain:8080" {
			t.Error("Failed to persist config across store instances")
		}
	})
}
