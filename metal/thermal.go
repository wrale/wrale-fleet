package metal

import (
	"context"
	"time"
)

// ThermalProfile defines thermal behavior requirements
type ThermalProfile string

const (
	ProfileQuiet   ThermalProfile = "QUIET"   // Prioritize noise reduction
	ProfileBalance ThermalProfile = "BALANCE" // Balance noise and cooling
	ProfileCool    ThermalProfile = "COOL"    // Prioritize cooling
	ProfileMax     ThermalProfile = "MAX"     // Maximum cooling
)

// ThermalState represents the current thermal status
type ThermalState struct {
	CommonState
	CPUTemp      float64        `json:"cpu_temp"`      // Celsius
	GPUTemp      float64        `json:"gpu_temp"`      // Celsius
	AmbientTemp  float64        `json:"ambient_temp"`  // Celsius
	FanSpeed     uint32         `json:"fan_speed"`     // Percentage (0-100)
	Throttled    bool           `json:"throttled"`
	Warnings     []string       `json:"warnings,omitempty"`
	Profile      ThermalProfile `json:"profile"`
}

// ThermalZone defines a physical area with thermal requirements
type ThermalZone struct {
	Name       string   `json:"name"`
	MaxTemp    float64  `json:"max_temp"`
	TargetTemp float64  `json:"target_temp"`
	Priority   int      `json:"priority"`
	Sensors    []string `json:"sensors"`
}

// ThermalManager defines the interface for thermal management
type ThermalManager interface {
	Monitor

	// State Management
	GetState() (ThermalState, error)
	GetTemperatures() (cpu, gpu, ambient float64, err error)
	GetProfile() (ThermalProfile, error)
	
	// Cooling Control
	SetFanSpeed(speed uint32) error
	SetThrottling(enabled bool) error
	SetProfile(profile ThermalProfile) error
	
	// Zone Management
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

// ThermalEvent represents a thermal incident
type ThermalEvent struct {
	CommonState
	Zone        string     `json:"zone"`
	Type        string     `json:"type"`
	Temperature float64    `json:"temperature"`
	Threshold   float64    `json:"threshold"`
	Message     string     `json:"message,omitempty"`
}

// CoolingCurve defines fan response to temperature
type CoolingCurve struct {
	Points       []float64            `json:"points"`      // Temperature points (°C)
	Speeds       []uint32             `json:"speeds"`      // Fan speeds for each point (%)
	ZoneWeights  map[string]float64   `json:"zone_weights,omitempty"`
	Hysteresis   float64              `json:"hysteresis"`  // Temperature difference for speed reduction
	SmoothSteps  int                  `json:"smooth_steps"` // Number of steps for speed changes
	RampTime     time.Duration        `json:"ramp_time"`   // Time to reach target speed
}

// ThermalManagerConfig holds configuration for thermal management
type ThermalManagerConfig struct {
	GPIO            GPIOController
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

// NewThermalManager creates a new thermal manager
func NewThermalManager(config ThermalManagerConfig, opts ...Option) (ThermalManager, error) {
	return internal.NewThermalManager(config, opts...)
}
