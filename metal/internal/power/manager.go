package power

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Manager handles power-related operations
type Manager struct {
	mux      sync.RWMutex
	state    metal.PowerState
	running  bool
	stopChan chan struct{}

	// Hardware interface
	gpio      metal.GPIO
	powerPins map[metal.PowerSource]string

	// ADC paths
	batteryADC string
	voltageADC string
	currentADC string

	// Configuration
	monitorInterval time.Duration
	voltageMin     float64
	voltageMax     float64
	currentMin     float64
	currentMax     float64
	onPowerChange  func(metal.PowerState)
	onPowerCritical func(metal.PowerState)
}

// New creates a new power manager
func New(cfg metal.PowerManagerConfig, opts ...metal.Option) (metal.PowerManager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}

	m := &Manager{
		gpio:            cfg.GPIO,
		powerPins:       cfg.PowerPins,
		monitorInterval: cfg.MonitorInterval,
		voltageMin:      cfg.VoltageMin,
		voltageMax:      cfg.VoltageMax,
		currentMin:      cfg.CurrentMin,
		currentMax:      cfg.CurrentMax,
		onPowerCritical: cfg.OnCritical,
		onPowerChange:   cfg.OnWarning,
		stopChan:        make(chan struct{}),
		state: metal.PowerState{
			CommonState: metal.CommonState{
				UpdatedAt: time.Now(),
			},
			AvailablePower: make(map[metal.PowerSource]bool),
		},
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, fmt.Errorf("option error: %w", err)
		}
	}

	// Set monitor interval default
	if m.monitorInterval == 0 {
		m.monitorInterval = defaultMonitorInterval
	}

	// Initialize power source pins
	for source, pin := range cfg.PowerPins {
		if err := m.gpio.ConfigurePin(pin, 0, metal.ModeInput); err != nil {
			return nil, fmt.Errorf("failed to configure power pin %s: %w", pin, err)
		}
		m.state.AvailablePower[source] = false
	}

	return m, nil
}

// Start begins power monitoring
func (m *Manager) Start(ctx context.Context) error {
	m.mux.Lock()
	if m.running {
		m.mux.Unlock()
		return fmt.Errorf("already running")
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mux.Unlock()

	return m.Monitor(ctx)
}

// Stop halts power monitoring
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

// GetState returns the current power state
func (m *Manager) GetState() (metal.PowerState, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state, nil
}

// GetSource returns the current power source
func (m *Manager) GetSource() (metal.PowerSource, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state.CurrentSource, nil
}

// GetVoltage returns the current voltage
func (m *Manager) GetVoltage() (float64, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state.Voltage, nil
}

// GetCurrent returns the current draw
func (m *Manager) GetCurrent() (float64, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state.CurrentDraw, nil
}

// SetPowerMode switches to the specified power source
func (m *Manager) SetPowerMode(source metal.PowerSource) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	pin, exists := m.powerPins[source]
	if !exists {
		return fmt.Errorf("power source %s not configured", source)
	}

	// Check if source is available
	available, err := m.gpio.GetPinState(pin)
	if err != nil {
		return fmt.Errorf("failed to check power source %s: %w", source, err)
	}
	if !available {
		return fmt.Errorf("power source %s not available", source)
	}

	m.state.CurrentSource = source
	return nil
}

// EnableCharging controls battery charging
func (m *Manager) EnableCharging(enable bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.state.Charging = enable
	return nil
}

// Monitor starts monitoring power state in the background
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.stopChan:
			return nil
		case <-ticker.C:
			if err := m.updatePowerState(ctx); err != nil {
				return fmt.Errorf("failed to update power state: %w", err)
			}
		}
	}
}

// WatchPower provides power state change notifications
func (m *Manager) WatchPower(ctx context.Context) (<-chan metal.PowerState, error) {
	ch := make(chan metal.PowerState, 1)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(m.monitorInterval)
		defer ticker.Stop()

		var lastState metal.PowerState
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopChan:
				return
			case <-ticker.C:
				state, err := m.GetState()
				if err != nil {
					continue
				}
				if state != lastState {
					lastState = state
					ch <- state
				}
			}
		}
	}()

	return ch, nil
}

