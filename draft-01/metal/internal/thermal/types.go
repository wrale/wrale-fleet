// Package thermal provides thermal management and policy enforcement 
package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// Constants for thermal management
const (
	minResponseDelay     = 100 * time.Millisecond
	defaultWarningDelay  = 5 * time.Second
	defaultCriticalDelay = 1 * time.Second
	defaultMonitorInterval = time.Second
	
	// Temperature thresholds in Celsius
	defaultWarningTemp  = 75.0
	defaultCriticalTemp = 85.0
	defaultTargetTemp   = 65.0
	
	// Fan speed limits
	minFanSpeed = 20  // Minimum speed to keep fan spinning
	maxFanSpeed = 100 // Maximum fan speed percentage
)

// ThermalState represents the current thermal status
type ThermalState struct {
	types.CommonState
	CPUTemp     float64   `json:"cpu_temp"`
	GPUTemp     float64   `json:"gpu_temp"`
	AmbientTemp float64   `json:"ambient_temp"`
	FanSpeed    uint32    `json:"fan_speed"`
	Throttled   bool      `json:"throttled"`
	Profile     Profile   `json:"profile"`
	Warnings    []Warning `json:"warnings"`
}

// Profile defines thermal management profiles
type Profile string

const (
	ProfileQuiet   Profile = "QUIET"
	ProfileBalance Profile = "BALANCE"
	ProfileCool    Profile = "COOL"
	ProfileMax     Profile = "MAX"
)

// Warning represents a thermal warning message
type Warning struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Config defines thermal manager configuration
type Config struct {
	GPIO            types.GPIOController
	MonitorInterval time.Duration
	FanControlPin   string
	ThrottlePin     string
	DefaultProfile  Profile
	CPUTempPath     string
	GPUTempPath     string
	AmbientTempPath string
	OnWarning       func(ThermalState)
	OnCritical      func(ThermalState)
}

// ThermalZone defines a physical area with thermal requirements
type ThermalZone struct {
	Name       string     `json:"name"`
	MaxTemp    float64    `json:"max_temp"`
	TargetTemp float64    `json:"target_temp"`
	Priority   int        `json:"priority"`
	Sensors    []string   `json:"sensors"`
}

// CoolingCurve defines fan response to temperature
type CoolingCurve struct {
	Points      []float64          `json:"points"`       // Temperature points (°C)
	Speeds      []uint32           `json:"speeds"`       // Fan speeds for each point (%)
	ZoneWeights map[string]float64 `json:"zone_weights"` // Weight for each thermal zone
	Hysteresis  float64           `json:"hysteresis"`   // Temperature difference for speed reduction
	SmoothSteps int               `json:"smooth_steps"` // Number of steps for speed changes
	RampTime    time.Duration     `json:"ramp_time"`    // Time to reach target speed
}

// DefaultConfig returns reasonable default configuration
func DefaultConfig() Config {
	return Config{
		MonitorInterval: defaultMonitorInterval,
		DefaultProfile:  ProfileBalance,
		CPUTempPath:    "/sys/class/thermal/thermal_zone0/temp",
		GPUTempPath:    "/sys/class/thermal/thermal_zone1/temp",
	}
}

// DefaultCoolingCurve returns a reasonable default cooling curve
func DefaultCoolingCurve() *CoolingCurve {
	return &CoolingCurve{
		Points:      []float64{40, 50, 60, 70, 80, 85},
		Speeds:      []uint32{20, 30, 50, 70, 85, 100},
		Hysteresis:  5.0,
		SmoothSteps: 5,
		RampTime:    time.Second * 2,
	}
}