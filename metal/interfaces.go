package metal

import "context"

// Base Monitor interface
type Monitor interface {
    // GetState returns current state
    GetState() interface{}
    
    // Close releases resources
    Close() error
}

// EventMonitor extends Monitor with event capabilities
type EventMonitor interface {
    Monitor
    WatchEvents(ctx context.Context) (<-chan interface{}, error)
}

// GPIO Management

type GPIO interface {
    Monitor
    
    // Pin configuration
    ConfigurePin(name string, pin uint, mode PinMode) error
    ConfigurePWM(name string, pin uint, config *PWMConfig) error
    
    // Pin operations
    GetPinState(name string) (bool, error)
    SetPinState(name string, state bool) error
    SetPWMDutyCycle(name string, duty uint32) error
    
    // Pin groups
    CreatePinGroup(name string, pins []uint) error
    SetGroupState(name string, states []bool) error
    GetGroupState(name string) ([]bool, error)
    
    // Pin info
    GetPinMode(name string) (PinMode, error)
    GetPinConfig(name string) (*PWMConfig, error)
    ListPins() []string
    
    // Simulation control
    SetSimulated(simulated bool)
    IsSimulated() bool
}

// Power Management

type PowerManager interface {
    Monitor
    
    // State management
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
    
    // Configuration
    SetVoltageThresholds(min, max float64) error
    SetCurrentThresholds(min, max float64) error
    ConfigurePowerSource(source PowerSource, pin string) error
    EnablePowerSource(source PowerSource, enable bool) error
    
    // Events
    OnCritical(func(PowerState))
    OnWarning(func(PowerState))
}

// Thermal Management

type ThermalManager interface {
    Monitor
    
    // Temperature management
    GetTemperature() (float64, error)
    GetTemperatures() (cpu, gpu, ambient float64, err error)
    GetProfile() (ThermalProfile, error)
    
    // Cooling control
    SetFanSpeed(speed uint32) error
    SetThrottling(enabled bool) error
    SetProfile(profile ThermalProfile) error
    
    // Zone management
    AddZone(zone ThermalZone) error
    GetZone(name string) (ThermalZone, error)
    ListZones() ([]ThermalZone, error)
    
    // Monitoring
    WatchTemperature(ctx context.Context) (<-chan ThermalState, error)
    WatchZone(ctx context.Context, name string) (<-chan ThermalState, error)
    
    // Events
    OnWarning(func(ThermalEvent))
    OnCritical(func(ThermalEvent))
}

// Security Management

type SecurityManager interface {
    Monitor
    EventMonitor
    
    // State management
    GetTamperState() (TamperState, error)
    GetSecurityLevel() (SecurityLevel, error)
    ValidateState() error
    
    // Security control
    SetSecurityLevel(level SecurityLevel) error
    ClearViolations() error
    ResetTamperState() error
    
    // Monitoring
    WatchState(ctx context.Context) (<-chan TamperState, error)
    WatchSensor(ctx context.Context, name string) (<-chan bool, error)
    
    // Policy management
    SetQuietHours(windows []TimeWindow) error
    SetMotionSensitivity(sensitivity float64) error
    SetVoltageThreshold(min float64) error
    
    // Events
    OnTamper(func(TamperEvent))
    OnViolation(func(TamperEvent))
}

// Diagnostics Management

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
    GetResourceUsage(component string) (map[string]float64, error)
    MonitorResources(ctx context.Context) (<-chan map[string]float64, error)
    
    // Events
    OnTestStart(func(TestType, string))
    OnTestComplete(func(TestResult))
}

// Configuration Types

type DiagnosticManagerConfig struct {
    GPIO            GPIO
    PowerManager    PowerManager
    ThermalManager  ThermalManager
    SecurityManager SecurityManager
    RetryAttempts   int
    LoadTestTime    DurationLimit
    MinVoltage      float64
    TempRange       [2]float64
    OnTestStart     func(TestType, string)
    OnTestComplete  func(TestResult)
}

type PowerManagerConfig struct {
    GPIO            GPIO
    MonitorInterval Duration
    PowerPins       map[PowerSource]string
    VoltageMin      float64
    VoltageMax      float64
    CurrentMin      float64
    CurrentMax      float64
    OnCritical      func(PowerState)
    OnWarning       func(PowerState)
}

type ThermalManagerConfig struct {
    GPIO            GPIO
    MonitorInterval Duration
    FanControlPin   string
    ThrottlePin     string
    CPUTempPath     string
    GPUTempPath     string
    AmbientTempPath string
    DefaultProfile  ThermalProfile
    Curve           *CoolingCurve
    OnWarning       func(ThermalEvent)
    OnCritical      func(ThermalEvent)
}

type SecurityManagerConfig struct {
    GPIO            GPIO
    StateStore      StateStore
    CaseSensor      string
    MotionSensor    string
    VoltageSensor   string
    DefaultLevel    SecurityLevel
    QuietHours      []TimeWindow
    VoltageMin      float64
    Sensitivity     float64
    OnTamper        func(TamperEvent)
    OnViolation     func(TamperEvent)
}

// StateStore defines interface for persisting security state
type StateStore interface {
    SaveState(ctx context.Context, deviceID string, state interface{}) error
    LoadState(ctx context.Context, deviceID string) (interface{}, error)
    LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}
