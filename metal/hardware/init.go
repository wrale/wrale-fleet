// Package hardware provides hardware initialization and factory functions
package hardware

import (
	"github.com/wrale/wrale-fleet/metal/gpio"
	"github.com/wrale/wrale-fleet/metal/secure"
	"github.com/wrale/wrale-fleet/metal/thermal"
)

// NewSecureMonitor creates a new security monitor with defaults
func NewSecureMonitor(deviceID string) (*secure.Manager, error) {
	gpioCtrl, err := gpio.New()
	if err != nil {
		return nil, err
	}

	return secure.New(secure.Config{
		GPIO:          gpioCtrl,
		CaseSensor:    "case_tamper",
		MotionSensor:  "motion_detect",  
		VoltageSensor: "voltage_mon",
		DeviceID:      deviceID,
	})
}

// NewThermalMonitor creates a new thermal monitor with defaults
func NewThermalMonitor(deviceID string) (*thermal.Monitor, error) {
	gpioCtrl, err := gpio.New()
	if err != nil {
		return nil, err
	}

	return thermal.New(thermal.Config{
		GPIO:          gpioCtrl,
		FanControlPin: "fan_control",
		ThrottlePin:   "cpu_throttle",
		CPUTempPath:   "/sys/class/thermal/thermal_zone0/temp",
		GPUTempPath:   "/sys/class/thermal/thermal_zone1/temp",
	})
}