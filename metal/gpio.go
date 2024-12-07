// Package metal provides hardware abstraction and management interfaces
package metal

import (
	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// GPIO defines the complete GPIO management interface
type GPIO interface {
	types.GPIOController
	types.Monitor

	// Pin Groups
	CreatePinGroup(name string, pins []uint) error
	SetGroupState(name string, states []bool) error
	GetGroupState(name string) ([]bool, error)
	
	// Pin Info
	GetPinMode(name string) (types.PinMode, error)
	GetPinConfig(name string) (*types.PWMConfig, error)
	ListPins() []string
	
	// Simulation
	SetSimulated(simulated bool)
	IsSimulated() bool
}