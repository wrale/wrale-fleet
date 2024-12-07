package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Manager handles thermal monitoring and control
type Manager struct {
	mux      sync.RWMutex
	state    ThermalState
	running  bool
	stopChan chan struct{}

	// Hardware
	gpio        metal.GPIO
	fanPin      string
	throttlePin string

	// Configuration
	profile         Profile
	monitorInterval time.Duration
	cpuTempPath     string
	gpuTempPath     string
	ambientTempPath string

	// Event handlers
	onWarning  func(ThermalState)
	onCritical func(ThermalState)

	// Fan control
	fanMinSpeed uint32
	fanMaxSpeed uint32
	fanStartTemp float64
	lastFanChange time.Time
}

// New creates a new thermal manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller required")
	}

	m := &Manager{
		gpio:            cfg.GPIO,
		fanPin:         cfg.FanControlPin,
		throttlePin:    cfg.ThrottlePin,
		cpuTempPath:    cfg.CPUTempPath,
		gpuTempPath:    cfg.GPUTempPath,
		ambientTempPath: cfg.AmbientTempPath,
		profile:        cfg.DefaultProfile,
		monitorInterval: cfg.MonitorInterval,
		onWarning:      cfg.OnWarning,
		onCritical:     cfg.OnCritical,
		stopChan:       make(chan struct{}),
		state: ThermalState{
			CommonState: metal.CommonState{
				UpdatedAt: time.Now(),
			},
			Profile: cfg.DefaultProfile,
		},
	}

	// Set monitor interval default
	if m.monitorInterval == 0 {
		m.monitorInterval = minResponseDelay
	}

	// Configure fan control
	if err := m.gpio.ConfigurePWM(m.fanPin, 0, &metal.PWMConfig{
		Frequency:  25000,
		DutyCycle:  0,
		Resolution: 100,
	}); err != nil {
		return nil, fmt.Errorf("failed to configure fan pin: %w", err)
	}

	// Configure throttle control
	if err := m.gpio.ConfigurePin(m.throttlePin, 0, metal.ModeOutput); err != nil {
		return nil, fmt.Errorf("failed to configure throttle pin: %w", err)
	}

	return m, nil
}

// Close releases resources
func (m *Manager) Close() error {
	if err := m.Stop(); err != nil {
		return err
	}

	// Set fan to safe state
	if err := m.SetFanSpeed(0); err != nil {
		return fmt.Errorf("failed to stop fan: %w", err)
	}

	// Disable throttling
	if err := m.SetThrottling(false); err != nil {
		return fmt.Errorf("failed to disable throttling: %w", err)
	}

	return nil
}

// Start begins thermal monitoring
func (m *Manager) Start(ctx context.Context) error {
	m.mux.Lock()
	if m.running {
		m.mux.Unlock()
		return fmt.Errorf("already running")
	}
	m.running = true
	m.stopChan = make(chan struct{})
	m.mux.Unlock()

	return m.monitor(ctx)
}

// Stop halts thermal monitoring
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

// GetState returns current thermal state
func (m *Manager) GetState() ThermalState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// SetFanSpeed updates cooling fan speed
func (m *Manager) SetFanSpeed(speed uint32) error {
	if speed > 100 {
		return fmt.Errorf("fan speed must be 0-100")
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
		return fmt.Errorf("failed to set fan speed: %w", err)
	}

	m.state.FanSpeed = speed
	m.lastFanChange = time.Now()
	return nil
}

// SetThrottling enables/disables CPU throttling
func (m *Manager) SetThrottling(enabled bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if err := m.gpio.SetPinState(m.throttlePin, enabled); err != nil {
		return fmt.Errorf("failed to set throttling: %w", err)
	}

	m.state.Throttled = enabled
	return nil
}

func (m *Manager) monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.stopChan:
			return nil
		case <-ticker.C:
			if err := m.updateThermalState(); err != nil {
				return fmt.Errorf("failed to update thermal state: %w", err)
			}
		}
	}
}

func (m *Manager) updateThermalState() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Read temperatures
	cpu := m.readTemperature(m.cpuTempPath)
	gpu := m.readTemperature(m.gpuTempPath)
	ambient := m.readTemperature(m.ambientTempPath)

	// Track old state for comparison
	oldState := m.state

	// Update state
	m.state.CPUTemp = cpu
	m.state.GPUTemp = gpu
	m.state.AmbientTemp = ambient
	m.state.UpdatedAt = time.Now()

	// Check for changes
	if !m.compareStates(oldState, m.state) {
		// Update cooling
		if err := m.updateCooling(); err != nil {
			return fmt.Errorf("cooling update failed: %w", err)
		}

		// Check thresholds
		m.checkThresholds()
	}

	return nil
}

func (m *Manager) compareStates(old, new ThermalState) bool {
	return old.CPUTemp == new.CPUTemp &&
		old.GPUTemp == new.GPUTemp &&
		old.AmbientTemp == new.AmbientTemp &&
		old.FanSpeed == new.FanSpeed &&
		old.Throttled == new.Throttled
}

func (m *Manager) readTemperature(path string) float64 {
	// TODO: Implement actual temperature reading
	return 45.0 // Return nominal temperature for now
}

func (m *Manager) updateCooling() error {
	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	var targetSpeed uint32
	switch m.profile {
	case ProfileQuiet:
		targetSpeed = m.calculateQuietSpeed(maxTemp)
	case ProfileCool:
		targetSpeed = m.calculateCoolSpeed(maxTemp)
	case ProfileMax:
		targetSpeed = 100
	default: // ProfileBalance
		targetSpeed = m.calculateBalanceSpeed(maxTemp)
	}

	if targetSpeed != m.state.FanSpeed {
		if err := m.SetFanSpeed(targetSpeed); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) checkThresholds() {
	// Example thresholds
	const (
		warningTemp  = 70.0
		criticalTemp = 80.0
	)

	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	if maxTemp >= criticalTemp {
		if m.onCritical != nil {
			m.onCritical(m.state)
		}
	} else if maxTemp >= warningTemp {
		if m.onWarning != nil {
			m.onWarning(m.state)
		}
	}
}

func (m *Manager) calculateQuietSpeed(temp float64) uint32 {
	if temp < m.fanStartTemp {
		return 0
	}
	speed := uint32((temp - m.fanStartTemp) * 5.0)
	if speed > 60 { // Cap quiet mode at 60%
		speed = 60
	}
	return speed
}

func (m *Manager) calculateCoolSpeed(temp float64) uint32 {
	if temp < m.fanStartTemp {
		return 20 // Minimum 20% in cool mode
	}
	speed := uint32((temp - m.fanStartTemp) * 10.0)
	if speed > 100 {
		speed = 100
	}
	return speed
}

func (m *Manager) calculateBalanceSpeed(temp float64) uint32 {
	if temp < m.fanStartTemp {
		return 10 // Minimum 10% in balanced mode
	}
	speed := uint32((temp - m.fanStartTemp) * 7.5)
	if speed > 100 {
		speed = 100
	}
	return speed
}
