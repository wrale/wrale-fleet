package diag

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/wrale/wrale-fleet/metal"
)

// Manager handles hardware diagnostics and testing
type Manager struct {
    mux      sync.RWMutex
    cfg      metal.DiagnosticManagerConfig
    running  bool
    results  []metal.TestResult
    testID   int

    // Test state
    currentTest    string
    onTestStart   func(metal.TestType, string)
    onTestComplete func(metal.TestResult)
}

// New creates a new hardware diagnostics manager
func New(cfg metal.DiagnosticManagerConfig, opts ...metal.Option) (metal.DiagnosticManager, error) {
    if cfg.GPIO == nil {
        return nil, fmt.Errorf("GPIO controller required")
    }

    m := &Manager{
        cfg:           cfg,
        onTestStart:   cfg.OnTestStart,
        onTestComplete: cfg.OnTestComplete,
        results:       make([]metal.TestResult, 0),
    }

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
    
    results := make([]metal.TestResult, len(m.results))
    copy(results, m.results)
    return results
}

// Close implements Monitor interface
func (m *Manager) Close() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.running = false
    m.results = nil
    m.currentTest = ""
    return nil
}

// Test execution

func (m *Manager) TestGPIO(ctx context.Context) error {
    m.startTest(metal.TestGPIO, "GPIO Subsystem")
    defer m.endTest()

    // Test digital I/O
    if err := m.testDigitalIO(); err != nil {
        return m.recordResult(metal.TestGPIO, "Digital I/O", metal.StatusFail, 0, 0, err)
    }

    // Test PWM outputs
    if err := m.testPWM(); err != nil {
        return m.recordResult(metal.TestGPIO, "PWM Output", metal.StatusFail, 0, 0, err)
    }

    return nil
}

func (m *Manager) TestPower(ctx context.Context) error {
    m.startTest(metal.TestPower, "Power Management")
    defer m.endTest()

    if m.cfg.PowerManager == nil {
        return m.recordResult(metal.TestPower, "Power Manager", metal.StatusSkipped, 0, 0, fmt.Errorf("no power manager"))
    }

    // Test power sources
    state, err := m.cfg.PowerManager.GetState()
    if err != nil {
        return m.recordResult(metal.TestPower, "Power State", metal.StatusFail, 0, 0, err)
    }

    // Check voltage
    voltage, err := m.cfg.PowerManager.GetVoltage()
    if err != nil {
        return m.recordResult(metal.TestPower, "Voltage", metal.StatusFail, voltage, m.cfg.MinVoltage, err)
    }

    if voltage < m.cfg.MinVoltage {
        return m.recordResult(metal.TestPower, "Voltage", metal.StatusFail, voltage, m.cfg.MinVoltage, 
            fmt.Errorf("voltage %.2fV below minimum %.2fV", voltage, m.cfg.MinVoltage))
    }

    return m.recordResult(metal.TestPower, "Power System", metal.StatusPass, voltage, m.cfg.MinVoltage, nil)
}

func (m *Manager) TestThermal(ctx context.Context) error {
    m.startTest(metal.TestThermal, "Thermal Management")
    defer m.endTest()

    if m.cfg.ThermalManager == nil {
        return m.recordResult(metal.TestThermal, "Thermal Manager", metal.StatusSkipped, 0, 0, fmt.Errorf("no thermal manager"))
    }

    // Get temperatures
    temp, err := m.cfg.ThermalManager.GetTemperature()
    if err != nil {
        return m.recordResult(metal.TestThermal, "Temperature", metal.StatusFail, 0, 0, err)
    }

    // Check CPU temperature range
    if temp < m.cfg.TempRange[0] || temp > m.cfg.TempRange[1] {
        return m.recordResult(metal.TestThermal, "CPU Temperature", metal.StatusFail, temp, m.cfg.TempRange[1],
            fmt.Errorf("CPU temperature %.1f°C outside range %.1f-%.1f°C", 
                temp, m.cfg.TempRange[0], m.cfg.TempRange[1]))
    }

    // Test fan control
    if err := m.testFanControl(); err != nil {
        return m.recordResult(metal.TestThermal, "Fan Control", metal.StatusFail, 0, 0, err)
    }

    return m.recordResult(metal.TestThermal, "Thermal System", metal.StatusPass, temp, m.cfg.TempRange[1], nil)
}

