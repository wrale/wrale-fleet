package gpio

import (
	"fmt"
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// validatePWMConfig checks PWM configuration values
func validatePWMConfig(cfg *types.PWMConfig) error {
	if cfg.Frequency > maxFrequency {
		return fmt.Errorf("frequency %d exceeds maximum %d Hz", cfg.Frequency, maxFrequency)
	}
	if cfg.DutyCycle > maxDutyCycle {
		return fmt.Errorf("duty cycle %d exceeds maximum %d%%", cfg.DutyCycle, maxDutyCycle)
	}
	if cfg.Resolution > maxResolution {
		return fmt.Errorf("resolution %d exceeds maximum %d bits", cfg.Resolution, maxResolution)
	}
	return nil
}

// configurePWM sets up PWM on a pin
func (c *Controller) configurePWM(name string, config *types.PWMConfig) error {
	if err := validatePWMConfig(config); err != nil {
		return err
	}

	// For simulated mode, just store the configuration
	if c.simulation {
		if sim, exists := c.simPins[name]; exists {
			sim.pwmConfig = config
			return nil
		}
		return fmt.Errorf("simulated pin %s not found", name)
	}

	// TODO: Implement hardware PWM configuration
	return nil
}

// updatePWM updates PWM settings on a pin
func (c *Controller) updatePWM(p *pin, duty uint32) error {
	if p.pwmConfig == nil {
		return fmt.Errorf("PWM not configured")
	}

	if duty > maxDutyCycle {
		return fmt.Errorf("duty cycle must be 0-100")
	}

	p.pwmConfig.DutyCycle = duty

	// For simulated mode, just update the configuration
	if c.simulation {
		return nil
	}

	// TODO: Implement hardware PWM update
	return nil
}

// disablePWM turns off PWM on a pin
func (c *Controller) disablePWM(name string) error {
	p, exists := c.pins[name]
	if !exists {
		return fmt.Errorf("pin %s not found", name)
	}

	if p.mode != types.ModePWM {
		return fmt.Errorf("pin %s not in PWM mode", name)
	}

	p.mode = types.ModeOutput
	p.pwmConfig = nil

	// For simulated mode, just update the configuration
	if c.simulation {
		if sim, exists := c.simPins[name]; exists {
			sim.pwmConfig = nil
			return nil
		}
	}

	// TODO: Implement hardware PWM disable
	return nil
}