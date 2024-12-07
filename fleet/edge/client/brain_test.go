package client

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
    "github.com/wrale/wrale-fleet/fleet/edge/agent"
)

func TestBrainClient(t *testing.T) {
    deviceID := types.DeviceID("test-device-1")
    
    t.Run("Sync State", func(t *testing.T) {
        // Setup test server
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Verify request
            if r.Method != "POST" {
                t.Errorf("Expected POST request, got %s", r.Method)
            }
            if r.Header.Get("Content-Type") != "application/json" {
                t.Error("Content-Type header not set correctly")
            }
            if r.Header.Get("X-Device-ID") != string(deviceID) {
                t.Error("Device ID header not set correctly")
            }

            // Parse request body
            var state types.DeviceState
            if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
                t.Errorf("Failed to decode request body: %v", err)
            }

            w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()

        client := NewBrainClient(server.URL, deviceID, nil)
        state := types.DeviceState{
            ID: deviceID,
            Metrics: types.DeviceMetrics{
                Temperature: 45.0,
                PowerUsage:  400.0,
            },
        }

        if err := client.SyncState(state); err != nil {
            t.Errorf("SyncState failed: %v", err)
        }
    })

    t.Run("Get Commands", func(t *testing.T) {
        expectedCommands := []agent.Command{
            {
                Type:     agent.CmdUpdateState,
                Priority: 1,
                ID:       "cmd-1",
            },
            {
                Type:     agent.CmdExecuteTask,
                Priority: 2,
                ID:       "cmd-2",
            },
        }

        // Setup test server
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method != "GET" {
                t.Errorf("Expected GET request, got %s", r.Method)
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(expectedCommands)
        }))
        defer server.Close()

        client := NewBrainClient(server.URL, deviceID, nil)
        commands, err := client.GetCommands()
        if err != nil {
            t.Errorf("GetCommands failed: %v", err)
        }

        if len(commands) != len(expectedCommands) {
            t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(commands))
        }
    })

    t.Run("Report Command Result", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method != "POST" {
                t.Errorf("Expected POST request, got %s", r.Method)
            }

            var result agent.CommandResult
            if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
                t.Errorf("Failed to decode request body: %v", err)
            }

            w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()

        client := NewBrainClient(server.URL, deviceID, nil)
        result := agent.CommandResult{
            CommandID: "cmd-1",
            Success: true,
            CompletedAt: time.Now(),
        }

        if err := client.ReportCommandResult(result); err != nil {
            t.Errorf("ReportCommandResult failed: %v", err)
        }
    })

    t.Run("Report Health", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method != "POST" {
                t.Errorf("Expected POST request, got %s", r.Method)
            }

            var healthReport map[string]interface{}
            if err := json.NewDecoder(r.Body).Decode(&healthReport); err != nil {
                t.Errorf("Failed to decode request body: %v", err)
            }

            if healthy, ok := healthReport["healthy"].(bool); !ok || !healthy {
                t.Error("Health status not properly encoded")
            }

            w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()

        client := NewBrainClient(server.URL, deviceID, nil)
        details := map[string]interface{}{
            "temperature": 45.0,
            "power":      400.0,
        }

        if err := client.ReportHealth(true, details); err != nil {
            t.Errorf("ReportHealth failed: %v", err)
        }
    })

    t.Run("Get Config", func(t *testing.T) {
        expectedConfig := map[string]interface{}{
            "update_interval": 30,
            "brain_endpoint": "http://brain:8080",
        }

        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method != "GET" {
                t.Errorf("Expected GET request, got %s", r.Method)
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(expectedConfig)
        }))
        defer server.Close()

        client := NewBrainClient(server.URL, deviceID, nil)
        config, err := client.GetConfig()
        if err != nil {
            t.Errorf("GetConfig failed: %v", err)
        }

        if config["update_interval"] != expectedConfig["update_interval"] {
            t.Error("Config not properly decoded")
        }
    })

    t.Run("Retry Logic", func(t *testing.T) {
        attempts := 0
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            attempts++
            if attempts < 3 {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()

        retryConfig := &RetryConfig{
            MaxAttempts:  3,
            InitialDelay: time.Millisecond,
            MaxDelay:     time.Millisecond * 10,
        }

        client := NewBrainClient(server.URL, deviceID, retryConfig)
        err := client.SyncState(types.DeviceState{ID: deviceID})
        if err != nil {
            t.Errorf("Expected successful retry, got error: %v", err)
        }

        if attempts != 3 {
            t.Errorf("Expected 3 attempts, got %d", attempts)
        }
    })
}
