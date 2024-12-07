// Package thermal provides thermal management and policy enforcement
package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Constants for thermal management
const (
	minResponseDelay     = 100 * time.Millisecond
	defaultWarningDelay  = 5 * time.Second
	defaultCriticalDelay = 1 * time.Second
)

// ThermalState represents the current thermal status
type ThermalState struct {
	metal.CommonState
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
	GPIO            metal.GPIO
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

// CoolingCurve defines fan response to temperature
type CoolingCurve struct {
	Points      []float64          `json:"points"`
	Speeds      []uint32           `json:"speeds"`
	ZoneWeights map[string]float64 `json:"zone_weights"`
	Hysteresis  float64           `json:"hysteresis"`
	SmoothSteps int               `json:"smooth_steps"`
	RampTime    time.Duration     `json:"ramp_time"`
}