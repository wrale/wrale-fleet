package gpio

import (
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// simPin represents a simulated GPIO pin for testing
type simPin struct {
	mu        sync.RWMutex
	value     bool
	mode      types.PinMode
	pwmConfig *types.PWMConfig
	interrupt chan bool
}

func newSimPin() *simPin {
	return &simPin{
		interrupt: make(chan bool, 1),
	}
}

// pin represents a physical GPIO pin
type pin struct {
	name      string
	mode      types.PinMode
	pwmConfig *types.PWMConfig
	value     bool
}

// Constants for pin configuration
const (
	defaultFrequency    = 1000  // Default PWM frequency in Hz
	defaultDutyCycle    = 0     // Default duty cycle (0-100)
	defaultResolution   = 8     // Default PWM resolution in bits
	maxFrequency       = 50000 // Maximum PWM frequency in Hz
	maxDutyCycle       = 100   // Maximum duty cycle
	maxResolution      = 16    // Maximum PWM resolution in bits
)

// defaultPWMConfig returns standard PWM configuration
func defaultPWMConfig() *types.PWMConfig {
	return &types.PWMConfig{
		Frequency:  defaultFrequency,
		DutyCycle:  defaultDutyCycle,
		Pull:      types.PullNone,
		Resolution: defaultResolution,
	}
}