package client

import (
	"testing"

	"github.com/wrale/wrale-fleet/metal/power"
)

type mockPowerManager struct {
	state power.PowerState
}

// Ensure mockPowerManager implements necessary interface
var _ power.Manager = (*mockPowerManager)(nil)

func (m *mockPowerManager) GetState() power.PowerState {
	return m.state
}

func TestGetPowerState(t *testing.T) {
	mockPower := &mockPowerManager{
		state: power.PowerState{
			Voltage: 5.0,
			Current: 1.0,
		},
	}

	client := NewMetalClient(mockPower)
	if client == nil {
		t.Fatal("failed to create metal client")
	}

	state, err := client.GetPowerState()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if state == nil {
		t.Fatal("expected non-nil state")
	}

	if state.Voltage != 5.0 {
		t.Errorf("expected voltage 5.0, got %v", state.Voltage)
	}

	if state.Current != 1.0 {
		t.Errorf("expected current 1.0, got %v", state.Current)
	}
}

func TestGetPowerStateNilManager(t *testing.T) {
	client := NewMetalClient(nil)
	if client == nil {
		t.Fatal("failed to create metal client")
	}

	state, err := client.GetPowerState()
	if err == nil {
		t.Error("expected error for nil power manager")
	}
	if state != nil {
		t.Error("expected nil state for nil power manager")
	}
}
