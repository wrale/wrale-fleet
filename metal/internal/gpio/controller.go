package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// simPin tracks simulated pin state
type simPin struct {
	value bool
	mode  metal.PinMode
	pull  metal.PullMode
}

// Controller manages GPIO pins and their states
type Controller struct {
	mux        sync.RWMutex
	pins       map[string]*pin
	groups     map[string][]string
	interrupts map[string]chan bool
	enabled    bool
	simulation bool

	// Simulated state
	simPins map[string]*simPin
}

type pin struct {
	name      string
	mode      metal.PinMode
	pull      metal.PullMode
	pwmConfig *metal.PWMConfig
	value     bool
}

// New creates a new GPIO controller
func New(opts ...metal.Option) (metal.GPIO, error) {
	c := &Controller{
		pins:       make(map[string]*pin),
		groups:     make(map[string][]string),
		interrupts: make(map[string]chan bool),
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

func (c *Controller) Start(ctx context.Context) error {
	// Initialize hardware access if not in simulation mode
	if !c.simulation {
		// Hardware initialization would go here
	}
	return nil
}

func (c *Controller) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Close all interrupt channels
	for _, ch := range c.interrupts {
		close(ch)
	}
	c.interrupts = make(map[string]chan bool)

	// Set all pins to safe state
	for _, p := range c.pins {
		if p.mode == metal.ModeOutput {
			c.setPinValue(p, false)
		}
	}

	c.enabled = false
	return nil
}

// ConfigurePin sets up a GPIO pin
func (c *Controller) ConfigurePin(name string, pinNum uint, mode metal.PinMode) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if !c.enabled {
		return fmt.Errorf("GPIO controller is disabled")
	}

	// Create pin
	p := &pin{
		name: name,
		mode: mode,
	}

	if c.simulation {
		c.simPins[name] = &simPin{
			mode: mode,
			pull: metal.PullNone,
		}
	} else {
		// Real hardware pin setup would go here
	}

	c.pins[name] = p
	return nil
}

// ConfigurePWM sets up a PWM output
func (c *Controller) ConfigurePWM(name string, pinNum uint, config metal.PWMConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	p, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not configured", name)
	}

	if p.mode != metal.ModePWM {
		return fmt.Errorf("pin %s not in PWM mode", name)
	}

	p.pwmConfig = &config
	return nil
}

// ConfigurePull sets pin pull-up/down
func (c *Controller) ConfigurePull(name string, pinNum uint, pull metal.PullMode) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	p, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not configured", name)
	}

	p.pull = pull
	if c.simulation {
		if sim, ok := c.simPins[name]; ok {
			sim.pull = pull
		}
	}
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

	return c.setPinValue(p, state)
}

// GetPinState reads pin state
func (c *Controller) GetPinState(name string) (bool, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return false, fmt.Errorf("pin %s not found", name)
	}

	if c.simulation {
		if sim, ok := c.simPins[name]; ok {
			return sim.value, nil
		}
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

	if duty > 100 {
		return fmt.Errorf("duty cycle must be 0-100")
	}

	p.pwmConfig.DutyCycle = duty
	return nil
}

// CreatePinGroup creates a named group of pins
func (c *Controller) CreatePinGroup(name string, pins []uint) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, exists := c.groups[name]; exists {
		return fmt.Errorf("group %s already exists", name)
	}

	pinNames := make([]string, len(pins))
	for i, pinNum := range pins {
		pinName := fmt.Sprintf("%s_%d", name, i)
		if err := c.ConfigurePin(pinName, pinNum, metal.ModeOutput); err != nil {
			return err
		}
		pinNames[i] = pinName
	}

	c.groups[name] = pinNames
	return nil
}

// SetGroupState sets all pins in a group
func (c *Controller) SetGroupState(name string, states []bool) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	pins, exists := c.groups[name]
	if !exists {
		return fmt.Errorf("group %s not found", name)
	}

	if len(states) != len(pins) {
		return fmt.Errorf("state count mismatch")
	}

	for i, pinName := range pins {
		if err := c.SetPinState(pinName, states[i]); err != nil {
			return err
		}
	}

	return nil
}

// GetGroupState reads all pins in a group
func (c *Controller) GetGroupState(name string) ([]bool, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	pins, exists := c.groups[name]
	if !exists {
		return nil, fmt.Errorf("group %s not found", name)
	}

	states := make([]bool, len(pins))
	for i, pinName := range pins {
		state, err := c.GetPinState(pinName)
		if err != nil {
			return nil, err
		}
		states[i] = state
	}

	return states, nil
}

// WatchPin sets up pin change monitoring
func (c *Controller) WatchPin(name string, edge string) (<-chan bool, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, exists := c.pins[name]; !exists {
		return nil, fmt.Errorf("pin %s not found", name)
	}

	ch := make(chan bool, 1)
	c.interrupts[name] = ch
	return ch, nil
}

// UnwatchPin stops monitoring a pin
func (c *Controller) UnwatchPin(name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	ch, exists := c.interrupts[name]
	if !exists {
		return nil // Already not watching
	}

	close(ch)
	delete(c.interrupts, name)
	return nil
}

// GetPinMode returns pin mode
func (c *Controller) GetPinMode(name string) (metal.PinMode, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return "", fmt.Errorf("pin %s not found", name)
	}

	return p.mode, nil
}

// GetPinConfig returns PWM configuration
func (c *Controller) GetPinConfig(name string) (metal.PWMConfig, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	p, exists := c.pins[name]
	if !exists {
		return metal.PWMConfig{}, fmt.Errorf("pin %s not found", name)
	}

	if p.pwmConfig == nil {
		return metal.PWMConfig{}, fmt.Errorf("pin %s not configured for PWM", name)
	}

	return *p.pwmConfig, nil
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

// Simulation control
func (c *Controller) SetSimulated(simulated bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.simulation = simulated
}

func (c *Controller) IsSimulated() bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.simulation
}

// Internal helpers

func (c *Controller) setPinValue(p *pin, value bool) error {
	p.value = value
	if c.simulation {
		if sim, ok := c.simPins[p.name]; ok {
			sim.value = value
		}
		return nil
	}
	// Real hardware pin control would go here
	return nil
}
