// Package thermal provides thermal management and policy enforcement
package thermal

import (
	"time"
)

// ThermalState represents the current thermal status
type ThermalState struct {
	CPUTemp      float64
	GPUTemp      float64
	AmbientTemp  float64
	FanSpeed     uint32
	Throttled    bool
	Warnings     []string
	UpdatedAt    time.Time
}

// Monitor handles hardware temperature monitoring and fan control
type Monitor struct {
	state      ThermalState
	config     Config
	fanControl string
	throttle   string
}

// Config defines the monitor configuration
type Config struct {
	MonitorInterval time.Duration
	FanControlPin   string
	ThrottlePin     string
	CPUTempPath     string
	GPUTempPath     string
	AmbientTempPath string
}

// ThermalProfile defines thermal behavior requirements
type ThermalProfile string

const (
	// Thermal profiles
	ProfileQuiet   ThermalProfile = "QUIET"   // Prioritize noise reduction
	ProfileBalance ThermalProfile = "BALANCE" // Balance noise and cooling
	ProfileCool    ThermalProfile = "COOL"    // Prioritize cooling
	ProfileMax     ThermalProfile = "MAX"     // Maximum cooling
	
	// Default monitoring intervals
	defaultStateInterval = 1 * time.Second
	defaultStatsInterval = 5 * time.Second
)

// ThermalPolicy defines cooling behavior and thresholds
type ThermalPolicy struct {
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
	
	// Timing parameters
	ResponseDelay  time.Duration // Minimum time between fan changes
	WarningDelay   time.Duration // Minimum time between warnings
	CriticalDelay  time.Duration // Minimum time between critical alerts
	
	// Environment
	AmbientOffset  float64       // Ambient temperature adjustment
	ThermalZones   []ThermalZone // Defined thermal zones
	
	// Callbacks
	OnWarning      func(ThermalEvent)
	OnCritical     func(ThermalEvent)
	OnStateChange  func(ThermalState)
}

// ThermalZone defines a physical area with thermal requirements
type ThermalZone struct {
	Name           string
	MaxTemp        float64
	TargetTemp     float64
	Priority       int
	Sensors        []string
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
	// Current temperatures
	CPUTemp        float64
	GPUTemp        float64
	AmbientTemp    float64
	
	// Fan metrics
	FanSpeed       uint32
	FanDuty        uint32
	FanRPM         uint32
	
	// Throttling
	ThrottleCount  uint64
	ThrottleTime   time.Duration
	LastThrottle   time.Time
	
	// Temperature trends
	CPUTrend       float64
	GPUTrend       float64
	AmbientTrend   float64
	
	// Status
	CurrentProfile ThermalProfile
	ActiveWarnings []string
	LastUpdate     time.Time
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