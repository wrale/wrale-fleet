package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Controller manages GPIO pins and their states
type Controller struct {
	mux        sync.RWMutex
	pins       map[string]*pin
	interrupts map[string]*interruptState
	enabled    bool
	simulation bool

	// Simulated state
	simPins map[string]*simPin
}

// New creates a new GPIO controller
func New(opts ...metal.Option) (metal.GPIO, error) {
	c := &Controller{
		pins:       make(map[string]*pin),
		interrupts: make(map[string]*interruptState),
		enabled:    true,
		simPins:    make(map[string]*simPin),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// ConfigurePin sets up a GPIO pin
func (c *Controller) ConfigurePin(name string, pinNum uint, mode metal.PinMode) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	p := &pin{
		name: name,
		mode: mode,
	}

	c.pins[name] = p
	return nil
}

// ConfigurePWM sets up a PWM output
func (c *Controller) ConfigurePWM(name string, pinNum uint, config *metal.PWMConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	p := &pin{
		name:      name,
		mode:      metal.ModePWM,
		pwmConfig: config,
	}

	c.pins[name] = p
	return nil
}

// SetPinState sets output pin state
func (c *Controller) SetPinState(name string, state bool) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	p, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	if p.mode != metal.ModeOutput {
		return fmt.Errorf("pin %s not configured for output", name)
	}

	p.value = state
	return nil
}

// GetPinState reads pin state
func (c *Controller) GetPinState(name string) (bool, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return false, fmt.Errorf("pin %s not found", name)
	}

	return p.value, nil
}

// SetPWMDutyCycle updates PWM duty cycle
func (c *Controller) SetPWMDutyCycle(name string, duty uint32) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	p, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	if p.mode != metal.ModePWM {
		return fmt.Errorf("pin %s not configured for PWM", name)
	}

	if p.pwmConfig == nil {
		return fmt.Errorf("pin %s PWM not configured", name)
	}

	if duty > maxDutyCycle {
		return fmt.Errorf("duty cycle must be 0-100")
	}

	p.pwmConfig.DutyCycle = duty
	return nil
}

// CreatePinGroup creates a group of pins
func (c *Controller) CreatePinGroup(name string, pins []uint) error {
	// TODO: Implement pin grouping
	return nil
}

// SetGroupState sets state for a pin group
func (c *Controller) SetGroupState(name string, states []bool) error {
	// TODO: Implement group state setting
	return nil
}

// GetGroupState gets state for a pin group
func (c *Controller) GetGroupState(name string) ([]bool, error) {
	// TODO: Implement group state getting
	return nil, nil
}

// GetPinMode gets the mode of a pin
func (c *Controller) GetPinMode(name string) (metal.PinMode, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return "", fmt.Errorf("pin %s not found", name)
	}

	return p.mode, nil
}

// GetPinConfig gets the PWM config of a pin
func (c *Controller) GetPinConfig(name string) (*metal.PWMConfig, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return nil, fmt.Errorf("pin %s not found", name)
	}

	if p.mode != metal.ModePWM {
		return nil, fmt.Errorf("pin %s not in PWM mode", name)
	}

	return p.pwmConfig, nil
}

// ListPins returns all configured pin names
func (c *Controller) ListPins() []string {
	c.mux.RLock()
	defer c.mux.RUnlock()

	pins := make([]string, 0, len(c.pins))
	for name := range c.pins {
		pins = append(pins, name)
	}
	return pins
}

// Monitor interface
func (c *Controller) GetState() interface{} {
	// Return current GPIO state
	pins := make(map[string]bool)
	c.mux.RLock()
	defer c.mux.RUnlock()
	
	for name, p := range c.pins {
		if p.mode == metal.ModeInput || p.mode == metal.ModeOutput {
			pins[name] = p.value
		}
	}
	return pins
}

func (c *Controller) WatchEvents(ctx context.Context) (<-chan interface{}, error) {
	ch := make(chan interface{}, 10)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				state := c.GetState()
				select {
				case ch <- state:
				default:
				}
			}
		}
	}()
	return ch, nil
}

// Close releases resources
func (c *Controller) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Close all interrupt channels
	for _, state := range c.interrupts {
		close(state.channel)
	}

	// Reset all pins to safe state
	for _, p := range c.pins {
		if p.mode == metal.ModeOutput {
			p.value = false
		}
	}

	c.enabled = false
	return nil
}

// Simulation control
func (c *Controller) SetSimulated(simulated bool) {
	c.simulation = simulated
}

func (c *Controller) IsSimulated() bool {
	return c.simulation
}
