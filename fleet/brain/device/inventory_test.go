package device

import (
    "context"
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestInventory(t *testing.T) {
    ctx := context.Background()
    inventory := NewInventory()

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

    t.Run("Register Device", func(t *testing.T) {
        err := inventory.RegisterDevice(ctx, testDevice)
        if err != nil {
            t.Errorf("Failed to register device: %v", err)
        }

        // Verify device exists
        state, err := inventory.GetDevice(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to get device: %v", err)
        }
        if state == nil {
            t.Fatal("Device not found after registration")
        }

        // Verify initial health status
        info := inventory.devices[testDevice.ID]
        if info.Health != HealthStatusHealthy {
            t.Errorf("Expected initial health status %s, got %s",
                HealthStatusHealthy, info.Health)
        }
    })

    t.Run("Update Device State", func(t *testing.T) {
        // Update device metrics to trigger health status change
        updatedDevice := testDevice
        updatedDevice.Metrics.Temperature = 85.0 // Above threshold
        updatedDevice.Metrics.CPULoad = 95.0     // Above threshold

        err := inventory.UpdateState(ctx, updatedDevice)
        if err != nil {
            t.Errorf("Failed to update device state: %v", err)
        }

        // Verify state update
        state, err := inventory.GetDevice(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to get updated device state: %v", err)
        }
        if state.Metrics.Temperature != updatedDevice.Metrics.Temperature {
            t.Errorf("Expected temperature %.2f, got %.2f",
                updatedDevice.Metrics.Temperature, state.Metrics.Temperature)
        }

        // Verify health status changed
        info := inventory.devices[testDevice.ID]
        if info.Health != HealthStatusUnhealthy {
            t.Errorf("Expected health status %s, got %s",
                HealthStatusUnhealthy, info.Health)
        }
    })

    t.Run("List Devices", func(t *testing.T) {
        // Add another device
        testDevice2 := testDevice
        testDevice2.ID = "test-device-2"
        err := inventory.RegisterDevice(ctx, testDevice2)
        if err != nil {
            t.Errorf("Failed to register second device: %v", err)
        }

        devices, err := inventory.ListDevices(ctx)
        if err != nil {
            t.Errorf("Failed to list devices: %v", err)
        }

        if len(devices) != 2 {
            t.Errorf("Expected 2 devices, got %d", len(devices))
        }
    })

    t.Run("Unregister Device", func(t *testing.T) {
        err := inventory.UnregisterDevice(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to unregister device: %v", err)
        }

        // Verify device is removed
        state, err := inventory.GetDevice(ctx, testDevice.ID)
        if err == nil {
            t.Error("Expected error getting unregistered device")
        }
        if state != nil {
            t.Error("Device still exists after unregistration")
        }
    })

    t.Run("Health Report", func(t *testing.T) {
        inventory := NewInventory() // Fresh inventory for clean health counts

        // Register devices with different health states
        devices := []types.DeviceState{
            {
                ID: "healthy-device",
                Metrics: types.DeviceMetrics{
                    Temperature: 45.0,
                    CPULoad:     50.0,
                },
            },
            {
                ID: "degraded-device",
                Metrics: types.DeviceMetrics{
                    Temperature: 75.0,
                    CPULoad:     92.0,
                },
            },
            {
                ID: "unhealthy-device",
                Metrics: types.DeviceMetrics{
                    Temperature: 85.0,
                    CPULoad:     95.0,
                },
            },
        }

        for _, device := range devices {
            err := inventory.RegisterDevice(ctx, device)
            if err != nil {
                t.Errorf("Failed to register device %s: %v", device.ID, err)
                continue
            }
            err = inventory.UpdateState(ctx, device)
            if err != nil {
                t.Errorf("Failed to update device %s state: %v", device.ID, err)
            }
        }

        report := inventory.GetHealthReport(ctx)

        expectedCounts := map[HealthStatus]int{
            HealthStatusHealthy:   1,
            HealthStatusDegraded:  1,
            HealthStatusUnhealthy: 1,
        }

        for status, expected := range expectedCounts {
            if report[status] != expected {
                t.Errorf("Expected %d devices with status %s, got %d",
                    expected, status, report[status])
            }
        }
    })

    t.Run("Contact Tracking", func(t *testing.T) {
        inventory := NewInventory()
        device := testDevice
        device.ID = "contact-test-device"

        // Register device
        err := inventory.RegisterDevice(ctx, device)
        if err != nil {
            t.Errorf("Failed to register device: %v", err)
        }

        initialContact := inventory.devices[device.ID].LastContact

        // Wait briefly
        time.Sleep(time.Millisecond * 10)

        // Update state
        err = inventory.UpdateState(ctx, device)
        if err != nil {
            t.Errorf("Failed to update device state: %v", err)
        }

        // Verify LastContact was updated
        updatedContact := inventory.devices[device.ID].LastContact
        if !updatedContact.After(initialContact) {
            t.Error("LastContact timestamp not updated")
        }
    })
}