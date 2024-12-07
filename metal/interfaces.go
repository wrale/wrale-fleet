// Package metal provides hardware abstraction and management interfaces for Raspberry Pi devices
package metal

import (
	"context"
	"time"
)

// PinMode defines valid GPIO pin modes
type PinMode string

const (
	ModeInput  PinMode = "INPUT"
	ModeOutput PinMode = "OUTPUT"
	ModePWM    PinMode = "PWM"
)

// GPIO defines the interface for GPIO operations
type GPIO interface {
	// Pin configuration
	ConfigurePin(name string, pin uint, mode PinMode) error
	ConfigurePWM(name string, pin uint, config *PWMConfig) error
	
	// Pin operations
	GetPinState(name string) (bool, error)
	SetPinState(name string, state bool) error
	SetPWMDutyCycle(name string, duty uint32) error
	
	// Pin monitoring
	WatchPin(name string, mode PinMode) (<-chan bool, error)
	UnwatchPin(name string) error
	
	// Resource cleanup
	Close() error
}

// PWMConfig defines PWM configuration options
type PWMConfig struct {
	Frequency  uint32    `json:"frequency"`
	DutyCycle  uint32    `json:"duty_cycle"`
	Resolution uint32    `json:"resolution"`
}

// CommonState contains fields shared by all hardware states
type CommonState struct {
	DeviceID  string    `json:"device_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Option defines a functional option for configuring hardware components
type Option func(interface{}) error

// CompareState compares hardware states ignoring UpdatedAt
func CompareState(a, b interface{}) bool {
	// TODO: Implement state comparison logic
	return true
}