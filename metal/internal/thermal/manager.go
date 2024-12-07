package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

const (
	minResponseDelay     = 100 * time.Millisecond
	defaultWarningDelay  = 5 * time.Second
	defaultCriticalDelay = 1 * time.Second
)

// Manager combines hardware monitoring and policy enforcement
type Manager struct {
	mux      sync.RWMutex
	state    metal.ThermalState
	running  bool
	stopChan chan struct{}

	// Hardware
	gpio        metal.GPIO
	fanPin      string
	throttlePin string
	zones       map[string]metal.ThermalZone

	// Configuration
	profile         metal.ThermalProfile
	monitorInterval time.Duration
	cpuTempPath     string
	gpuTempPath     string
	ambientTempPath string

	// Event handlers
	onWarning  func(metal.ThermalEvent)
	onCritical func(metal.ThermalEvent)

	// Fan control
	fanMinSpeed uint32
	fanMaxSpeed uint32
	fanStartTemp float64
	lastFanChange time.Time
}

// New creates a new thermal manager
func New(cfg metal.ThermalManagerConfig, opts ...metal.Option) (metal.ThermalManager, error) {
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
		zones:          make(map[string]metal.ThermalZone),
		stopChan:       make(chan struct{}),
		state: metal.ThermalState{
			CommonState: metal.CommonState{
				UpdatedAt: time.Now(),
			},
			Profile: cfg.DefaultProfile,
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
		m.monitorInterval = minResponseDelay
	}

	// Configure fan control
	if err := m.gpio.ConfigurePin(m.fanPin, 0, metal.ModePWM); err != nil {
		return nil, fmt.Errorf("failed to configure fan pin: %w", err)
	}

	// Configure throttle control
	if err := m.gpio.ConfigurePin(m.throttlePin, 0, metal.ModeOutput); err != nil {
		return nil, fmt.Errorf("failed to configure throttle pin: %w", err)
	}

	return m, nil
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

	return m.Monitor(ctx)
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
func (m *Manager) GetState() (metal.ThermalState, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state, nil
}

// GetTemperatures returns all temperature readings
func (m *Manager) GetTemperatures() (cpu, gpu, ambient float64, err error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state.CPUTemp, m.state.GPUTemp, m.state.AmbientTemp, nil
}

// GetProfile returns current thermal profile
func (m *Manager) GetProfile() (metal.ThermalProfile, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state.Profile, nil
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

// SetProfile changes thermal management profile
func (m *Manager) SetProfile(profile metal.ThermalProfile) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.profile = profile
	m.state.Profile = profile
	return nil
}

// Zone Management

func (m *Manager) AddZone(zone metal.ThermalZone) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, exists := m.zones[zone.Name]; exists {
		return fmt.Errorf("zone %s already exists", zone.Name)
	}

	m.zones[zone.Name] = zone
	return nil
}

func (m *Manager) GetZone(name string) (metal.ThermalZone, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	zone, exists := m.zones[name]
	if !exists {
		return metal.ThermalZone{}, fmt.Errorf("zone %s not found", name)
	}
	return zone, nil
}

func (m *Manager) ListZones() ([]metal.ThermalZone, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	zones := make([]metal.ThermalZone, 0, len(m.zones))
	for _, zone := range m.zones {
		zones = append(zones, zone)
	}
	return zones, nil
}

// Monitoring

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
			if err := m.updateThermalState(); err != nil {
				return fmt.Errorf("failed to update thermal state: %w", err)
			}
		}
	}
}

func (m *Manager) WatchTemperature(ctx context.Context) (<-chan metal.ThermalState, error) {
	ch := make(chan metal.ThermalState, 1)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(m.monitorInterval)
		defer ticker.Stop()

		var lastState metal.ThermalState
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

func (m *Manager) WatchZone(ctx context.Context, name string) (<-chan metal.ThermalState, error) {
	if _, err := m.GetZone(name); err != nil {
		return nil, err
	}
	return m.WatchTemperature(ctx)
}

// Event handlers

func (m *Manager) OnWarning(fn func(metal.ThermalEvent)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.onWarning = fn
}

func (m *Manager) OnCritical(fn func(metal.ThermalEvent)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.onCritical = fn
}

// Internal helpers

func (m *Manager) updateThermalState() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Read temperatures
	cpu := m.readTemperature(m.cpuTempPath)
	gpu := m.readTemperature(m.gpuTempPath)
	ambient := m.readTemperature(m.ambientTempPath)

	// Update state
	stateChanged := m.state.CPUTemp != cpu ||
		m.state.GPUTemp != gpu ||
		m.state.AmbientTemp != ambient

	m.state.CPUTemp = cpu
	m.state.GPUTemp = gpu
	m.state.AmbientTemp = ambient
	m.state.UpdatedAt = time.Now()

	if stateChanged {
		// Update cooling
		if err := m.updateCooling(); err != nil {
			return fmt.Errorf("cooling update failed: %w", err)
		}

		// Check thresholds for each zone
		for _, zone := range m.zones {
			if err := m.checkZoneThresholds(zone); err != nil {
				return fmt.Errorf("zone %s threshold check failed: %w", zone.Name, err)
			}
		}
	}

	return nil
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
	case metal.ProfileQuiet:
		targetSpeed = m.calculateQuietSpeed(maxTemp)
	case metal.ProfileCool:
		targetSpeed = m.calculateCoolSpeed(maxTemp)
	case metal.ProfileMax:
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

func (m *Manager) checkZoneThresholds(zone metal.ThermalZone) error {
	temp := m.getZoneTemp(zone)

	if temp >= zone.MaxTemp {
		if m.onCritical != nil {
			m.onCritical(metal.ThermalEvent{
				CommonState: metal.CommonState{
					UpdatedAt: time.Now(),
				},
				Zone:        zone.Name,
				Type:        "CRITICAL",
				Temperature: temp,
				Threshold:   zone.MaxTemp,
			})
		}
	} else if temp >= zone.TargetTemp {
		if m.onWarning != nil {
			m.onWarning(metal.ThermalEvent{
				CommonState: metal.CommonState{
					UpdatedAt: time.Now(),
				},
				Zone:        zone.Name,
				Type:        "WARNING",
				Temperature: temp,
				Threshold:   zone.TargetTemp,
			})
		}
	}

	return nil
}

func (m *Manager) getZoneTemp(zone metal.ThermalZone) float64 {
	// TODO: Implement proper zone temperature aggregation
	return m.state.CPUTemp
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
