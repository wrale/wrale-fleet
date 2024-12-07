package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/wrale/wrale-fleet/metal"
	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// Manager implements thermal management and monitoring
type Manager struct {
	sync.RWMutex
	cfg       Config
	state     ThermalState
	curve     *CoolingCurve
	zones     map[string]ThermalZone
	ctx       context.Context
	cancel    context.CancelFunc
	fanPWM    *types.PWMConfig
	enabled   bool
}

// New creates a new thermal manager
func New(cfg Config) (metal.ThermalManager, error) {
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = defaultMonitorInterval
	}
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = ProfileBalance
	}
	if cfg.FanControlPin == "" {
		return nil, fmt.Errorf("fan control pin required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		zones:   make(map[string]ThermalZone),
		enabled: true,
		fanPWM: &types.PWMConfig{
			Frequency:  25000,  // 25kHz
			DutyCycle:  0,     // Start with fan off
			Pull:      types.PullNone,
			Resolution: 8,     // 8-bit resolution
		},
	}

	// Set up fan control
	if err := cfg.GPIO.ConfigurePWM(cfg.FanControlPin, 18, m.fanPWM); err != nil {
		return nil, fmt.Errorf("failed to configure fan PWM: %v", err)
	}

	// Configure throttle pin if provided
	if cfg.ThrottlePin != "" {
		if err := cfg.GPIO.ConfigurePin(cfg.ThrottlePin, 17, types.ModeOutput); err != nil {
			return nil, fmt.Errorf("failed to configure throttle pin: %v", err)
		}
	}

	// Set initial profile
	m.state.Profile = cfg.DefaultProfile

	// Start monitoring
	go m.monitor()

	return m, nil
}

// GetThermalState returns current thermal state
func (m *Manager) GetThermalState() (metal.ThermalState, error) {
	m.RLock()
	defer m.RUnlock()
	return m.state, nil
}

// GetState implements Monitor interface
func (m *Manager) GetState() interface{} {
	m.RLock()
	defer m.RUnlock()
	return m.state
}

// GetTemperatures returns current temperature readings
func (m *Manager) GetTemperatures() (cpu, gpu, ambient float64, err error) {
	m.RLock()
	defer m.RUnlock()
	return m.state.CPUTemp, m.state.GPUTemp, m.state.AmbientTemp, nil
}

// GetProfile returns current thermal profile
func (m *Manager) GetProfile() (metal.ThermalProfile, error) {
	m.RLock()
	defer m.RUnlock()
	return metal.ThermalProfile(m.state.Profile), nil
}

// SetFanSpeed sets fan speed percentage
func (m *Manager) SetFanSpeed(speed uint32) error {
	if speed > 100 {
		return fmt.Errorf("fan speed must be 0-100")
	}

	m.Lock()
	defer m.Unlock()

	if !m.enabled {
		return fmt.Errorf("thermal management disabled")
	}

	return m.cfg.GPIO.SetPWMDutyCycle(m.cfg.FanControlPin, speed)
}

// SetThrottling enables/disables CPU throttling
func (m *Manager) SetThrottling(enabled bool) error {
	if m.cfg.ThrottlePin == "" {
		return fmt.Errorf("throttle pin not configured")
	}

	m.Lock()
	defer m.Unlock()

	if !m.enabled {
		return fmt.Errorf("thermal management disabled")
	}

	if err := m.cfg.GPIO.SetPinState(m.cfg.ThrottlePin, enabled); err != nil {
		return fmt.Errorf("failed to set throttling: %v", err)
	}

	m.state.Throttled = enabled
	return nil
}

// SetProfile changes thermal management profile
func (m *Manager) SetProfile(profile metal.ThermalProfile) error {
	m.Lock()
	defer m.Unlock()

	if !m.enabled {
		return fmt.Errorf("thermal management disabled")
	}

	m.state.Profile = Profile(profile)
	return nil
}

// AddZone adds a thermal zone definition
func (m *Manager) AddZone(zone metal.ThermalZone) error {
	m.Lock()
	defer m.Unlock()

	if !m.enabled {
		return fmt.Errorf("thermal management disabled")
	}

	m.zones[zone.Name] = ThermalZone{
		Name:       zone.Name,
		MaxTemp:    zone.MaxTemp,
		TargetTemp: zone.TargetTemp,
		Priority:   zone.Priority,
		Sensors:    zone.Sensors,
	}
	return nil
}

