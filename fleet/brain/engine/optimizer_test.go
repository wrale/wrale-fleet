package engine

import (
    "context"
    "testing"

    "github.com/wrale/wrale-fleet/fleet/brain/device"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestOptimizer(t *testing.T) {
    ctx := context.Background()
    inventory := device.NewInventory()
    topology := device.NewTopologyManager(inventory)
    analyzer := NewAnalyzer(inventory, topology)
    optimizer := NewOptimizer(inventory, topology, analyzer)

    // Register test rack
    err := topology.RegisterRack(ctx, "rack-1", device.RackConfig{
        MaxUnits:    42,
        PowerLimit:  5000.0,
        CoolingZone: "zone-1",
    })
    if err != nil {
        t.Fatalf("Failed to register test rack: %v", err)
    }

    // Initialize test devices with varying resource utilization
    devices := []types.DeviceState{
        {
            ID: "low-usage-device",
            Status: "active",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    100.0,
                types.ResourceMemory: 100.0,
            },
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 1,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 45.0,
                PowerUsage:  300.0,
                CPULoad:     30.0,
                MemoryUsage: 40.0,
            },
        },
        {
            ID: "high-usage-device",
            Status: "active",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    100.0,
                types.ResourceMemory: 100.0,
            },
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 2,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 65.0,
                PowerUsage:  700.0,
                CPULoad:     90.0,
                MemoryUsage: 85.0,
            },
        },
    }

    // Register test devices
    for _, d := range devices {
        err := inventory.RegisterDevice(ctx, d)
        if err != nil {
            t.Fatalf("Failed to register device %s: %v", d.ID, err)
        }
    }

    t.Run("Optimize Resources", func(t *testing.T) {
        deviceStates, err := inventory.ListDevices(ctx)
        if err != nil {
            t.Fatalf("Failed to list devices: %v", err)
        }

        optimized, err := optimizer.OptimizeResources(ctx, deviceStates)
        if err != nil {
            t.Fatalf("Failed to optimize resources: %v", err)
        }

        // Verify high usage device gets reduced allocation
        for _, device := range optimized {
            if device.ID == "high-usage-device" {
                if device.Resources[types.ResourceCPU] >= 100.0 {
                    t.Error("Expected reduced CPU allocation for high usage device")
                }
            }
        }
    })

    t.Run("Resource Utilization", func(t *testing.T) {
        utilization, err := optimizer.GetResourceUtilization(ctx)
        if err != nil {
            t.Fatalf("Failed to get resource utilization: %v", err)
        }

        // Verify utilization calculations
        expectedCPU := (30.0 + 90.0) / 2.0 // Average of two devices
        if utilization[types.ResourceCPU] != expectedCPU {
            t.Errorf("Expected CPU utilization %.2f, got %.2f", expectedCPU, utilization[types.ResourceCPU])
        }
    })

    t.Run("Task Placement", func(t *testing.T) {
        task := types.Task{
            ID: "test-task",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    20.0,
                types.ResourceMemory: 30.0,
            },
            DeviceIDs: make([]types.DeviceID, 1),
        }

        placements, err := optimizer.SuggestPlacements(ctx, task)
        if err != nil {
            t.Fatalf("Failed to suggest placements: %v", err)
        }

        if len(placements) == 0 {
            t.Fatal("No placements suggested")
        }

        // Should prefer low usage device
        if placements[0] != "low-usage-device" {
            t.Error("Expected low usage device as first placement suggestion")
        }
    })

    t.Run("Power Optimization", func(t *testing.T) {
        // Add a high power device
        highPowerDevice := types.DeviceState{
            ID: "high-power-device",
            Status: "active",
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 3,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                PowerUsage: 2000.0,
                CPULoad:    70.0,
            },
        }

        err := inventory.RegisterDevice(ctx, highPowerDevice)
        if err != nil {
            t.Fatalf("Failed to register high power device: %v", err)
        }

        recommendations, err := optimizer.OptimizePowerDistribution(ctx)
        if err != nil {
            t.Fatalf("Failed to optimize power: %v", err)
        }

        // Should recommend power rebalancing due to high power device
        foundPowerRec := false
        for _, rec := range recommendations {
            if rec.Action == "rebalance_power" {
                foundPowerRec = true
                break
            }
        }

        if !foundPowerRec {
            t.Error("Expected power rebalance recommendation")
        }
    })

    t.Run("Thermal Optimization", func(t *testing.T) {
        // Add a hot device
        hotDevice := types.DeviceState{
            ID: "hot-device",
            Status: "active",
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 4,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 85.0,
                CPULoad:    60.0,
            },
        }

        err := inventory.RegisterDevice(ctx, hotDevice)
        if err != nil {
            t.Fatalf("Failed to register hot device: %v", err)
        }

        recommendations, err := optimizer.OptimizeZone(ctx, "zone-1")
        if err != nil {
            t.Fatalf("Failed to optimize zone: %v", err)
        }

        // Should recommend workload redistribution for thermal balance
        foundThermalRec := false
        for _, rec := range recommendations {
            if rec.Action == "redistribute_workload" {
                foundThermalRec = true
                break
            }
        }

        if !foundThermalRec {
            t.Error("Expected thermal redistribution recommendation")
        }
    })
}