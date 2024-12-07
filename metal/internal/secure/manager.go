package secure

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/wrale-fleet/metal"
)

// Manager handles physical security monitoring and response
type Manager struct {
	mux  sync.RWMutex
	gpio metal.GPIO

	// Sensor pin names
	caseSensor   string
	motionSensor string
	voltSensor   string

	// Device identification
	deviceID string

	// State management
	state      metal.TamperState
	stateStore metal.StateStore

	// Callbacks for security events
	onTamper func(metal.TamperState)
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
			if tamperState, ok := state.(metal.TamperState); ok {
				m.state = tamperState
			}
		}
	}

	return m, nil
}

// GetState returns the current security state
func (m *Manager) GetState() metal.TamperState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// Monitor starts security state monitoring
func (m *Manager) Monitor(ctx context.Context) error {
	// Setup GPIO interrupt channels
	caseChan, err := m.gpio.WatchPin(m.caseSensor, metal.ModeInput)
	if err != nil {
		return fmt.Errorf("failed to monitor case sensor: %w", err)
	}

	motionChan, err := m.gpio.WatchPin(m.motionSensor, metal.ModeInput)
	if err != nil {
		return fmt.Errorf("failed to monitor motion sensor: %w", err)
	}

	voltageChan, err := m.gpio.WatchPin(m.voltSensor, metal.ModeInput)
	if err != nil {
		return fmt.Errorf("failed to monitor voltage sensor: %w", err)
	}

	// Monitor all sensors
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		
		case caseOpen := <-caseChan:
			m.handleCaseEvent(caseOpen)
		
		case motion := <-motionChan:
			m.handleMotionEvent(motion)
		
		case voltage := <-voltageChan:
			m.handleVoltageEvent(voltage)
		}
	}
}

func (m *Manager) handleCaseEvent(caseOpen bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state.CaseOpen != caseOpen {
		m.state.CaseOpen = caseOpen
		if caseOpen && m.onTamper != nil {
			m.onTamper(m.state)
		}
		m.saveState()
	}
}

func (m *Manager) handleMotionEvent(motion bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state.MotionDetected != motion {
		m.state.MotionDetected = motion
		if motion && m.onTamper != nil {
			m.onTamper(m.state)
		}
		m.saveState()
	}
}

func (m *Manager) handleVoltageEvent(voltageOK bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state.VoltageNormal != voltageOK {
		m.state.VoltageNormal = voltageOK
		if !voltageOK && m.onTamper != nil {
			m.onTamper(m.state)
		}
		m.saveState()
	}
}

func (m *Manager) saveState() {
	if m.stateStore != nil {
		if err := m.stateStore.SaveState(context.Background(), m.deviceID, m.state); err != nil {
			// Log error but continue - state storage is non-critical
			fmt.Printf("Failed to save security state: %v\n", err)
		}
	}
}

// Config represents the configuration for security manager
type Config struct {
	GPIO          metal.GPIO
	StateStore    metal.StateStore
	CaseSensor    string
	MotionSensor  string
	VoltageSensor string
	DeviceID      string
	OnTamper      func(metal.TamperState)
}
