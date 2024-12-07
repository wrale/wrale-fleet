package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// Edge represents interrupt trigger edges
type Edge string

const (
	Rising  Edge = "RISING"
	Falling Edge = "FALLING"
	Both    Edge = "BOTH"

	// Default debounce time for hardware interrupts
	defaultDebounceTime = 50 * time.Millisecond
)

// InterruptHandler is called when an interrupt occurs
type InterruptHandler func(pin string, state bool)

// InterruptConfig configures interrupt behavior
type InterruptConfig struct {
	Edge         Edge
	DebounceTime time.Duration
	Handler      InterruptHandler
}

// interruptState tracks interrupt configuration and state for a pin
type interruptState struct {
	config      InterruptConfig
	lastTrigger time.Time
	enabled     bool
	channel     chan bool
}

func newInterruptState(cfg InterruptConfig) *interruptState {
	if cfg.DebounceTime == 0 {
		cfg.DebounceTime = defaultDebounceTime
	}
	return &interruptState{
		config:  cfg,
		enabled: true,
		channel: make(chan bool, 1),
	}
}

// EnableInterrupt enables interrupt detection on a pin
func (c *Controller) EnableInterrupt(name string, cfg InterruptConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	pin, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	// Configure pin for input
	pin.mode = types.ModeInput

	// Initialize interrupt tracking
	if c.interrupts == nil {
		c.interrupts = make(map[string]*interruptState)
	}

	// Store interrupt configuration
	c.interrupts[name] = newInterruptState(cfg)

	return nil
}

// DisableInterrupt disables interrupt detection on a pin
func (c *Controller) DisableInterrupt(name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	state, exists := c.interrupts[name]
	if !exists {
		return fmt.Errorf("no interrupt configured for pin %s", name)
	}

	state.enabled = false
	close(state.channel)
	delete(c.interrupts, name)
	return nil
}

// handleInterrupt processes a pin interrupt event
func (c *Controller) handleInterrupt(name string, state bool) {
	c.mux.RLock()
	interrupt, exists := c.interrupts[name]
	if !exists || !interrupt.enabled {
		c.mux.RUnlock()
		return
	}

	// Check debounce
	now := time.Now()
	if now.Sub(interrupt.lastTrigger) < interrupt.config.DebounceTime {
		c.mux.RUnlock()
		return
	}

	// Update last trigger time
	interrupt.lastTrigger = now

	// Get handler and channel for notification
	handler := interrupt.config.Handler
	ch := interrupt.channel
	c.mux.RUnlock()

	// Call handler if configured
	if handler != nil {
		handler(name, state)
	}

	// Send state to monitoring channel if active
	select {
	case ch <- state:
	default:
	}
}

// monitorPin continuously monitors a pin for state changes
func (c *Controller) monitorPin(ctx context.Context, name string) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	var lastState bool
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mux.RLock()
			pin, exists := c.pins[name]
			if !exists {
				c.mux.RUnlock()
				return
			}
			state := pin.value
			c.mux.RUnlock()

			if state != lastState {
				c.handleInterrupt(name, state)
				lastState = state
			}
		}
	}
}

// startMonitoring begins monitoring all interrupt-enabled pins
func (c *Controller) startMonitoring(ctx context.Context) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	// Start monitoring each pin with interrupts enabled
	for name, state := range c.interrupts {
		if state.enabled {
			go c.monitorPin(ctx, name)
		}
	}

	return nil
}

// WatchPin creates a channel for monitoring pin state changes
func (c *Controller) WatchPin(name string, mode types.PinMode) (<-chan bool, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	pin, exists := c.pins[name]
	if !exists {
		return nil, fmt.Errorf("pin %s not found", name)
	}

	// Configure interrupt if not already set up
	if _, hasInterrupt := c.interrupts[name]; !hasInterrupt {
		c.interrupts[name] = newInterruptState(InterruptConfig{
			Edge:         Both,
			DebounceTime: defaultDebounceTime,
		})
	}

	return c.interrupts[name].channel, nil
}