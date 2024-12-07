package client

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/power"
)

func TestMetalClient_Basic(t *testing.T) {
	client := NewMetalClient("http://localhost:8080")

	t.Run("Power State Updates", func(t *testing.T) {
		powerState := &power.PowerState{
			State: "standby",
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

		if state.State != powerState.State {
			t.Errorf("Power state mismatch - got: %v, want: %v",
				state.State, powerState.State)
		}
	})
}

func TestMetalClient_ErrorHandling(t *testing.T) {
	// Use invalid URL to force errors
	client := NewMetalClient("http://invalid-host:9999")

	t.Run("Power State Error", func(t *testing.T) {
		powerState := &power.PowerState{
			State: "standby",
			UpdatedAt: time.Now(),
		}

		if err := client.UpdatePowerState(powerState); err == nil {
			t.Error("Expected error for invalid host")
		}
	})
}