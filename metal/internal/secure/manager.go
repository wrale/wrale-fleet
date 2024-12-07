package secure

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	state    TamperState
	onTamper func(TamperEvent)

	// Monitoring
	running  bool
	stopChan chan struct{}
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
		onTamper:     cfg.OnTamper,
		state: TamperState{
			CommonState: metal.CommonState{
				DeviceID:  cfg.DeviceID,
				UpdatedAt: time.Now(),
			},
			LastCheck: time.Now(),
		},
		stopChan: make(chan struct{}),
	}

	// Configure input pins
	pins := map[string]uint{
		cfg.CaseSensor:    0,
		cfg.MotionSensor:  1,
		cfg.VoltageSensor: 2,
	}

	for name, pin := range pins {
		if err := m.gpio.ConfigurePin(name, pin, metal.ModeInput); err != nil {
			return nil, fmt.Errorf("failed to configure pin %s: %w", name, err)
		}
	}

	return m, nil
}

// Start begins security monitoring
func (m *Manager) Start(ctx context.Context) error {
	m.mux.Lock()
	if m.running {
		m.mux.Unlock()
		return fmt.Errorf("already running")
	}
	m.running = true
	m.mux.Unlock()

	return m.monitor(ctx)
}

// Stop halts security monitoring
func (m *Manager) Stop() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	close(m.stopChan)
	return nil
}

// Close releases resources
func (m *Manager) Close() error {
	return m.Stop()
}

// GetState returns the current security state
func (m *Manager) GetState() TamperState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// Monitor starts security state monitoring
func (m *Manager) monitor(ctx context.Context) error {
	// Setup watch channels
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

	defer func() {
		m.gpio.UnwatchPin(m.caseSensor)
		m.gpio.UnwatchPin(m.motionSensor)
		m.gpio.UnwatchPin(m.voltSensor)
	}()

	// Monitor all sensors
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-m.stopChan:
			return nil
		
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
		m.state.UpdatedAt = time.Now()
		m.state.LastCheck = time.Now()

		if caseOpen && m.onTamper != nil {
			m.onTamper(TamperEvent{
				CommonState: m.state.CommonState,
				Type:        "CASE_OPEN",
				Severity:    "HIGH",
				Description: "Case intrusion detected",
				State:       m.state,
			})
		}
	}
}

func (m *Manager) handleMotionEvent(motion bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state.MotionDetected != motion {
		m.state.MotionDetected = motion
		m.state.UpdatedAt = time.Now()
		m.state.LastCheck = time.Now()

		if motion && m.onTamper != nil {
			m.onTamper(TamperEvent{
				CommonState: m.state.CommonState,
				Type:        "MOTION",
				Severity:    "MEDIUM",
				Description: "Motion detected",
				State:       m.state,
			})
		}
	}
}

func (m *Manager) handleVoltageEvent(voltageOK bool) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.state.VoltageNormal != voltageOK {
		m.state.VoltageNormal = voltageOK
		m.state.UpdatedAt = time.Now()
		m.state.LastCheck = time.Now()

		if !voltageOK && m.onTamper != nil {
			m.onTamper(TamperEvent{
				CommonState: m.state.CommonState,
				Type:        "VOLTAGE",
				Severity:    "HIGH",
				Description: "Voltage anomaly detected",
				State:       m.state,
			})
		}
	}
}
