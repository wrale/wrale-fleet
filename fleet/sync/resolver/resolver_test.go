package resolver

import (
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestConflictResolver(t *testing.T) {
    resolver := NewResolver(100)

    t.Run("Detect Conflicts", func(t *testing.T) {
        baseTime := time.Now()
        states := []*types.VersionedState{
            {
                Version: "v1",
                State: types.DeviceState{
                    ID:     "test-device",
                    Status: "active",
                    Metrics: types.DeviceMetrics{
                        Temperature: 45.0,
                        PowerUsage:  400.0,
                    },
                },
                UpdatedAt: baseTime,
            },
            {
                Version: "v2",
                State: types.DeviceState{
                    ID:     "test-device",
                    Status: "standby",
                    Metrics: types.DeviceMetrics{
                        Temperature: 45.0,
                        PowerUsage:  200.0,
                    },
                },
                UpdatedAt: baseTime,
            },
        }

        changes, err := resolver.DetectConflicts(states)
        if err != nil {
            t.Errorf("Failed to detect conflicts: %v", err)
        }
        if len(changes) == 0 {
            t.Error("Expected conflicts not detected")
        }

        for _, change := range changes {
            if _, exists := change.Changes["status"]; !exists {
                t.Error("Status change not detected")
            }
            if _, exists := change.Changes["power_usage"]; !exists {
                t.Error("Power usage change not detected")
            }
        }
    })

    t.Run("Resolve Conflicts", func(t *testing.T) {
        changes := []*types.StateChange{
            {
                PrevVersion: "v1",
                NewVersion:  "v2",
                Changes: map[string]interface{}{
                    "status":      "standby",
                    "power_usage": float64(200.0),
                },
                Timestamp: time.Now(),
            },
            {
                PrevVersion: "v1",
                NewVersion:  "v3",
                Changes: map[string]interface{}{
                    "status":      "error",
                    "temperature": float64(50.0),
                },
                Timestamp: time.Now().Add(time.Second),
            },
        }

        resolved, err := resolver.ResolveConflicts(changes)
        if err != nil {
            t.Errorf("Failed to resolve conflicts: %v", err)
        }
        if resolved == nil {
            t.Fatal("No resolved state produced")
        }

        if resolved.State.Status != "error" {
            t.Errorf("Expected status 'error', got '%s'", resolved.State.Status)
        }
        if resolved.State.Metrics.Temperature != 50.0 {
            t.Errorf("Expected temperature 50.0, got %.1f", resolved.State.Metrics.Temperature)
        }
    })

    t.Run("Validate Resolution", func(t *testing.T) {
        // Test valid state
        validState := &types.VersionedState{
            Version: "test-version",
            State: types.DeviceState{
                ID:     "test-device",
                Status: "active",
                Metrics: types.DeviceMetrics{
                    Temperature: 45.0,
                    CPULoad:     50.0,
                },
            },
            UpdatedAt: time.Now(),
            UpdatedBy: "test",
        }

        if err := resolver.ValidateResolution(validState); err != nil {
            t.Errorf("Failed to validate valid state: %v", err)
        }

        // Test invalid states
        invalidStates := []struct {
            name  string
            state *types.VersionedState
        }{
            {
                name: "missing version",
                state: &types.VersionedState{
                    State: types.DeviceState{
                        ID:     "test-device",
                        Status: "active",
                    },
                    UpdatedAt: time.Now(),
                    UpdatedBy: "test",
                },
            },
            {
                name: "missing device ID",
                state: &types.VersionedState{
                    Version: "test-version",
                    State: types.DeviceState{
                        Status: "active",
                    },
                    UpdatedAt: time.Now(),
                    UpdatedBy: "test",
                },
            },
            {
                name: "missing status",
                state: &types.VersionedState{
                    Version: "test-version",
                    State: types.DeviceState{
                        ID: "test-device",
                    },
                    UpdatedAt: time.Now(),
                    UpdatedBy: "test",
                },
            },
            {
                name: "invalid temperature",
                state: &types.VersionedState{
                    Version: "test-version",
                    State: types.DeviceState{
                        ID:     "test-device",
                        Status: "active",
                        Metrics: types.DeviceMetrics{
                            Temperature: 101.0,
                        },
                    },
                    UpdatedAt: time.Now(),
                    UpdatedBy: "test",
                },
            },
            {
                name: "invalid CPU load",
                state: &types.VersionedState{
                    Version: "test-version",
                    State: types.DeviceState{
                        ID:     "test-device",
                        Status: "active",
                        Metrics: types.DeviceMetrics{
                            CPULoad: 150.0,
                        },
                    },
                    UpdatedAt: time.Now(),
                    UpdatedBy: "test",
                },
            },
        }

        for _, tc := range invalidStates {
            t.Run(tc.name, func(t *testing.T) {
                if err := resolver.ValidateResolution(tc.state); err == nil {
                    t.Error("Expected validation error, got nil")
                }
            })
        }
    })

    t.Run("Resolution History", func(t *testing.T) {
        // Create test data
        changes := []*types.StateChange{
            {
                PrevVersion: "v1",
                NewVersion:  "v2",
                Changes: map[string]interface{}{
                    "status": "standby",
                },
                Timestamp: time.Now(),
            },
        }

        // Resolve conflicts multiple times
        for i := 0; i < 3; i++ {
            if _, err := resolver.ResolveConflicts(changes); err != nil {
                t.Errorf("Failed to resolve conflicts: %v", err)
            }
        }

        // Check history
        history := resolver.GetResolutionHistory()
        if len(history) != 3 {
            t.Errorf("Expected 3 history records, got %d", len(history))
        }

        // Verify history records
        for _, record := range history {
            if record.Changes == nil {
                t.Error("History record missing changes")
            }
            if record.Result == nil {
                t.Error("History record missing result")
            }
            if record.ResolvedAt.IsZero() {
                t.Error("History record missing timestamp")
            }
        }
    })

    t.Run("History Size Limit", func(t *testing.T) {
        smallResolver := NewResolver(2)
        changes := []*types.StateChange{
            {
                PrevVersion: "v1",
                NewVersion:  "v2",
                Changes: map[string]interface{}{
                    "status": "standby",
                },
                Timestamp: time.Now(),
            },
        }

        // Add more records than the limit
        for i := 0; i < 4; i++ {
            if _, err := smallResolver.ResolveConflicts(changes); err != nil {
                t.Errorf("Failed to resolve conflicts: %v", err)
            }
        }

        history := smallResolver.GetResolutionHistory()
        if len(history) != 2 {
            t.Errorf("Expected history length of 2, got %d", len(history))
        }

        // Verify we kept the most recent records
        for _, record := range history {
            if record.ResolvedAt.IsZero() {
                t.Error("Missing timestamp in recent history")
            }
        }
    })
}
