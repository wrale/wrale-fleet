// Package hw handles hardware initialization
package hw

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/gpio"
	"github.com/wrale/wrale-fleet/metal/internal/monitors"
)

// MonitorConfig holds hardware monitor configuration
type MonitorConfig struct {
	DeviceID string
}

// securityMonitor implements hardware security monitoring
type securityMonitor struct {
	gpio        monitors.GPIOController
	caseSensor  string
	motionSensor string
	voltSensor  string
	state       monitors.TamperState
}

// NewSecurityMonitor creates a security monitor instance
func NewSecurityMonitor(cfg MonitorConfig) (monitors.SecurityMonitor, error) {
	gpioCtrl, err := gpio.New()
	if err != nil {
		return nil, err
	}

	return &securityMonitor{
		gpio:        gpioCtrl,
		caseSensor:  "case_tamper",
		motionSensor: "motion_detect",
		voltSensor:  "voltage_mon",
	}, nil
}

// thermalMonitor implements hardware thermal monitoring
type thermalMonitor struct {
	gpio         monitors.GPIOController
	fanPin       string
	throttlePin  string
	cpuTemp      string
	gpuTemp      string
	ambientTemp  string
	state        monitors.ThermalState
}

// NewThermalMonitor creates a thermal monitor instance
func NewThermalMonitor(cfg MonitorConfig) (monitors.ThermalMonitor, error) {
	gpioCtrl, err := gpio.New()
	if err != nil {
		return nil, err
	}

	return &thermalMonitor{
		gpio:        gpioCtrl,
		fanPin:      "fan_control",
		throttlePin: "cpu_throttle",
		cpuTemp:     "/sys/class/thermal/thermal_zone0/temp",
		gpuTemp:     "/sys/class/thermal/thermal_zone1/temp",
		ambientTemp: "/sys/class/thermal/thermal_zone2/temp",
	}, nil
}

// Monitor starts continuous security monitoring
func (m *securityMonitor) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.checkSecurity(); err != nil {
				return err
			}
		}
	}
}

// GetTamperState returns the current security state
func (m *securityMonitor) GetTamperState() monitors.TamperState {
	return m.state
}

func (m *securityMonitor) checkSecurity() error {
	// Check case sensor
	caseOpen, err := m.gpio.GetPinState(m.caseSensor)
	if err != nil {
		return err
	}

	// Check motion sensor
	motion, err := m.gpio.GetPinState(m.motionSensor)
	if err != nil {
		return err
	}

	// Check voltage sensor
	voltageOK, err := m.gpio.GetPinState(m.voltSensor)
	if err != nil {
		return err
	}

	m.state = monitors.TamperState{
		CaseOpen:       caseOpen,
		MotionDetected: motion,
		VoltageNormal:  voltageOK,
		LastCheck:      time.Now(),
	}

	return nil
}

// Monitor starts continuous thermal monitoring
func (m *thermalMonitor) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updateThermalState(); err != nil {
				return err
			}
		}
	}
}

// GetThermalState returns current temperature state
func (m *thermalMonitor) GetThermalState() monitors.ThermalState {
	return m.state
}

// SetFanSpeed updates fan speed percentage
func (m *thermalMonitor) SetFanSpeed(speed uint32) error {
	if speed > 100 {
		speed = 100
	}
	
	if err := m.gpio.SetPWMDutyCycle(m.fanPin, speed); err != nil {
		return err
	}

	m.state.FanSpeed = speed
	return nil
}

// SetThrottling enables/disables CPU throttling
func (m *thermalMonitor) SetThrottling(enabled bool) error {
	if err := m.gpio.SetPinState(m.throttlePin, enabled); err != nil {
		return err
	}

	m.state.Throttled = enabled
	return nil
}

func (m *thermalMonitor) updateThermalState() error {
	// TODO: Read actual temperatures from files
	m.state.CPUTemp = 45.0
	m.state.GPUTemp = 40.0
	m.state.AmbientTemp = 25.0
	m.state.UpdatedAt = time.Now()
	
	return nil
}