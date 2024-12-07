package service

import (
    "context"
    "testing"

    "github.com/wrale/wrale-fleet/fleet/device"
    "github.com/wrale/wrale-fleet/fleet/types"
)

// mockMetalClient for testing
type mockMetalClient struct {
    execErrors   map[types.DeviceID]error
    deviceErrors map[types.DeviceID]error
    metrics      map[types.DeviceID]types.DeviceMetrics
}

func newMockMetalClient() *mockMetalClient {
    return &mockMetalClient{
        execErrors:   make(map[types.DeviceID]error),
        deviceErrors: make(map[types.DeviceID]error),
        metrics:      make(map[types.DeviceID]types.DeviceMetrics),
    }
}

func (m *mockMetalClient) ExecuteOperation(ctx context.Context, deviceID types.DeviceID, operation string) error {
    if err, exists := m.execErrors[deviceID]; exists {
        return err
    }
    return nil
}

func (m *mockMetalClient) GetDeviceMetrics(ctx context.Context, deviceID types.DeviceID) (*types.DeviceMetrics, error) {
    if err, exists := m.deviceErrors[deviceID]; exists {
        return nil, err
    }
    if metrics, exists := m.metrics[deviceID]; exists {
        return &metrics, nil
    }
    return &types.DeviceMetrics{
        Temperature: 45.0,
        PowerUsage:  400.0,
        CPULoad:     50.0,
        MemoryUsage: 60.0,
    }, nil
}

func TestService(t *testing.T) {
    ctx := context.Background()
    metalClient := newMockMetalClient()
    service := NewService(metalClient)

    // Test data
    testDevice := types.DeviceState{
        ID: "test-device-1",
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
    }

    testRack := device.RackConfig{
        MaxUnits:    42,
        PowerLimit:  5000.0,
        CoolingZone: "zone-1",
    }

    t.Run("Device Management", func(t *testing.T) {
        // Register device
        err := service.RegisterDevice(ctx, testDevice)
        if err != nil {
            t.Fatalf("Failed to register device: %v", err)
        }

        // Get device state
        state, err := service.GetDeviceState(ctx, testDevice.ID)
        if err != nil {
            t.Fatalf("Failed to get device state: %v", err)
        }
        if state == nil {
            t.Fatal("Device state not found")
        }

        // List devices
        devices, err := service.ListDevices(ctx)
        if err != nil {
            t.Fatalf("Failed to list devices: %v", err)
        }
        if len(devices) != 1 {
            t.Errorf("Expected 1 device, got %d", len(devices))
        }

        // Update device state
        testDevice.Status = "updated"
        err = service.UpdateDeviceState(ctx, testDevice)
        if err != nil {
            t.Fatalf("Failed to update device state: %v", err)
        }

        // Verify update
        state, err = service.GetDeviceState(ctx, testDevice.ID)
        if err != nil {
            t.Fatalf("Failed to get updated device state: %v", err)
        }
        if state.Status != "updated" {
            t.Errorf("Expected status 'updated', got '%s'", state.Status)
        }
    })

    t.Run("Physical Management", func(t *testing.T) {
        // Register rack
        err := service.RegisterRack(ctx, "rack-1", testRack)
        if err != nil {
            t.Fatalf("Failed to register rack: %v", err)
        }

        // Update device location
        newLocation := types.PhysicalLocation{
            Rack:     "rack-1",
            Position: 2,
            Zone:     "zone-1",
        }
        err = service.UpdateDeviceLocation(ctx, testDevice.ID, newLocation)
        if err != nil {
            t.Fatalf("Failed to update device location: %v", err)
        }

        // Get device location
        location, err := service.GetDeviceLocation(ctx, testDevice.ID)
        if err != nil {
            t.Fatalf("Failed to get device location: %v", err)
        }
        if location.Position != 2 {
            t.Errorf("Expected position 2, got %d", location.Position)
        }

        // Get devices in zone
        zoneDevices, err := service.GetDevicesInZone(ctx, "zone-1")
        if err != nil {
            t.Fatalf("Failed to get devices in zone: %v", err)
        }
        if len(zoneDevices) != 1 {
            t.Errorf("Expected 1 device in zone, got %d", len(zoneDevices))
        }
    })

    t.Run("Task Management", func(t *testing.T) {
        task := types.Task{
            ID:        "task-1",
            DeviceIDs: []types.DeviceID{testDevice.ID},
            Operation: "test_operation",
            Priority:  1,
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    20.0,
                types.ResourceMemory: 30.0,
            },
        }

        // Schedule task
        err := service.ScheduleTask(ctx, task)
        if err != nil {
            t.Fatalf("Failed to schedule task: %v", err)
        }

        // Execute task
        err = service.ExecuteTask(ctx, task)
        if err != nil {
            t.Fatalf("Failed to execute task: %v", err)
        }

        // Get task
        taskEntry, err := service.GetTask(ctx, task.ID)
        if err != nil {
            t.Fatalf("Failed to get task: %v", err)
        }
        if taskEntry == nil {
            t.Fatal("Task not found")
        }

        // List tasks
        tasks, err := service.ListTasks(ctx)
        if err != nil {
            t.Fatalf("Failed to list tasks: %v", err)
        }
        if len(tasks) != 1 {
            t.Errorf("Expected 1 task, got %d", len(tasks))
        }
    })

    t.Run("Analysis and Optimization", func(t *testing.T) {
        // Analyze fleet
        analysis, err := service.AnalyzeFleet(ctx)
        if err != nil {
            t.Fatalf("Failed to analyze fleet: %v", err)
        }
        if analysis.TotalDevices != 1 {
            t.Errorf("Expected 1 device in analysis, got %d", analysis.TotalDevices)
        }

        // Get alerts
        alerts, err := service.GetAlerts(ctx)
        if err != nil {
            t.Fatalf("Failed to get alerts: %v", err)
        }
        if len(alerts) > 0 {
            t.Errorf("Expected no alerts for healthy device, got %d", len(alerts))
        }

        // Get recommendations
        recommendations, err := service.GetRecommendations(ctx)
        if err != nil {
            t.Fatalf("Failed to get recommendations: %v", err)
        }

        // Test placement suggestions
        task := types.Task{
            ID: "placement-test",
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    20.0,
                types.ResourceMemory: 30.0,
            },
        }

        placements, err := service.SuggestPlacements(ctx, task)
        if err != nil {
            t.Fatalf("Failed to suggest placements: %v", err)
        }
        if len(placements) == 0 {
            t.Error("No placement suggestions provided")
        }
    })
}