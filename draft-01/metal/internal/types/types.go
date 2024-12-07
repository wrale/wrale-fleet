package types

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

// PullMode defines pin pull-up/down modes
type PullMode string

const (
	PullNone PullMode = "NONE"
	PullUp   PullMode = "UP"
	PullDown PullMode = "DOWN"
)

// PWMConfig holds PWM pin configuration
type PWMConfig struct {
	Frequency  uint32   `json:"frequency"`  // PWM frequency in Hz
	DutyCycle  uint32   `json:"duty_cycle"` // Duty cycle (0-100)
	Pull      PullMode `json:"pull"`       // Pull-up/down configuration
	Resolution uint32   `json:"resolution"` // PWM resolution in bits
}

// CommonState contains fields shared by all hardware states
type CommonState struct {
	DeviceID  string    `json:"device_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GPIOEvent represents a GPIO state change
type GPIOEvent struct {
	CommonState
	Pin       string    `json:"pin"`
	Value     bool      `json:"value"`
	Mode      PinMode   `json:"mode"`
	Timestamp time.Time `json:"timestamp"`
}

// Monitor defines common monitoring capabilities
type Monitor interface {
	// State monitoring
	GetState() interface{}
	
	// Event streaming
	WatchEvents(ctx context.Context) (<-chan interface{}, error)
	
	// Resource cleanup
	Close() error
}

// GPIOController defines basic GPIO operations
type GPIOController interface {
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

// FileStore defines common file operations
type FileStore interface {
	SaveState(deviceID string, state interface{}) error
	LoadState(deviceID string) (interface{}, error)
}