func (m *Manager) TestSecurity(ctx context.Context) error {
    m.startTest(metal.TestSecurity, "Security System")
    defer m.endTest()

    if m.cfg.SecurityManager == nil {
        return m.recordResult(metal.TestSecurity, "Security Manager", metal.StatusSkipped, 0, 0, fmt.Errorf("no security manager"))
    }

    // Check tamper detection 
    state, err := m.cfg.SecurityManager.GetTamperState()
    if err != nil {
        return m.recordResult(metal.TestSecurity, "Security State", metal.StatusFail, 0, 0, err)
    }

    if state.CaseOpen {
        return m.recordResult(metal.TestSecurity, "Case Tamper", metal.StatusFail, 0, 0, fmt.Errorf("case tamper detected"))
    }

    if !state.VoltageNormal {
        return m.recordResult(metal.TestSecurity, "Voltage Monitor", metal.StatusFail, 0, 0, fmt.Errorf("voltage tamper detected"))
    }

    return m.recordResult(metal.TestSecurity, "Security System", metal.StatusPass, 0, 0, nil)
}

func (m *Manager) RunAll(ctx context.Context) error {
    tests := []struct {
        name string
        fn   func(context.Context) error
    }{
        {"GPIO", m.TestGPIO},
        {"Power", m.TestPower},
        {"Thermal", m.TestThermal},
        {"Security", m.TestSecurity},
    }

    for _, test := range tests {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := test.fn(ctx); err != nil {
                return fmt.Errorf("%s test failed: %w", test.name, err)
            }
        }
    }

    return nil
}

func (m *Manager) RunSelected(ctx context.Context, types []metal.TestType) error {
    for _, t := range types {
        var err error
        switch t {
        case metal.TestGPIO:
            err = m.TestGPIO(ctx)
        case metal.TestPower:
            err = m.TestPower(ctx)
        case metal.TestThermal:
            err = m.TestThermal(ctx)
        case metal.TestSecurity:
            err = m.TestSecurity(ctx)
        default:
            return fmt.Errorf("unknown test type: %s", t)
        }
        if err != nil {
            return err
        }
    }
    return nil
}

// Test Management

func (m *Manager) AbortTests(ctx context.Context) error {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.currentTest = ""
    return nil
}

func (m *Manager) GetTestStatus(testID string) (*metal.TestResult, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()

    for _, result := range m.results {
        if result.Component == testID {
            return &result, nil
        }
    }
    return nil, fmt.Errorf("test %s not found", testID)
}

func (m *Manager) ListTestResults() ([]metal.TestResult, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()
    results := make([]metal.TestResult, len(m.results))
    copy(results, m.results)
    return results, nil
}

// Component Management

func (m *Manager) ValidateComponent(ctx context.Context, component string) error {
    return nil // TODO: Implement component validation
}

func (m *Manager) CalibrateComponent(ctx context.Context, component string) error {
    return nil // TODO: Implement component calibration
}

// Event Handling

func (m *Manager) OnTestComplete(fn func(metal.TestResult)) {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.onTestComplete = fn
}

func (m *Manager) OnTestStart(fn func(metal.TestType, string)) {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.onTestStart = fn
}

// Internal helpers

func (m *Manager) startTest(typ metal.TestType, component string) {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.currentTest = component
    m.testID++

    if m.onTestStart != nil {
        m.onTestStart(typ, component)
    }
}

func (m *Manager) endTest() {
    m.mux.Lock()
    defer m.mux.Unlock()
    m.currentTest = ""
}

func (m *Manager) recordResult(typ metal.TestType, component string, status metal.TestStatus, reading, expected float64, err error) error {
    result := metal.TestResult{
        Type:        typ,
        Component:   component,
        Status:      status,
        Reading:     reading,
        Expected:    expected,
        Description: component,
        Timestamp:   time.Now(),
    }
    if err != nil {
        result.Error = err.Error()
    }

    m.mux.Lock()
    m.results = append(m.results, result)
    if m.onTestComplete != nil {
        m.onTestComplete(result)
    }
    m.mux.Unlock()

    return err
}

// Test implementation helpers

func (m *Manager) testDigitalIO() error {
    // TODO: Implement digital I/O testing
    return nil
}

func (m *Manager) testPWM() error {
    // TODO: Implement PWM testing
    return nil
}

func (m *Manager) testFanControl() error {
    // TODO: Implement fan control testing
    return nil
}
