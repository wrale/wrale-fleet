package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestStateManager(t *testing.T) {
	ctx := context.Background()
	sm := NewStateManager()

	testDevice := types.DeviceState{
		ID: "test-device-1",
		Status: "active",
		Resources: map[types.ResourceType]float64{
			types.ResourceCPU:    50.0,
			types.ResourceMemory: 60.0,
			types.ResourcePower:  400.0,
		},
		Location: types.PhysicalLocation{
			Rack:     "rack-1",
			Position: 1,
			Zone:     "zone-1",
		},
		Metrics: types.DeviceMetrics{
			Temperature: 45.5,
			PowerUsage:  385.0,
			CPULoad:     48.5,
			MemoryUsage: 58.2,
		},
	}

	t.Run("Update and Get Device State", func(t *testing.T) {
		// Update state
		err := sm.UpdateDeviceState(ctx, testDevice)
		if err != nil {
			t.Errorf("Failed to update device state: %v", err)
		}

		// Get state back
		state, err := sm.GetDeviceState(ctx, testDevice.ID)
		if err != nil {
			t.Errorf("Failed to get device state: %v", err)
		}
		if state == nil {
			t.Fatal("Device state not found after update")
		}

		// Verify state
		if state.ID != testDevice.ID {
			t.Errorf("Expected device ID %s, got %s", testDevice.ID, state.ID)
		}
		if state.Status != testDevice.Status {
			t.Errorf("Expected status %s, got %s", testDevice.Status, state.Status)
		}
		if state.Metrics.Temperature != testDevice.Metrics.Temperature {
			t.Errorf("Expected temperature %.2f, got %.2f", 
				testDevice.Metrics.Temperature, state.Metrics.Temperature)
		}

		// Verify LastUpdated was set
		if state.LastUpdated.IsZero() {
			t.Error("LastUpdated timestamp not set")
		}
	})

	t.Run("List Devices", func(t *testing.T) {
		// Add another device
		testDevice2 := testDevice
		testDevice2.ID = "test-device-2"
		err := sm.UpdateDeviceState(ctx, testDevice2)
		if err != nil {
			t.Errorf("Failed to update second device state: %v", err)
		}

		// List all devices
		devices, err := sm.ListDevices(ctx)
		if err != nil {
			t.Errorf("Failed to list devices: %v", err)
		}

		if len(devices) != 2 {
			t.Errorf("Expected 2 devices, got %d", len(devices))
		}

		// Verify both devices are present
		found1, found2 := false, false
		for _, device := range devices {
			if device.ID == testDevice.ID {
				found1 = true
			}
			if device.ID == testDevice2.ID {
				found2 = true
			}
		}

		if !found1 || !found2 {
			t.Error("Not all devices returned in list")
		}
	})

	t.Run("Remove Device", func(t *testing.T) {
		// Remove a device
		err := sm.RemoveDevice(ctx, testDevice.ID)
		if err != nil {
			t.Errorf("Failed to remove device: %v", err)
		}

		// Verify it's gone
		state, err := sm.GetDeviceState(ctx, testDevice.ID)
		if err != nil {
			t.Errorf("Error getting removed device: %v", err)
		}
		if state != nil {
			t.Error("Device still exists after removal")
		}

		// Verify only one device remains
		devices, err := sm.ListDevices(ctx)
		if err != nil {
			t.Errorf("Failed to list devices: %v", err)
		}
		if len(devices) != 1 {
			t.Errorf("Expected 1 device after removal, got %d", len(devices))
		}
	})

	t.Run("Update Existing Device", func(t *testing.T) {
		// Initial state
		err := sm.UpdateDeviceState(ctx, testDevice)
		if err != nil {
			t.Errorf("Failed to set initial state: %v", err)
		}

		// Update metrics
		updatedDevice := testDevice
		updatedDevice.Metrics.Temperature = 50.0
		updatedDevice.Metrics.CPULoad = 75.0
		time.Sleep(time.Millisecond) // Ensure time difference

		err = sm.UpdateDeviceState(ctx, updatedDevice)
		if err != nil {
			t.Errorf("Failed to update device: %v", err)
		}

		// Verify update
		state, err := sm.GetDeviceState(ctx, testDevice.ID)
		if err != nil {
			t.Errorf("Failed to get updated device: %v", err)
		}
		if state.Metrics.Temperature != updatedDevice.Metrics.Temperature {
			t.Errorf("Expected temperature %.2f, got %.2f",
				updatedDevice.Metrics.Temperature, state.Metrics.Temperature)
		}
		if state.Metrics.CPULoad != updatedDevice.Metrics.CPULoad {
			t.Errorf("Expected CPU load %.2f, got %.2f",
				updatedDevice.Metrics.CPULoad, state.Metrics.CPULoad)
		}

		// Verify LastUpdated was updated
		if state.LastUpdated.Equal(testDevice.LastUpdated) {
			t.Error("LastUpdated timestamp not updated")
		}
	})
}