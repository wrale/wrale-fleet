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
    gpio     metal.GPIO
    fanPin   string
    state    struct {
        temperature float64
        fanSpeed   uint32
        enabled    bool
    }

    // Configuration
    minTemp  float64
    maxTemp  float64
    
    // Callbacks
    onWarning  func(float64)
    onCritical func(float64)

    // Control
    stopChan chan struct{}
    running  bool
}

// New creates a new thermal manager
func New(gpio metal.GPIO, fanPin string, opts ...metal.Option) (metal.ThermalManager, error) {
    if gpio == nil {
        return nil, fmt.Errorf("GPIO controller required")
    }

    m := &Manager{
        gpio:     gpio,
        fanPin:   fanPin,
        stopChan: make(chan struct{}),
        minTemp:  40.0, // Default thresholds
        maxTemp:  80.0,
    }

    // Configure fan PWM
    if err := gpio.ConfigurePWM(fanPin, 0, &metal.PWMConfig{
        Frequency:  25000,
        DutyCycle:  0,
        Resolution: 8,
    }); err != nil {
        return nil, fmt.Errorf("failed to configure fan PWM: %w", err)
    }

    // Apply options
    for _, opt := range opts {
        if err := opt(m); err != nil {
            return nil, fmt.Errorf("option error: %w", err)
        }
    }

    return m, nil
}

// GetState implements Monitor interface
func (m *Manager) GetState() interface{} {
    m.mux.RLock()
    defer m.mux.RUnlock()
    return m.state
}

// Close implements Monitor interface
func (m *Manager) Close() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    if m.running {
        close(m.stopChan)
        m.running = false
    }

    // Stop fan
    return m.gpio.SetPWMDutyCycle(m.fanPin, 0)
}

// Temperature Control

func (m *Manager) GetTemperature() (float64, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()
    return m.state.temperature, nil
}

func (m *Manager) SetCoolingMode(mode string) error {
    m.mux.Lock()
    defer m.mux.Unlock()

    switch mode {
    case "active":
        m.state.enabled = true
    case "passive":
        m.state.enabled = false
        // Stop fan
        if err := m.gpio.SetPWMDutyCycle(m.fanPin, 0); err != nil {
            return fmt.Errorf("failed to stop fan: %w", err)
        }
        m.state.fanSpeed = 0
    default:
        return fmt.Errorf("unknown cooling mode: %s", mode)
    }

    return nil
}

// Fan Control

func (m *Manager) GetFanSpeed() (uint32, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()
    return m.state.fanSpeed, nil
}

func (m *Manager) SetFanSpeed(speed uint32) error {
    if speed > 100 {
        return fmt.Errorf("fan speed must be 0-100")
    }

    m.mux.Lock()
    defer m.mux.Unlock()

    if !m.state.enabled {
        return fmt.Errorf("cooling is disabled")
    }

    if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
        return fmt.Errorf("failed to set fan speed: %w", err)
    }

    m.state.fanSpeed = speed
    return nil
}

// Event Handlers

func (m *Manager) OnWarning(fn func(float64)) {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.onWarning = fn
}

func (m *Manager) OnCritical(fn func(float64)) {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.onCritical = fn
}

// Monitor temperature in background
func (m *Manager) Monitor(ctx context.Context) error {
    m.mux.Lock()
    if m.running {
        m.mux.Unlock()
        return fmt.Errorf("already monitoring")
    }
    m.running = true
    m.stopChan = make(chan struct{})
    m.mux.Unlock()

    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-m.stopChan:
            return nil
        case <-ticker.C:
            if err := m.updateTemperature(); err != nil {
                return fmt.Errorf("failed to update temperature: %w", err)
            }
        }
    }
}

// Stop monitoring
func (m *Manager) Stop() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    if !m.running {
        return nil
    }

    close(m.stopChan)
    m.running = false
    return nil
}

// Internal helpers

func (m *Manager) updateTemperature() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    // Read temperature (simulation for now)
    temp := m.state.temperature
    temp += 0.1 // Simulated temperature rise

    // Check thresholds
    if temp >= m.maxTemp {
        if m.onCritical != nil {
            m.onCritical(temp)
        }
        // Max cooling
        if err := m.gpio.SetPWMDutyCycle(m.fanPin, 100); err != nil {
            return fmt.Errorf("failed to set max cooling: %w", err)
        }
        m.state.fanSpeed = 100
    } else if temp >= m.minTemp {
        if m.onWarning != nil {
            m.onWarning(temp)
        }
        // Scale fan 0-100% between min and max temp
        speed := uint32((temp - m.minTemp) / (m.maxTemp - m.minTemp) * 100)
        if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
            return fmt.Errorf("failed to set fan speed: %w", err)
        }
        m.state.fanSpeed = speed
    }

    m.state.temperature = temp
    return nil
}

// Set temperature thresholds
func (m *Manager) SetThresholds(min, max float64) error {
    if min >= max {
        return fmt.Errorf("minimum temperature must be less than maximum")
    }

    m.mux.Lock()
    defer m.mux.Unlock()

    m.minTemp = min
    m.maxTemp = max
    return nil
}
