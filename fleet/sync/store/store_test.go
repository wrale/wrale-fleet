package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestFileStore(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "wrale-sync-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	t.Run("State Operations", func(t *testing.T) {
		deviceState := types.DeviceState{
			ID:     "test-device-1",
			Status: "active",
			Metrics: types.DeviceMetrics{
				Temperature: 45.0,
				PowerUsage:  400.0,
			},
		}

		testState := &synctypes.VersionedState{
			Version:   synctypes.StateVersion("1"),
			State:     deviceState,
			Timestamp: time.Now(),
		}

		// Save state
		if err := store.SaveState(testState); err != nil {
			t.Errorf("Failed to save state: %v", err)
		}

		// Get state back
		state, err := store.GetState(testState.Version)
		if err != nil {
			t.Errorf("Failed to get state: %v", err)
		}
		if state.Version != testState.Version {
			t.Errorf("Version mismatch - Got: %v, Want: %v",
				state.Version, testState.Version)
		}

		// List versions
		versions, err := store.ListVersions()
		if err != nil {
			t.Errorf("Failed to list versions: %v", err)
		}
		if len(versions) != 1 {
			t.Errorf("Expected 1 version, got %d", len(versions))
		}
		if versions[0] != testState.Version {
			t.Errorf("Version mismatch in list - Got: %v, Want: %v",
				versions[0], testState.Version)
		}
	})

	t.Run("Change Tracking", func(t *testing.T) {
		// Create test change
		oldState := types.DeviceState{
			ID:     "test-device-1",
			Status: "active",
		}
		newState := types.DeviceState{
			ID:     "test-device-1",
			Status: "updated",
		}

		change := &synctypes.StateChange{
			DeviceID:    "test-device-1",
			PrevVersion: synctypes.StateVersion("1"),
			NewVersion:  synctypes.StateVersion("2"),
			OldState:    &oldState,
			NewState:    newState,
			Timestamp:   time.Now(),
			Source:      "test",
			Changes:     []string{"status"},
		}

		// Track change
		if err := store.TrackChange(change); err != nil {
			t.Errorf("Failed to track change: %v", err)
		}

		// Get changes
		pastTime := time.Now().Add(-time.Hour)
		changes, err := store.GetChanges(pastTime)
		if err != nil {
			t.Errorf("Failed to get changes: %v", err)
		}
		if len(changes) != 1 {
			t.Errorf("Expected 1 change, got %d", len(changes))
		}
		if changes[0].NewVersion != change.NewVersion {
			t.Errorf("Version mismatch in change - Got: %v, Want: %v",
				changes[0].NewVersion, change.NewVersion)
		}
	})

	t.Run("Store Persistence", func(t *testing.T) {
		// Create new store instance with same directory
		newStore, err := NewFileStore(tempDir)
		if err != nil {
			t.Fatalf("Failed to create new store: %v", err)
		}

		// Verify state persists
		state, err := newStore.GetState(synctypes.StateVersion("1"))
		if err != nil {
			t.Errorf("Failed to get state from new store: %v", err)
		}
		if state == nil {
			t.Error("State not found in new store instance")
		}

		// Verify changes persist
		changes, err := newStore.GetChanges(time.Now().Add(-time.Hour))
		if err != nil {
			t.Errorf("Failed to get changes from new store: %v", err)
		}
		if len(changes) != 1 {
			t.Errorf("Expected 1 change in new store, got %d", len(changes))
		}
	})

	t.Run("Multiple States", func(t *testing.T) {
		deviceStates := []types.DeviceState{
			{
				ID:     "device-1",
				Status: "active",
			},
			{
				ID:     "device-2",
				Status: "active",
			},
		}

		states := []*synctypes.VersionedState{
			{
				Version:   synctypes.StateVersion("3"),
				State:     deviceStates[0],
				Timestamp: time.Now(),
			},
			{
				Version:   synctypes.StateVersion("4"),
				State:     deviceStates[1],
				Timestamp: time.Now(),
			},
		}

		// Save states
		for _, state := range states {
			if err := store.SaveState(state); err != nil {
				t.Errorf("Failed to save state %v: %v", state.Version, err)
			}
		}

		// Verify all states are listed
		versions, err := store.ListVersions()
		if err != nil {
			t.Errorf("Failed to list versions: %v", err)
		}
		if len(versions) != 3 { // Including previous state
			t.Errorf("Expected 3 versions, got %d", len(versions))
		}
	})

	t.Run("File Operations", func(t *testing.T) {
		// Test JSON read/write
		type testData struct {
			Field string `json:"field"`
		}
		data := testData{Field: "test"}
		path := filepath.Join(tempDir, "test.json")

		// Write
		if err := writeJSON(path, data); err != nil {
			t.Errorf("Failed to write JSON: %v", err)
		}

		// Read
		var readData testData
		if err := readJSON(path, &readData); err != nil {
			t.Errorf("Failed to read JSON: %v", err)
		}
		if readData.Field != data.Field {
			t.Errorf("Data mismatch - Got: %s, Want: %s",
				readData.Field, data.Field)
		}
	})
}
