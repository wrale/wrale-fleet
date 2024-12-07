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
    mux    sync.RWMutex
    gpio   metal.GPIO
    store  metal.StateStore

    // Sensor pin names
    caseSensor    string
    motionSensor  string
    voltSensor    string

    // Device identification
    deviceID string

    // State management
    state     metal.TamperState
    policy    SecurityPolicy
    metrics   SecurityMetrics
    onTamper  func(metal.TamperEvent)

    // Monitoring
    running   bool
    stopChan  chan struct{}
}

// Config configures the security manager
type Config struct {
    GPIO          metal.GPIO
    StateStore    metal.StateStore
    DeviceID      string
    CaseSensor    string
    MotionSensor  string
    VoltageSensor string
    OnTamper      func(metal.TamperEvent)
}

// New creates a new security manager
func New(cfg Config) (*Manager, error) {
    if cfg.GPIO == nil {
        return nil, fmt.Errorf("GPIO controller is required")
    }
    if cfg.StateStore == nil {
        return nil, fmt.Errorf("state store is required")
    }
    if cfg.DeviceID == "" {
        return nil, fmt.Errorf("device ID is required")
    }

    m := &Manager{
        gpio:         cfg.GPIO,
        store:        cfg.StateStore,
        caseSensor:   cfg.CaseSensor,
        motionSensor: cfg.MotionSensor,
        voltSensor:   cfg.VoltageSensor,
        deviceID:     cfg.DeviceID,
        onTamper:     cfg.OnTamper,
        state: metal.TamperState{
            CommonState: metal.CommonState{
                DeviceID:  cfg.DeviceID,
                UpdatedAt: time.Now(),
            },
            SecurityLevel: metal.SecurityLow,
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

// GetTamperState returns the current security state
func (m *Manager) GetTamperState() (metal.TamperState, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()
    return m.state, nil
}

// GetSecurityLevel returns the current security level
func (m *Manager) GetSecurityLevel() (metal.SecurityLevel, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()
    return m.state.SecurityLevel, nil
}

// ValidateState validates current security state
func (m *Manager) ValidateState() error {
    m.mux.RLock()
    state := m.state
    m.mux.RUnlock()

    if state.CaseOpen {
        return fmt.Errorf("case tamper detected")
    }
    if state.MotionDetected {
        return fmt.Errorf("motion detected")
    }
    if !state.VoltageNormal {
        return fmt.Errorf("voltage anomaly detected")
    }
    return nil
}

// SetSecurityLevel changes the security enforcement level
func (m *Manager) SetSecurityLevel(level metal.SecurityLevel) error {
    m.mux.Lock()
    defer m.mux.Unlock()
    
    m.state.SecurityLevel = level
    m.state.UpdatedAt = time.Now()
    
    return m.store.SaveState(context.Background(), m.deviceID, m.state)
}

// ClearViolations resets security violation flags
func (m *Manager) ClearViolations() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.state.Violations = nil
    m.state.UpdatedAt = time.Now()

    return m.store.SaveState(context.Background(), m.deviceID, m.state)
}

// ResetTamperState resets tamper detection state
func (m *Manager) ResetTamperState() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.state.CaseOpen = false
    m.state.MotionDetected = false
    m.state.VoltageNormal = true
    m.state.UpdatedAt = time.Now()
    m.state.Violations = nil

    return m.store.SaveState(context.Background(), m.deviceID, m.state)
}

// WatchState provides a channel for monitoring state changes
func (m *Manager) WatchState(ctx context.Context) (<-chan metal.TamperState, error) {
    ch := make(chan metal.TamperState, 10)
    
    go func() {
        defer close(ch)
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                m.mux.RLock()
                state := m.state
                m.mux.RUnlock()

                select {
                case ch <- state:
                default:
                }
            }
        }
    }()

    return ch, nil
}

// WatchSensor monitors a specific sensor
func (m *Manager) WatchSensor(ctx context.Context, name string) (<-chan bool, error) {
    // Validate sensor name
    switch name {
    case m.caseSensor, m.motionSensor, m.voltSensor:
        break
    default:
        return nil, fmt.Errorf("unknown sensor: %s", name)
    }
    
    // Get GPIO channel for sensor
    ch, err := m.gpio.WatchPin(name)
    if err != nil {
        return nil, fmt.Errorf("failed to watch sensor %s: %w", name, err)
    }

    return ch, nil
}

