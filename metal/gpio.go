// Package metal provides hardware abstraction and management interfaces
package metal

import "time"

// PinMode defines valid GPIO pin modes
type PinMode string

const (
	ModeInput  PinMode = "INPUT"
	ModeOutput PinMode = "OUTPUT"
	ModePWM    PinMode = "PWM"
)

// PullMode defines pin pull-up/down modes
type PullMode string

const (
	PullNone PullMode = "NONE"
	PullUp   PullMode = "UP"
	PullDown PullMode = "DOWN"
)

// PWMConfig holds PWM pin configuration
type PWMConfig struct {
	Frequency uint32   // PWM frequency in Hz
	DutyCycle uint32   // Initial duty cycle (0-100)
	Pull      PullMode // Pull-up/down configuration
}

// GPIO defines the complete GPIO management interface
type GPIO interface {
	GPIOController
	Monitor

	// Pin Configuration
	ConfigurePWM(name string, pin uint, config PWMConfig) error
	ConfigurePull(name string, pin uint, pull PullMode) error
	
	// Pin Groups
	CreatePinGroup(name string, pins []uint) error
	SetGroupState(name string, states []bool) error
	GetGroupState(name string) ([]bool, error)
	
	// Interrupts
	WatchPin(name string, edge string) (<-chan bool, error)
	UnwatchPin(name string) error
	
	// Pin Info
	GetPinMode(name string) (PinMode, error)
	GetPinConfig(name string) (PWMConfig, error)
	ListPins() []string
	
	// Simulation
	SetSimulated(simulated bool)
	IsSimulated() bool
}

// GPIOEvent represents a GPIO state change
type GPIOEvent struct {
	CommonState
	Pin       string    `json:"pin"`
	Value     bool      `json:"value"`
	Mode      PinMode   `json:"mode"`
	Timestamp time.Time `json:"timestamp"`
}