// GetZone returns zone configuration
func (m *Manager) GetZone(name string) (metal.ThermalZone, error) {
	m.RLock()
	defer m.RUnlock()

	zone, exists := m.zones[name]
	if !exists {
		return metal.ThermalZone{}, fmt.Errorf("zone %s not found", name)
	}
	return metal.ThermalZone(zone), nil
}

// ListZones returns all thermal zones
func (m *Manager) ListZones() ([]metal.ThermalZone, error) {
	m.RLock()
	defer m.RUnlock()

	zones := make([]metal.ThermalZone, 0, len(m.zones))
	for _, z := range m.zones {
		zones = append(zones, metal.ThermalZone(z))
	}
	return zones, nil
}

// monitor runs the thermal monitoring loop
func (m *Manager) monitor() {
	ticker := time.NewTicker(m.cfg.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if err := m.update(); err != nil {
				// TODO: Better error handling
				continue
			}
		}
	}
}

// update reads sensors and updates thermal state
func (m *Manager) update() error {
	m.Lock()
	defer m.Unlock()

	// Update temperatures
	if cpu, err := readTemp(m.cfg.CPUTempPath); err == nil {
		m.state.CPUTemp = cpu
	}
	if gpu, err := readTemp(m.cfg.GPUTempPath); err == nil {
		m.state.GPUTemp = gpu
	}
	if ambient, err := readTemp(m.cfg.AmbientTempPath); err == nil {
		m.state.AmbientTemp = ambient
	}

	m.state.UpdatedAt = time.Now()

	// Apply thermal policy
	if err := m.applyPolicy(); err != nil {
		return fmt.Errorf("failed to apply thermal policy: %v", err)
	}

	return nil
}

// applyPolicy implements thermal management logic
func (m *Manager) applyPolicy() error {
	// Get current max temperature
	maxTemp := m.state.CPUTemp
	if m.state.GPUTemp > maxTemp {
		maxTemp = m.state.GPUTemp
	}

	// Check for critical temperature
	if maxTemp >= defaultCriticalTemp {
		if m.cfg.OnCritical != nil {
			m.cfg.OnCritical(m.state)
		}
		if err := m.SetThrottling(true); err != nil {
			return err
		}
		if err := m.SetFanSpeed(maxFanSpeed); err != nil {
			return err
		}
		return nil
	}

	// Check for warning temperature
	if maxTemp >= defaultWarningTemp {
		if m.cfg.OnWarning != nil {
			m.cfg.OnWarning(m.state)
		}
	}

	// Apply cooling curve if configured
	if m.curve != nil {
		speed := calculateFanSpeed(maxTemp, m.curve)
		if err := m.SetFanSpeed(speed); err != nil {
			return err
		}
	}

	return nil
}

// Close stops thermal monitoring and releases resources
func (m *Manager) Close() error {
	m.Lock()
	defer m.Unlock()

	m.enabled = false
	m.cancel()

	// Stop fan
	if err := m.cfg.GPIO.SetPWMDutyCycle(m.cfg.FanControlPin, 0); err != nil {
		return fmt.Errorf("failed to stop fan: %v", err)
	}

	// Disable throttling
	if m.cfg.ThrottlePin != "" {
		if err := m.cfg.GPIO.SetPinState(m.cfg.ThrottlePin, false); err != nil {
			return fmt.Errorf("failed to disable throttling: %v", err)
		}
	}

	return nil
}

// WatchEvents implements Monitor interface
func (m *Manager) WatchEvents(ctx context.Context) (<-chan interface{}, error) {
	ch := make(chan interface{}, 10)
	go func() {
		ticker := time.NewTicker(m.cfg.MonitorInterval)
		defer ticker.Stop()
		defer close(ch)

		var lastState ThermalState
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.RLock()
				if m.state != lastState {
					select {
					case ch <- m.state:
						lastState = m.state
					default:
					}
				}
				m.RUnlock()
			}
		}
	}()
	return ch, nil
}

// Helper functions

func readTemp(path string) (float64, error) {
	// TODO: Implement actual temperature reading
	return 0, nil
}

func calculateFanSpeed(temp float64, curve *CoolingCurve) uint32 {
	// TODO: Implement fan curve calculation
	return minFanSpeed
}