// SetQuietHours configures quiet period windows
func (m *Manager) SetQuietHours(windows []metal.TimeWindow) error {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.policy.QuietHours = windows
    return nil
}

// SetMotionSensitivity adjusts motion detection threshold
func (m *Manager) SetMotionSensitivity(sensitivity float64) error {
    if sensitivity < 0 || sensitivity > 1 {
        return fmt.Errorf("sensitivity must be between 0 and 1")
    }

    m.mux.Lock()
    defer m.mux.Unlock()

    m.policy.MotionThreshold = sensitivity
    return nil
}

// SetVoltageThreshold sets minimum voltage threshold
func (m *Manager) SetVoltageThreshold(min float64) error {
    if min <= 0 {
        return fmt.Errorf("voltage threshold must be positive")
    }

    m.mux.Lock()
    defer m.mux.Unlock()

    m.policy.VoltageRange[0] = min
    return nil
}

// monitor handles continuous security monitoring
func (m *Manager) monitor(ctx context.Context) error {
    // Setup watch channels
    caseCh, err := m.gpio.WatchPin(m.caseSensor)
    if err != nil {
        return fmt.Errorf("failed to monitor case sensor: %w", err)
    }

    motionCh, err := m.gpio.WatchPin(m.motionSensor)
    if err != nil {
        return fmt.Errorf("failed to monitor motion sensor: %w", err)
    }

    voltageCh, err := m.gpio.WatchPin(m.voltSensor)
    if err != nil {
        return fmt.Errorf("failed to monitor voltage sensor: %w", err)
    }

    // Monitor all sensors
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()

        case <-m.stopChan:
            return nil
        
        case caseOpen := <-caseCh:
            m.handleCaseEvent(caseOpen)
        
        case motion := <-motionCh:
            m.handleMotionEvent(motion)
        
        case voltage := <-voltageCh:
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
        
        if caseOpen && m.onTamper != nil {
            m.onTamper(metal.TamperEvent{
                CommonState: m.state.CommonState,
                Type:       "CASE_OPEN",
                Severity:   metal.SecurityHigh,
                Description: "Case intrusion detected",
                State:      m.state,
            })
            m.state.Violations = append(m.state.Violations, "Case tamper detected")
        }

        _ = m.store.SaveState(context.Background(), m.deviceID, m.state)
    }
}

func (m *Manager) handleMotionEvent(motion bool) {
    m.mux.Lock()
    defer m.mux.Unlock()

    if m.state.MotionDetected != motion {
        m.state.MotionDetected = motion
        m.state.UpdatedAt = time.Now()
        
        if motion && m.onTamper != nil && m.isInQuietHours() {
            m.onTamper(metal.TamperEvent{
                CommonState: m.state.CommonState,
                Type:       "MOTION",
                Severity:   metal.SecurityMedium,
                Description: "Motion detected during quiet hours",
                State:      m.state,
            })
            m.state.Violations = append(m.state.Violations, "Motion during quiet hours")
        }

        _ = m.store.SaveState(context.Background(), m.deviceID, m.state)
    }
}

func (m *Manager) handleVoltageEvent(voltageOK bool) {
    m.mux.Lock()
    defer m.mux.Unlock()

    if m.state.VoltageNormal != voltageOK {
        m.state.VoltageNormal = voltageOK
        m.state.UpdatedAt = time.Now()
        
        if !voltageOK && m.onTamper != nil {
            m.onTamper(metal.TamperEvent{
                CommonState: m.state.CommonState,
                Type:       "VOLTAGE",
                Severity:   metal.SecurityHigh,
                Description: "Voltage anomaly detected",
                State:      m.state,
            })
            m.state.Violations = append(m.state.Violations, "Voltage anomaly")
        }

        _ = m.store.SaveState(context.Background(), m.deviceID, m.state)
    }
}

func (m *Manager) isInQuietHours() bool {
    now := time.Now()
    
    for _, window := range m.policy.QuietHours {
        if now.After(window.Start) && now.Before(window.End) {
            return true
        }
    }
    
    return false
}