// WatchSource provides power source change notifications
func (m *Manager) WatchSource(ctx context.Context) (<-chan metal.PowerSource, error) {
	ch := make(chan metal.PowerSource, 1)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(m.monitorInterval)
		defer ticker.Stop()

		var lastSource metal.PowerSource
		for {
			select {
			case <-ctx.Done():
				return
			case <-m.stopChan:
				return
			case <-ticker.C:
				source, err := m.GetSource()
				if err != nil {
					continue
				}
				if source != lastSource {
					lastSource = source
					ch <- source
				}
			}
		}
	}()

	return ch, nil
}

// Configuration methods

func (m *Manager) SetVoltageThresholds(min, max float64) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if min > max {
		return fmt.Errorf("minimum voltage cannot be greater than maximum")
	}

	m.voltageMin = min
	m.voltageMax = max
	return nil
}

func (m *Manager) SetCurrentThresholds(min, max float64) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if min > max {
		return fmt.Errorf("minimum current cannot be greater than maximum")
	}

	m.currentMin = min
	m.currentMax = max
	return nil
}

func (m *Manager) ConfigurePowerSource(source metal.PowerSource, pin string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if err := m.gpio.ConfigurePin(pin, 0, metal.ModeInput); err != nil {
		return fmt.Errorf("failed to configure power pin %s: %w", pin, err)
	}

	m.powerPins[source] = pin
	m.state.AvailablePower[source] = false
	return nil
}

func (m *Manager) EnablePowerSource(source metal.PowerSource, enable bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, exists := m.powerPins[source]; !exists {
		return fmt.Errorf("power source %s not configured", source)
	}

	m.state.AvailablePower[source] = enable
	return nil
}

// Event handlers

func (m *Manager) OnCritical(fn func(metal.PowerState)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.onPowerCritical = fn
}

func (m *Manager) OnWarning(fn func(metal.PowerState)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.onPowerChange = fn
}

// Internal helpers

func (m *Manager) updatePowerState(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check power sources
	for source, pin := range m.powerPins {
		available, err := m.gpio.GetPinState(pin)
		if err != nil {
			return fmt.Errorf("failed to check power source %s: %w", source, err)
		}

		changed := m.state.AvailablePower[source] != available
		m.state.AvailablePower[source] = available

		if !available && m.state.CurrentSource == source {
			// Current source lost, need to switch
			m.handleSourceFailure()
		}
	}

	// Read voltage and current
	voltage := m.readVoltage()
	current := m.readCurrent()

	// Update state
	stateChanged := m.state.Voltage != voltage || m.state.CurrentDraw != current
	m.state.Voltage = voltage
	m.state.CurrentDraw = current
	m.state.UpdatedAt = time.Now()

	// Handle state changes
	if stateChanged {
		if m.onPowerChange != nil {
			m.onPowerChange(m.state)
		}

		// Check for critical conditions
		if voltage < m.voltageMin || voltage > m.voltageMax ||
			current < m.currentMin || current > m.currentMax {
			if m.onPowerCritical != nil {
				m.onPowerCritical(m.state)
			}
		}
	}

	return nil
}

func (m *Manager) handleSourceFailure() {
	// Try to find alternative power source
	for source, available := range m.state.AvailablePower {
		if available && source != m.state.CurrentSource {
			m.state.CurrentSource = source
			// Notify of source change
			if m.onPowerChange != nil {
				m.onPowerChange(m.state)
			}
			return
		}
	}

	// No alternative found - critical situation
	if m.onPowerCritical != nil {
		m.onPowerCritical(m.state)
	}
}

func (m *Manager) readVoltage() float64 {
	// TODO: Implement actual voltage reading
	return 5.0 // Return nominal voltage for now
}

func (m *Manager) readCurrent() float64 {
	// TODO: Implement actual current reading
	return 0.5 // Return nominal current for now
}
