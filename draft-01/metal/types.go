package metal

import "time"

// Option configures a hardware component
type Option func(interface{}) error

// CommonState contains fields shared by all hardware states
type CommonState struct {
    DeviceID  string    `json:"device_id"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Test Types

type TestType string

const (
    TestGPIO     TestType = "GPIO"
    TestPower    TestType = "POWER"
    TestThermal  TestType = "THERMAL"
    TestSecurity TestType = "SECURITY"
)

type TestStatus string

const (
    StatusPass    TestStatus = "PASS"
    StatusFail    TestStatus = "FAIL"
    StatusWarning TestStatus = "WARNING"
    StatusSkipped TestStatus = "SKIPPED"
)

type TestResult struct {
    CommonState
    Type        TestType      `json:"type"`
    Component   string        `json:"component"`
    Status      TestStatus    `json:"status"`
    Reading     float64       `json:"reading,omitempty"`
    Expected    float64       `json:"expected,omitempty"`
    Duration    time.Duration `json:"duration"`
    Description string        `json:"description"`
    Error       string        `json:"error,omitempty"`
}

// Power Types

type PowerSource string

const (
    MainPower    PowerSource = "MAIN"
    BatteryPower PowerSource = "BATTERY"
    SolarPower   PowerSource = "SOLAR"
)

type PowerState struct {
    CommonState
    BatteryLevel     float64                `json:"battery_level"`
    Charging         bool                    `json:"charging"`
    Voltage          float64                `json:"voltage"`
    CurrentDraw      float64                `json:"current_draw"`
    CurrentSource    PowerSource            `json:"current_source"`
    AvailablePower   map[PowerSource]bool   `json:"available_power"`
    PowerConsumption float64                `json:"power_consumption"`
    Warnings         []string               `json:"warnings,omitempty"`
}

type PowerEvent struct {
    CommonState
    Source    PowerSource `json:"source"`
    Type      string     `json:"type"`
    Reading   float64    `json:"reading"`
    Threshold float64    `json:"threshold"`
    Message   string     `json:"message,omitempty"`
}

// Security Types

type SecurityLevel string

const (
    SecurityLow    SecurityLevel = "LOW"
    SecurityMedium SecurityLevel = "MEDIUM"
    SecurityHigh   SecurityLevel = "HIGH"
)

type TamperState struct {
    CommonState
    CaseOpen       bool          `json:"case_open"`
    MotionDetected bool          `json:"motion_detected"`
    VoltageNormal  bool          `json:"voltage_normal"`
    SecurityLevel  SecurityLevel `json:"security_level"`
    Violations     []string      `json:"violations,omitempty"`
}

type TimeWindow struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type TamperEvent struct {
    CommonState
    Type        string        `json:"type"`
    Severity    SecurityLevel `json:"severity"`
    Description string        `json:"description"`
    State       TamperState   `json:"state"`
    Details     interface{}   `json:"details,omitempty"`
}

// Thermal Types

type ThermalProfile string

const (
    ProfileQuiet   ThermalProfile = "QUIET"
    ProfileBalance ThermalProfile = "BALANCE"
    ProfileCool    ThermalProfile = "COOL"
    ProfileMax     ThermalProfile = "MAX"
)

type ThermalState struct {
    CommonState
    CPUTemp      float64        `json:"cpu_temp"`
    GPUTemp      float64        `json:"gpu_temp"`
    AmbientTemp  float64        `json:"ambient_temp"`
    FanSpeed     uint32         `json:"fan_speed"`
    Throttled    bool           `json:"throttled"`
    Warnings     []string       `json:"warnings,omitempty"`
    Profile      ThermalProfile `json:"profile"`
}

type ThermalZone struct {
    Name       string   `json:"name"`
    MaxTemp    float64  `json:"max_temp"`
    TargetTemp float64  `json:"target_temp"`
    Priority   int      `json:"priority"`
    Sensors    []string `json:"sensors"`
}

type ThermalEvent struct {
    CommonState
    Zone        string     `json:"zone"`
    Type        string     `json:"type"`
    Temperature float64    `json:"temperature"`
    Threshold   float64    `json:"threshold"`
    Message     string     `json:"message,omitempty"`
}

type CoolingCurve struct {
    Points       []float64            `json:"points"`
    Speeds       []uint32             `json:"speeds"`
    ZoneWeights  map[string]float64   `json:"zone_weights,omitempty"`
    Hysteresis   float64              `json:"hysteresis"`
    SmoothSteps  int                  `json:"smooth_steps"`
    RampTime     time.Duration        `json:"ramp_time"`
}

// GPIO Types

type PinMode string

const (
    ModeInput  PinMode = "INPUT"
    ModeOutput PinMode = "OUTPUT"
    ModePWM    PinMode = "PWM"
)

type PullMode string

const (
    PullNone PullMode = "NONE"
    PullUp   PullMode = "UP"
    PullDown PullMode = "DOWN"
)

type PWMConfig struct {
    Frequency  uint32   `json:"frequency"`
    DutyCycle  uint32   `json:"duty_cycle"`
    Pull       PullMode `json:"pull"`
    Resolution uint32   `json:"resolution"`
}

// Configuration Types

type DiagnosticManagerConfig struct {
    GPIO            GPIO
    PowerManager    PowerManager
    ThermalManager  ThermalManager
    SecurityManager SecurityManager
    RetryAttempts   int
    LoadTestTime    time.Duration
    MinVoltage      float64
    TempRange       [2]float64
    OnTestStart     func(TestType, string)
    OnTestComplete  func(TestResult)
}

type PowerManagerConfig struct {
    GPIO            GPIO
    MonitorInterval time.Duration
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
    MonitorInterval time.Duration
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
