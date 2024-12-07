// Package monitors defines hardware monitor interfaces
package monitors

import (
	"context"
	"time"
)

// SecurityMonitor defines hardware-level security monitoring
type SecurityMonitor interface {
	// Monitor starts security monitoring
	Monitor(ctx context.Context) error
	
	// GetTamperState returns current tamper detection state
	GetTamperState() TamperState
}

// ThermalMonitor defines hardware-level thermal monitoring
type ThermalMonitor interface {
	// Monitor starts thermal monitoring
	Monitor(ctx context.Context) error
	
	// GetThermalState returns current temperature state
	GetThermalState() ThermalState
	
	// SetFanSpeed updates cooling fan speed (0-100%)
	SetFanSpeed(speed uint32) error
	
	// SetThrottling enables/disables CPU throttling
	SetThrottling(enabled bool) error
}

// TamperState represents hardware security status
type TamperState struct {
	CaseOpen       bool
	MotionDetected bool
	VoltageNormal  bool
	LastCheck      time.Time
}

// ThermalState represents hardware temperature status
type ThermalState struct {
	CPUTemp     float64
	GPUTemp     float64
	AmbientTemp float64
	FanSpeed    uint32
	Throttled   bool
	UpdatedAt   time.Time
}

// GPIOController defines GPIO operations needed by monitors
type GPIOController interface {
	ConfigurePin(name string, pin uint, mode string) error
	GetPinState(name string) (bool, error)
	SetPinState(name string, state bool) error
	SetPWMDutyCycle(name string, duty uint32) error
}