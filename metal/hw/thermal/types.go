package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet-metal-hw/gpio"
)

// Default monitoring interval
const defaultMonitorInterval = 1 * time.Second

// ThermalState represents current thermal conditions
type ThermalState struct {
	CPUTemp     float64   // CPU temperature in Celsius
	GPUTemp     float64   // GPU temperature in Celsius
	AmbientTemp float64   // Ambient temperature in Celsius
	FanSpeed    uint32    // Current fan speed percentage
	Throttled   bool      // Whether system is throttled
	UpdatedAt   time.Time // Last update timestamp
}

// Config holds thermal monitor configuration
type Config struct {
	GPIO            *gpio.Controller
	MonitorInterval time.Duration
	CPUTempPath     string    // sysfs path to CPU temperature
	GPUTempPath     string    // sysfs path to GPU temperature
	AmbientTempPath string    // sysfs path to ambient temperature sensor
	FanControlPin   string    // GPIO pin for fan control
	ThrottlePin     string    // GPIO pin for throttling control
	
	// Simple callbacks for hardware events
	OnStateChange   func(ThermalState)
}