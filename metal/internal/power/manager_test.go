package power

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// mockGPIO implements a test GPIO controller
type mockGPIO struct {
	sync.Mutex
	pins map[string]bool
}

func newMockGPIO() *mockGPIO {
	return &mockGPIO{
		pins: make(map[string]bool),
	}
}

func (m *mockGPIO) ConfigurePin(name string, pin uint, mode metal.PinMode) error {
	return nil
}

func (m *mockGPIO) GetPinState(name string) (bool, error) {
	m.Lock()
	defer m.Unlock()
	return m.pins[name], nil
}

func (m *mockGPIO) SetPinState(name string, state bool) error {
	m.Lock()
	defer m.Unlock()
	m.pins[name] = state
	return nil
}

func (m *mockGPIO) Close() error {
	return nil
}

func (m *mockGPIO) Start(ctx context.Context) error {
	return nil
}

func (m *mockGPIO) Stop() error {
	return nil
}

func TestPowerManager(t *testing.T) {
	gpio := newMockGPIO()

	manager, err := New(Config{
		GPIO: gpio,
		PowerPins: map[metal.PowerSource]string{
			metal.MainPower:    "main_power",
			metal.BatteryPower: "battery_power",
		},
		BatteryADCPath:  "/dev/null",
		VoltageADCPath:  "/dev/null",
		CurrentADCPath:  "/dev/null",
		MonitorInterval: 100 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("Failed to create power manager: %v", err)
	}

	t.Run("Power Source Monitoring", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			errCh <- manager.Monitor(ctx)
		}()

		// Give monitor time to run
		select {
		case err := <-errCh:
			if err != nil && err != context.DeadlineExceeded {
				t.Errorf("Monitor failed: %v", err)
			}
		case <-time.After(300 * time.Millisecond):
			t.Error("Monitor did not complete in time")
		}

		// Check state is being updated
		state := manager.GetState()
		if state.UpdatedAt.IsZero() {
			t.Error("State not being updated")
		}
	})
}
