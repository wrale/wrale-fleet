// Package thermal provides thermal management and policy enforcement
package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/core/policy"
)

// ThermalState represents the current thermal status
type ThermalState struct {
	CPUTemp     float64
	GPUTemp     float64
	AmbientTemp float64
	FanSpeed    uint32
	Throttled   bool
	Warnings    []string
	UpdatedAt   time.Time
}

// ThermalProfile defines thermal behavior requirements
type ThermalProfile string

const (
	// Thermal profiles
	ProfileQuiet   ThermalProfile = "QUIET"   // Prioritize noise reduction
	ProfileBalance ThermalProfile = "BALANCE" // Balance noise and cooling
	ProfileCool    ThermalProfile = "COOL"    // Prioritize cooling
	ProfileMax     ThermalProfile = "MAX"     // Maximum cooling
	
	// Default timing values
	defaultMonitorInterval = 1 * time.Second
	defaultWarningDelay   = 5 * time.Second
	defaultCriticalDelay  = 1 * time.Second
)

// ThermalPolicy defines cooling behavior and thresholds
type ThermalPolicy struct {
	policy.BasePolicy

	// Active profile
	Profile ThermalProfile

	// Temperature thresholds (Celsius)
	CPUWarning     float64
	CPUCritical    float64
	GPUWarning     float64
	GPUCritical    float64
	AmbientWarning float64
	AmbientCritical float64

	// Cooling thresholds
	FanStartTemp    float64 // Temperature to start fan
	FanMinSpeed     uint32  // Minimum fan speed (%)
	FanMaxSpeed     uint32  // Maximum fan speed (%)
	FanRampRate     float64 // Speed change per degree
	ThrottleTemp    float64 // Temperature to enable throttling
	
	// Environment
	AmbientOffset  float64      // Ambient temperature adjustment
	ThermalZones   []ThermalZone // Defined thermal zones
	
	// Callbacks
	OnWarning      func(ThermalEvent)
	OnCritical     func(ThermalEvent)
	OnStateChange  func(ThermalState)
}

// ThermalZone defines a physical area with thermal requirements
type ThermalZone struct {
	Name       string
	MaxTemp    float64
	TargetTemp float64
	Priority   int
	Sensors    []string
}

// ThermalEvent represents a thermal incident
type ThermalEvent struct {
	DeviceID    string
	Zone        string
	Type        string
	Temperature float64
	Threshold   float64
	State       ThermalState
	Timestamp   time.Time
	Details     map[string]interface{}
}

// ThermalMetrics tracks thermal performance
type ThermalMetrics struct {
	policy.Metrics
	CPUTemp        float64        `json:"cpu_temp"`
	GPUTemp        float64        `json:"gpu_temp"`
	AmbientTemp    float64        `json:"ambient_temp"`
	FanSpeed       uint32         `json:"fan_speed"`
	CurrentProfile ThermalProfile `json:"profile"`
}

// Config defines hardware monitor configuration
type Config struct {
	GPIO            *GPIOController
	MonitorInterval time.Duration
	FanControlPin   string
	ThrottlePin     string
	CPUTempPath     string
	GPUTempPath     string
	AmbientTempPath string
	OnStateChange   func(ThermalState)
}

// CoolingCurve defines fan response to temperature
type CoolingCurve struct {
	// Temperature points (°C)
	Points []float64
	
	// Fan speeds for each point (%)
	Speeds []uint32
	
	// Optional zone weights
	ZoneWeights map[string]float64
	
	// Curve characteristics
	Hysteresis  float64 // Temperature difference for speed reduction
	SmoothSteps int     // Number of steps for speed changes
	RampTime    time.Duration // Time to reach target speed
}

// GPIOController defines GPIO operations needed for thermal control
type GPIOController interface {
	ConfigurePin(name string, pin uint, mode string) error
	SetPWMDutyCycle(name string, duty uint32) error
	GetPinState(name string) (bool, error)
	SetPinState(name string, state bool) error
}