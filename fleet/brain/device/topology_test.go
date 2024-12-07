package device

import (
    "context"
    "testing"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestTopologyManager(t *testing.T) {
    ctx := context.Background()
    inventory := NewInventory()
    topology := NewTopologyManager(inventory)

    // Test rack configuration
    testRack := RackConfig{
        MaxUnits:    42,
        PowerLimit:  5000.0,
        CoolingZone: "zone-1",
    }

    testDevice := types.DeviceState{
        ID: "test-device-1",
        Location: types.PhysicalLocation{
            Rack:     "rack-1",
            Position: 1,
            Zone:     "zone-1",
        },
        Metrics: types.DeviceMetrics{
            Temperature: 45.0,
            PowerUsage:  400.0,
        },
    }

    t.Run("Register Rack", func(t *testing.T) {
        err := topology.RegisterRack(ctx, "rack-1", testRack)
        if err != nil {
            t.Errorf("Failed to register rack: %v", err)
        }

        // Verify rack configuration
        if config, exists := topology.rackConfigs["rack-1"]; !exists {
            t.Error("Rack not found after registration")
        } else {
            if config.MaxUnits != testRack.MaxUnits {
                t.Errorf("Expected MaxUnits %d, got %d", testRack.MaxUnits, config.MaxUnits)
            }
            if config.PowerLimit != testRack.PowerLimit {
                t.Errorf("Expected PowerLimit %.2f, got %.2f", testRack.PowerLimit, config.PowerLimit)
            }
            if config.CoolingZone != testRack.CoolingZone {
                t.Errorf("Expected CoolingZone %s, got %s", testRack.CoolingZone, config.CoolingZone)
            }
        }
    })

    t.Run("Device Location Management", func(t *testing.T) {
        // Register device
        err := inventory.RegisterDevice(ctx, testDevice)
        if err != nil {
            t.Errorf("Failed to register device: %v", err)
        }

        // Get location
        location, err := topology.GetLocation(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to get device location: %v", err)
        }
        if location == nil {
            t.Fatal("Location not found")
        }

        // Verify location
        if location.Rack != testDevice.Location.Rack {
            t.Errorf("Expected rack %s, got %s", testDevice.Location.Rack, location.Rack)
        }
        if location.Position != testDevice.Location.Position {
            t.Errorf("Expected position %d, got %d", testDevice.Location.Position, location.Position)
        }
    })

    t.Run("Validate Location", func(t *testing.T) {
        validLocation := types.PhysicalLocation{
            Rack:     "rack-1",
            Position: 1,
            Zone:     "zone-1",
        }

        err := topology.ValidateLocation(ctx, validLocation)
        if err != nil {
            t.Errorf("Failed to validate valid location: %v", err)
        }

        invalidLocation := types.PhysicalLocation{
            Rack:     "rack-1",
            Position: 43, // Beyond max units
            Zone:     "zone-1",
        }

        err = topology.ValidateLocation(ctx, invalidLocation)
        if err == nil {
            t.Error("Expected error for invalid position")
        }

        nonexistentRack := types.PhysicalLocation{
            Rack:     "nonexistent-rack",
            Position: 1,
            Zone:     "zone-1",
        }

        err = topology.ValidateLocation(ctx, nonexistentRack)
        if err == nil {
            t.Error("Expected error for nonexistent rack")
        }
    })

    t.Run("Zone Management", func(t *testing.T) {
        // Add another device in same zone
        testDevice2 := testDevice
        testDevice2.ID = "test-device-2"
        testDevice2.Location.Position = 2

        err := inventory.RegisterDevice(ctx, testDevice2)
        if err != nil {
            t.Errorf("Failed to register second device: %v", err)
        }

        // Get devices in zone
        devices, err := topology.GetDevicesInZone(ctx, "zone-1")
        if err != nil {
            t.Errorf("Failed to get devices in zone: %v", err)
        }

        if len(devices) != 2 {
            t.Errorf("Expected 2 devices in zone, got %d", len(devices))
        }
    })

    t.Run("Rack Power Usage", func(t *testing.T) {
        power, err := topology.GetRackPowerUsage(ctx, "rack-1")
        if err != nil {
            t.Errorf("Failed to get rack power usage: %v", err)
        }

        expectedPower := 800.0 // 400W per device * 2 devices
        if power != expectedPower {
            t.Errorf("Expected power usage %.2f, got %.2f", expectedPower, power)
        }
    })

    t.Run("Update Location", func(t *testing.T) {
        newLocation := types.PhysicalLocation{
            Rack:     "rack-1",
            Position: 3,
            Zone:     "zone-1",
        }

        err := topology.UpdateLocation(ctx, testDevice.ID, newLocation)
        if err != nil {
            t.Errorf("Failed to update device location: %v", err)
        }

        location, err := topology.GetLocation(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to get updated location: %v", err)
        }
        if location.Position != newLocation.Position {
            t.Errorf("Expected position %d, got %d", newLocation.Position, location.Position)
        }
    })

    t.Run("Unregister Rack", func(t *testing.T) {
        err := topology.UnregisterRack(ctx, "rack-1")
        if err != nil {
            t.Errorf("Failed to unregister rack: %v", err)
        }

        if _, exists := topology.rackConfigs["rack-1"]; exists {
            t.Error("Rack still exists after unregistration")
        }

        // Verify location validation fails after rack removal
        err = topology.ValidateLocation(ctx, testDevice.Location)
        if err == nil {
            t.Error("Expected error validating location after rack removal")
        }
    })
}