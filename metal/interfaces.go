// Package metal provides hardware abstraction and management interfaces for Raspberry Pi devices
package metal

import (
	"context"
	"time"
)

// GPIOController defines the interface for GPIO operations
type GPIOController interface {
	// ConfigurePin sets up a GPIO pin with the given mode
	ConfigurePin(name string, pin uint, mode string) error
	
	// GetPinState returns the current state of a pin
	GetPinState(name string) (bool, error)
	
	// SetPinState sets the state of a digital pin
	SetPinState(name string, state bool) error
	
	// SetPWMDutyCycle sets PWM duty cycle (0-100)
	SetPWMDutyCycle(name string, duty uint32) error
	
	// Close releases resources
	Close() error
}

// Monitor defines the common interface for hardware monitoring
type Monitor interface {
	// Start begins monitoring with the given context
	Start(ctx context.Context) error
	
	// Stop halts monitoring
	Stop() error
	
	// Close releases resources
	Close() error
}

// Option defines a functional option for configuring hardware components
type Option func(interface{}) error

// CommonState contains fields shared by all hardware states
type CommonState struct {
	DeviceID  string    `json:"device_id"`
	UpdatedAt time.Time `json:"updated_at"`
}
