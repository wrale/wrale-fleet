package engine

import (
    "context"
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/device"
    "github.com/wrale/wrale-fleet/fleet/types"
)

func TestAnalyzer(t *testing.T) {
    ctx := context.Background()
    inventory := device.NewInventory()
    topology := device.NewTopologyManager(inventory)
    analyzer := NewAnalyzer(inventory, topology)

    // Register test rack
    err := topology.RegisterRack(ctx, "rack-1", device.RackConfig{
        MaxUnits:    42,
        PowerLimit:  5000.0,
        CoolingZone: "zone-1",
    })
    if err != nil {
        t.Fatalf("Failed to register test rack: %v", err)
    }

    // Initialize test devices with different states
    devices := []types.DeviceState{
        {
            ID: "healthy-device",
            Status: "active",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    50.0,
                types.ResourceMemory: 60.0,
            },
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 1,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 45.0,
                PowerUsage:  400.0,
                CPULoad:     50.0,
                MemoryUsage: 60.0,
            },
        },
        {
            ID: "high-temp-device",
            Status: "active",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    50.0,
                types.ResourceMemory: 60.0,
            },
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 2,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 82.0,
                PowerUsage:  450.0,
                CPULoad:     70.0,
                MemoryUsage: 75.0,
            },
        },
        {
            ID: "high-load-device",
            Status: "active",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    50.0,
                types.ResourceMemory: 60.0,
            },
            Location: types.PhysicalLocation{
                Rack:     "rack-1",
                Position: 3,
                Zone:     "zone-1",
            },
            Metrics: types.DeviceMetrics{
                Temperature: 65.0,
                PowerUsage:  600.0,
                CPULoad:     92.0,
                MemoryUsage: 88.0,
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

    t.Run("Analyze State", func(t *testing.T) {
        analysis, err := analyzer.AnalyzeState(ctx)
        if err != nil {
            t.Fatalf("Failed to analyze state: %v", err)
        }

        // Verify basic analysis results
        if analysis.TotalDevices != 3 {
            t.Errorf("Expected 3 total devices, got %d", analysis.TotalDevices)
        }
        if analysis.HealthyDevices != 1 {
            t.Errorf("Expected 1 healthy device, got %d", analysis.HealthyDevices)
        }

        // Verify resource usage calculations
        if usage, exists := analysis.ResourceUsage[types.ResourceCPU]; !exists || usage == 0 {
            t.Error("CPU usage not properly calculated")
        }
        if usage, exists := analysis.ResourceUsage[types.ResourceMemory]; !exists || usage == 0 {
            t.Error("Memory usage not properly calculated")
        }
    })

    t.Run("Generate Alerts", func(t *testing.T) {
        alerts, err := analyzer.GetAlerts(ctx)
        if err != nil {
            t.Fatalf("Failed to get alerts: %v", err)
        }

        // Should have temperature alert for high-temp-device
        foundTempAlert := false
        for _, alert := range alerts {
            if alert.DeviceID == "high-temp-device" && alert.Severity == "critical" {
                foundTempAlert = true
                break
            }
        }
        if !foundTempAlert {
            t.Error("Expected critical temperature alert for high-temp-device")
        }

        // Should have CPU alert for high-load-device
        foundLoadAlert := false
        for _, alert := range alerts {
            if alert.DeviceID == "high-load-device" && alert.Severity == "critical" {
                foundLoadAlert = true
                break
            }
        }
        if !foundLoadAlert {
            t.Error("Expected critical CPU load alert for high-load-device")
        }
    })

    t.Run("Generate Recommendations", func(t *testing.T) {
        recommendations, err := analyzer.GetRecommendations(ctx)
        if err != nil {
            t.Fatalf("Failed to get recommendations: %v", err)
        }

        // Should have cooling recommendation for high-temp-device
        foundCoolingRec := false
        for _, rec := range recommendations {
            if rec.Action == "optimize_cooling" {
                for _, id := range rec.DeviceIDs {
                    if id == "high-temp-device" {
                        foundCoolingRec = true
                        break
                    }
                }
            }
        }
        if !foundCoolingRec {
            t.Error("Expected cooling optimization recommendation for high-temp-device")
        }

        // Should have workload balancing recommendation for high-load-device
        foundLoadRec := false
        for _, rec := range recommendations {
            if rec.Action == "balance_workload" {
                for _, id := range rec.DeviceIDs {
                    if id == "high-load-device" {
                        foundLoadRec = true
                        break
                    }
                }
            }
        }
        if !foundLoadRec {
            t.Error("Expected workload balancing recommendation for high-load-device")
        }
    })

    t.Run("Alert Priority", func(t *testing.T) {
        // Get alerts ordered by severity
        alerts, err := analyzer.GetAlerts(ctx)
        if err != nil {
            t.Fatalf("Failed to get alerts: %v", err)
        }

        var lastSeverity string
        for _, alert := range alerts {
            if lastSeverity != "" && alert.Severity > lastSeverity {
                t.Error("Alerts not properly ordered by severity")
            }
            lastSeverity = alert.Severity
        }
    })

    t.Run("Recommendation Priority", func(t *testing.T) {
        recommendations, err := analyzer.GetRecommendations(ctx)
        if err != nil {
            t.Fatalf("Failed to get recommendations: %v", err)
        }

        var lastPriority int
        for _, rec := range recommendations {
            if lastPriority != 0 && rec.Priority > lastPriority {
                t.Error("Recommendations not properly ordered by priority")
            }
            lastPriority = rec.Priority
        }
    })
}