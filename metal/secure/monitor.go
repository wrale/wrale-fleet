package secure

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager is the base hardware security monitor
type Manager struct {
	mux  sync.RWMutex
	gpio GPIOController

	// Sensor pin names
	caseSensor   string
	motionSensor string
	voltSensor   string

	// Device identification
	deviceID string

	// State management
	state      TamperState
	stateStore StateStore

	// Callbacks for security events
	onTamper func(TamperState)
}

// New creates a new security manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	m := &Manager{
		gpio:         cfg.GPIO,
		caseSensor:   cfg.CaseSensor,
		motionSensor: cfg.MotionSensor,
		voltSensor:   cfg.VoltageSensor,
		deviceID:     cfg.DeviceID,
		stateStore:   cfg.StateStore,
		onTamper:     cfg.OnTamper,
	}

	// Load last known state if store is available
	if m.stateStore != nil {
		if state, err := m.stateStore.LoadState(context.Background(), m.deviceID); err == nil {
			m.state = state
		}
	}

	return m, nil
}

// Monitor starts continuous security monitoring of hardware
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.checkSecurity(ctx); err != nil {
				return fmt.Errorf("security check failed: %w", err)
			}
		}
	}
}

// checkSecurity performs raw hardware security checks
func (m *Manager) checkSecurity(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check case sensor
	caseOpen, err := m.gpio.GetPinState(m.caseSensor)
	if err != nil {
		return fmt.Errorf("failed to check case sensor: %w", err)
	}

	// Check motion sensor
	motion, err := m.gpio.GetPinState(m.motionSensor)
	if err != nil {
		return fmt.Errorf("failed to check motion sensor: %w", err)
	}

	// Check voltage sensor
	voltageOK, err := m.gpio.GetPinState(m.voltSensor)
	if err != nil {
		return fmt.Errorf("failed to check voltage sensor: %w", err)
	}

	// Update state
	m.state = TamperState{
		CaseOpen:       caseOpen,
		MotionDetected: motion,
		VoltageNormal:  voltageOK,
		LastCheck:      time.Now(),
	}

	// Save state if configured
	if m.stateStore != nil {
		if err := m.stateStore.SaveState(ctx, m.deviceID, m.state); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	// Notify of raw state changes through callback
	if m.onTamper != nil && (caseOpen || motion || !voltageOK) {
		m.onTamper(m.state)
	}

	return nil
}

// GetState returns the current security state
func (m *Manager) GetState() TamperState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}