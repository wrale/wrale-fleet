package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Monitor handles thermal hardware monitoring and control
type Monitor struct {
	mux   sync.RWMutex
	state ThermalState

	// Hardware interface
	gpio        GPIOController
	fanPin      string
	throttlePin string

	// Temperature paths
	cpuTemp     string
	gpuTemp     string
	ambientTemp string

	// Configuration
	monitorInterval time.Duration
	onStateChange   func(ThermalState)
}

// New creates a new thermal monitor
func New(cfg Config) (*Monitor, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}

	// Set defaults
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = defaultMonitorInterval
	}

	m := &Monitor{
		gpio:            cfg.GPIO,
		fanPin:          cfg.FanControlPin,
		throttlePin:     cfg.ThrottlePin,
		cpuTemp:         cfg.CPUTempPath,
		gpuTemp:         cfg.GPUTempPath,
		ambientTemp:     cfg.AmbientTempPath,
		monitorInterval: cfg.MonitorInterval,
		onStateChange:   cfg.OnStateChange,
	}

	if m.fanPin != "" {
		if err := m.initializeFanControl(); err != nil {
			return nil, fmt.Errorf("failed to initialize fan: %w", err)
		}
	}

	return m, nil
}

// GetState returns the current thermal state
func (m *Monitor) GetState() ThermalState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// SetFanSpeed updates the fan speed percentage
func (m *Monitor) SetFanSpeed(speed uint32) error {
	if m.fanPin == "" {
		return fmt.Errorf("fan control not configured")
	}

	if speed > 100 {
		speed = 100
	}

	if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
		return fmt.Errorf("failed to set fan speed: %w", err)
	}

	m.mux.Lock()
	m.state.FanSpeed = speed
	m.Unlock()

	return nil
}

// SetThrottling enables/disables CPU throttling
func (m *Monitor) SetThrottling(enabled bool) error {
	if m.throttlePin == "" {
		return fmt.Errorf("throttle control not configured")
	}

	if err := m.gpio.SetPinState(m.throttlePin, enabled); err != nil {
		return fmt.Errorf("failed to set throttling: %w", err)
	}

	m.mux.Lock()
	m.state.Throttled = enabled
	m.Unlock()

	return nil
}

// Monitor starts continuous hardware monitoring
func (m *Monitor) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updateThermalState(); err != nil {
				return fmt.Errorf("failed to update thermal state: %w", err)
			}
		}
	}
}

// initializeFanControl configures PWM for fan control
func (m *Monitor) initializeFanControl() error {
	if err := m.gpio.ConfigurePin(m.fanPin, 1, "pwm"); err != nil {
		return fmt.Errorf("failed to configure fan pin: %w", err)
	}
	return nil
}

// updateThermalState reads the latest temperature values
func (m *Monitor) updateThermalState() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// TODO: Implement actual temperature reading from files
	// For now using mock values
	m.state.CPUTemp = 45.0
	m.state.GPUTemp = 40.0
	m.state.AmbientTemp = 25.0
	m.state.UpdatedAt = time.Now()

	if m.onStateChange != nil {
		m.onStateChange(m.state)
	}

	return nil
}