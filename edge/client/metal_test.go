package client

import (
	"testing"

	"github.com/wrale/wrale-fleet/metal/hw/power"
)

type mockPowerManager struct {
	state *power.State
}

func (m *mockPowerManager) GetState() *power.State {
	return m.state
}

func TestGetPowerState(t *testing.T) {
	mockPower := &mockPowerManager{
		state: &power.State{
			Voltage: 5.0,
			Current: 1.0,
		},
	}

	client := NewMetalClient(mockPower)

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