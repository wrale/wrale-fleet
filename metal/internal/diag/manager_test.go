package diag

import (
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

type mockGPIO struct {
	metal.GPIO
	initialized bool
	pins        map[string]bool
}

func newMockGPIO() *mockGPIO {
	return &mockGPIO{
		pins: make(map[string]bool),
	}
}

func (m *mockGPIO) Initialize() error {
	m.initialized = true
	return nil
}

func (m *mockGPIO) SetPinState(name string, state bool) error {
	m.pins[name] = state
	return nil
}

func (m *mockGPIO) GetPinState(name string) (bool, error) {
	return m.pins[name], nil
}

func TestNew(t *testing.T) {
	gpio := newMockGPIO()
	cfg := Config{
		GPIO:          gpio,
		RetryAttempts: 3,
		LoadTestTime:  30 * time.Second,
		MinVoltage:    4.8,
		TempRange:     [2]float64{-10, 50},
	}

	mgr, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	// Test default config values
	if mgr.cfg.RetryAttempts != 3 {
		t.Errorf("Wrong retry attempts: got %d, want 3", mgr.cfg.RetryAttempts)
	}

	if mgr.cfg.LoadTestTime != 30*time.Second {
		t.Errorf("Wrong load test time: got %v, want 30s", mgr.cfg.LoadTestTime)
	}

	if mgr.cfg.MinVoltage != 4.8 {
		t.Errorf("Wrong min voltage: got %f, want 4.8", mgr.cfg.MinVoltage)
	}

	if len(mgr.cfg.TempRange) != 2 || mgr.cfg.TempRange[0] != -10 || mgr.cfg.TempRange[1] != 50 {
		t.Errorf("Wrong temp range: got %v, want [-10, 50]", mgr.cfg.TempRange)
	}
}

func TestNewNoGPIO(t *testing.T) {
	cfg := Config{}
	mgr, err := New(cfg)
	
	if err == nil {
		t.Error("New() should fail without GPIO")
	}
	if mgr != nil {
		t.Error("Manager should be nil when creation fails")
	}
}
