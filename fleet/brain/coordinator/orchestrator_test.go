package coordinator

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// mockMetalClient implements MetalClient interface for testing
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

func TestOrchestrator(t *testing.T) {
    ctx := context.Background()
    scheduler := NewScheduler()
    stateManager := NewStateManager()
    metalClient := newMockMetalClient()
    orchestrator := NewOrchestrator(scheduler, stateManager, metalClient)

    // Initialize test device
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
    }

    // Add device to state manager
    err := stateManager.UpdateDeviceState(ctx, testDevice)
    if err != nil {
        t.Fatalf("Failed to initialize device state: %v", err)
    }

    t.Run("Execute Successful Task", func(t *testing.T) {
        task := types.Task{
            ID:        "task-1",
            DeviceIDs: []types.DeviceID{testDevice.ID},
            Operation: "test_operation",
            Priority:  1,
            Resources: map[types.ResourceType]float64{
                types.ResourceCPU:    20.0,
                types.ResourceMemory: 30.0,
            },
            Status:    "pending",
            CreatedAt: time.Now(),
        }

        // Schedule and execute task
        err := scheduler.Schedule(ctx, task)
        if err != nil {
            t.Fatalf("Failed to schedule task: %v", err)
        }

        err = orchestrator.ExecuteTask(ctx, task)
        if err != nil {
            t.Errorf("Failed to execute task: %v", err)
        }

        // Verify task completed
        taskEntry, err := scheduler.GetTask(ctx, task.ID)
        if err != nil {
            t.Errorf("Failed to get task: %v", err)
        }
        if taskEntry.State != TaskStateCompleted {
            t.Errorf("Expected task state %s, got %s", TaskStateCompleted, taskEntry.State)
        }

        // Verify device state updated
        deviceState, err := stateManager.GetDeviceState(ctx, testDevice.ID)
        if err != nil {
            t.Errorf("Failed to get device state: %v", err)
        }
        if deviceState.LastUpdated.IsZero() {
            t.Error("Device state not updated after task execution")
        }
    })

    t.Run("Execute Task with Device Error", func(t *testing.T) {
        // Set up error for device
        expectedErr := fmt.Errorf("device operation failed")
        metalClient.execErrors[testDevice.ID] = expectedErr

        task := types.Task{
            ID:        "task-2",
            DeviceIDs: []types.DeviceID{testDevice.ID},
            Operation: "test_operation",
            Priority:  1,
            CreatedAt: time.Now(),
        }

        // Schedule and execute task
        err := scheduler.Schedule(ctx, task)
        if err != nil {
            t.Fatalf("Failed to schedule task: %v", err)
        }

        err = orchestrator.ExecuteTask(ctx, task)
        if err == nil {
            t.Error("Expected error from task execution, got nil")
        }

        // Verify task marked as failed
        taskEntry, err := scheduler.GetTask(ctx, task.ID)
        if err != nil {
            t.Errorf("Failed to get task: %v", err)
        }
        if taskEntry.State != TaskStateFailed {
            t.Errorf("Expected task state %s, got %s", TaskStateFailed, taskEntry.State)
        }

        // Clean up error
        delete(metalClient.execErrors, testDevice.ID)
    })

    t.Run("Execute Task with Multiple Devices", func(t *testing.T) {
        // Add second test device
        testDevice2 := testDevice
        testDevice2.ID = "test-device-2"
        err := stateManager.UpdateDeviceState(ctx, testDevice2)
        if err != nil {
            t.Fatalf("Failed to initialize second device state: %v", err)
        }

        task := types.Task{
            ID:        "task-3",
            DeviceIDs: []types.DeviceID{testDevice.ID, testDevice2.ID},
            Operation: "test_operation",
            Priority:  1,
            CreatedAt: time.Now(),
        }

        // Schedule and execute task
        err = scheduler.Schedule(ctx, task)
        if err != nil {
            t.Fatalf("Failed to schedule task: %v", err)
        }

        err = orchestrator.ExecuteTask(ctx, task)
        if err != nil {
            t.Errorf("Failed to execute multi-device task: %v", err)
        }

        // Verify both devices updated
        for _, deviceID := range task.DeviceIDs {
            state, err := stateManager.GetDeviceState(ctx, deviceID)
            if err != nil {
                t.Errorf("Failed to get device state for %s: %v", deviceID, err)
            }
            if state.LastUpdated.IsZero() {
                t.Errorf("Device %s state not updated after task execution", deviceID)
            }
        }
    })

    t.Run("Execute Task with Invalid Device", func(t *testing.T) {
        task := types.Task{
            ID:        "task-4",
            DeviceIDs: []types.DeviceID{"invalid-device"},
            Operation: "test_operation",
            Priority:  1,
            CreatedAt: time.Now(),
        }

        // Schedule and execute task
        err := scheduler.Schedule(ctx, task)
        if err != nil {
            t.Fatalf("Failed to schedule task: %v", err)
        }

        err = orchestrator.ExecuteTask(ctx, task)
        if err == nil {
            t.Error("Expected error when executing task with invalid device")
        }

        // Verify task marked as failed
        taskEntry, err := scheduler.GetTask(ctx, task.ID)
        if err != nil {
            t.Errorf("Failed to get task: %v", err)
        }
        if taskEntry.State != TaskStateFailed {
            t.Errorf("Expected task state %s, got %s", TaskStateFailed, taskEntry.State)
        }
    })
}