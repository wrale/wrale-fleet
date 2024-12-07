package metal

import "context"

// DiagnosticManager provides hardware diagnostic capabilities
type DiagnosticManager interface {
    Monitor
    
    // Test execution
    TestGPIO(ctx context.Context) error
    TestPower(ctx context.Context) error 
    TestThermal(ctx context.Context) error
    TestSecurity(ctx context.Context) error
    RunAll(ctx context.Context) error
    RunSelected(ctx context.Context, types []TestType) error

    // Test management
    AbortTests(ctx context.Context) error
    GetTestStatus(testID string) (*TestResult, error)
    ListTestResults() ([]TestResult, error)

    // Component management
    ValidateComponent(ctx context.Context, component string) error
    CalibrateComponent(ctx context.Context, component string) error

    // Event handlers
    OnTestStart(func(TestType, string))
    OnTestComplete(func(TestResult))
}

// PowerManager provides power management capabilities
type PowerManager interface {
    Monitor
    
    // Core operations
    GetState() (PowerState, error)
    GetSource() (PowerSource, error)
    GetVoltage() (float64, error)
    GetCurrent() (float64, error)
    
    // Power control
    SetPowerMode(source PowerSource) error
    EnableCharging(enable bool) error
    
    // Monitoring
    Start(ctx context.Context) error
    Stop() error
    WatchPower(ctx context.Context) (<-chan PowerState, error)
    WatchSource(ctx context.Context) (<-chan PowerSource, error)

    // Event handlers
    OnCritical(func(PowerState))
    OnWarning(func(PowerState))
}

// ThermalManager provides thermal management capabilities
type ThermalManager interface {
    Monitor
    
    // Temperature control
    GetTemperature() (float64, error)
    SetCoolingMode(mode string) error
    
    // Fan control 
    GetFanSpeed() (uint32, error)
    SetFanSpeed(speed uint32) error
    
    // Events
    OnWarning(func(float64))
    OnCritical(func(float64))
}

// SecurityManager provides physical security monitoring
type SecurityManager interface {
    Monitor
    EventMonitor
    
    // Security state
    GetTamperState() (TamperState, error)
    ResetTamperState() error
    
    // Events
    OnTamper(func(TamperEvent))
}

// GPIO provides hardware GPIO access
type GPIO interface {
    Monitor
    
    // Pin configuration
    ConfigurePin(name string, pin uint, mode string) error
    ConfigurePWM(name string, pin uint, config *PWMConfig) error
    CreatePinGroup(name string, pins []uint) error
    
    // Pin operations
    GetPinState(name string) (bool, error)
    SetPinState(name string, state bool) error
    SetPWMDutyCycle(name string, duty uint32) error
    
    // Pin groups
    SetGroupState(name string, states []bool) error
    GetGroupState(name string) ([]bool, error)
    
    // Pin monitoring
    WatchPin(name string, mode string) (<-chan bool, error)
    UnwatchPin(name string) error
}

// PWMConfig holds PWM pin configuration
type PWMConfig struct {
    Frequency  uint32 `json:"frequency"`  // PWM frequency in Hz
    DutyCycle  uint32 `json:"duty_cycle"` // Duty cycle (0-100)
    Pull       string `json:"pull"`       // Pull-up/down configuration
    Resolution uint32 `json:"resolution"` // PWM resolution in bits
}
