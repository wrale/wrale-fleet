// TODO: Need to merge functionality with core's thermal.go

package thermal

import (
	"fmt"

	"github.com/wrale/wrale-fleet/metal/gpio"
)

// Fan speed ranges
const (
	fanPWMFrequency = 25000 // 25kHz standard for PC fans
)

// InitializeFanControl sets up PWM for fan control
func (m *Monitor) InitializeFanControl() error {
	if m.fanPin == "" {
		return nil // No fan control configured
	}

	err := m.gpio.ConfigurePWM(m.fanPin, nil, gpio.PWMConfig{
		Frequency: fanPWMFrequency,
		DutyCycle: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to configure fan PWM: %w", err)
	}

	// Enable PWM output
	if err := m.gpio.EnablePWM(m.fanPin); err != nil {
		return err
	}

	return nil
}

// SetFanSpeed sets the fan speed percentage (0-100)
func (m *Monitor) SetFanSpeed(speed uint32) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.fanPin == "" {
		return nil
	}

	// Clamp speed to valid range
	if speed > 100 {
		speed = 100
	}

	if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
		return fmt.Errorf("failed to set fan PWM: %w", err)
	}
	m.state.FanSpeed = speed
	return nil
}

// SetThrottling controls the throttling GPIO pin
func (m *Monitor) SetThrottling(enabled bool) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.throttlePin == "" {
		return nil
	}

	if err := m.gpio.SetPinState(m.throttlePin, enabled); err != nil {
		return fmt.Errorf("failed to set throttling state: %w", err)
	}
	m.state.Throttled = enabled
	return nil
}

// Close releases fan control resources
func (m *Monitor) Close() error {
	if m.fanPin != "" {
		if err := m.gpio.DisablePWM(m.fanPin); err != nil {
			return fmt.Errorf("failed to disable fan PWM: %w", err)
		}
	}
	return nil
}
