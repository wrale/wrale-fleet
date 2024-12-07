package client

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/power"
)

// mockPowerManager implements power.Manager interface for testing
type mockPowerManager struct {
	state power.PowerState
}

// Verify interface compliance at compile time
var _ power.Manager = (*mockPowerManager)(nil)

func (m *mockPowerManager) GetState() power.PowerState {
	return m.state
}

func TestMetalClient_Basic(t *testing.T) {
	client := NewMetalClient("http://localhost:8080")

	t.Run("Power State Updates", func(t *testing.T) {
		powerState := &power.PowerState{
			State:     "standby",
			UpdatedAt: time.Now(),
		}

		if err := client.UpdatePowerState(powerState); err != nil {
			t.Errorf("Failed to update power state: %v", err)
		}

		// Get current state to verify
		state, err := client.GetPowerState()
		if err != nil {
			t.Errorf("Failed to get power state: %v", err)
		}

		if state != nil && state.State != powerState.State {
			t.Errorf("Power state mismatch - got: %v, want: %v",
				state.State, powerState.State)
		}
	})
}

func TestMetalClient_DirectHardware(t *testing.T) {
	mockPower := &mockPowerManager{
		state: power.PowerState{
			State:     "active",
			UpdatedAt: time.Now(),
		},
	}

	client := NewMetalClient("http://localhost:8080", mockPower)

	state, err := client.GetPowerState()
	if err != nil {
		t.Errorf("Failed to get power state: %v", err)
	}

	if state.State != "active" {
		t.Errorf("Power state mismatch - got: %v, want: active", state.State)
	}
}

func TestMetalClient_ErrorHandling(t *testing.T) {
	// Use invalid URL to force errors
	client := NewMetalClient("http://invalid-host:9999")

	t.Run("Power State Error", func(t *testing.T) {
		powerState := &power.PowerState{
			State:     "standby",
			UpdatedAt: time.Now(),
		}

		if err := client.UpdatePowerState(powerState); err == nil {
			t.Error("Expected error for invalid host")
		}

		// Verify error on state retrieval
		if _, err := client.GetPowerState(); err == nil {
			t.Error("Expected error getting power state from invalid host")
		}
	})
}