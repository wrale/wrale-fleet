package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

func TestMetalClient(t *testing.T) {
	t.Run("Get Metrics", func(t *testing.T) {
		expectedMetrics := types.DeviceMetrics{
			Temperature: 45.0,
			PowerUsage:  400.0,
			CPULoad:     50.0,
			MemoryUsage: 60.0,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedMetrics)
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)
		metrics, err := client.GetMetrics()
		if err != nil {
			t.Errorf("GetMetrics failed: %v", err)
		}

		if metrics.Temperature != expectedMetrics.Temperature {
			t.Errorf("Expected temperature %.2f, got %.2f",
				expectedMetrics.Temperature, metrics.Temperature)
		}
	})

	t.Run("Update Power State", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("Expected PUT request, got %s", r.Method)
			}

			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if state := payload["state"]; state != "standby" {
				t.Errorf("Expected state 'standby', got '%s'", state)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)
		if err := client.UpdatePowerState("standby"); err != nil {
			t.Errorf("UpdatePowerState failed: %v", err)
		}
	})

	t.Run("Execute Operation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if op := payload["operation"]; op != "test_operation" {
				t.Errorf("Expected operation 'test_operation', got '%s'", op)
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)
		if err := client.ExecuteOperation("test_operation"); err != nil {
			t.Errorf("ExecuteOperation failed: %v", err)
		}
	})

	t.Run("Get Health Status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)
		healthy, err := client.GetHealthStatus()
		if err != nil {
			t.Errorf("GetHealthStatus failed: %v", err)
		}
		if !healthy {
			t.Error("Expected healthy status")
		}
	})

	t.Run("Run Diagnostics", func(t *testing.T) {
		expectedResults := map[string]interface{}{
			"hardware_check":           true,
			"temperature_within_range": true,
			"power_stable":             true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedResults)
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)
		results, err := client.RunDiagnostics()
		if err != nil {
			t.Errorf("RunDiagnostics failed: %v", err)
		}

		for key, expected := range expectedResults {
			if actual, ok := results[key]; !ok || actual != expected {
				t.Errorf("Expected %v for %s, got %v", expected, key, actual)
			}
		}
	})

	t.Run("Error Handling", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewMetalClient(server.URL)

		// Test various operations with error response
		if _, err := client.GetMetrics(); err == nil {
			t.Error("Expected error from GetMetrics")
		}

		if err := client.UpdatePowerState("standby"); err == nil {
			t.Error("Expected error from UpdatePowerState")
		}

		if err := client.ExecuteOperation("test"); err == nil {
			t.Error("Expected error from ExecuteOperation")
		}

		if _, err := client.GetHealthStatus(); err == nil {
			t.Error("Expected error from GetHealthStatus")
		}
	})
}
