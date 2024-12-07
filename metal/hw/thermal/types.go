package thermal

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/gpio"
)

// Default monitoring interval
const defaultMonitorInterval = 1 * time.Second

// ThermalState represents current thermal conditions
type ThermalState struct {
	CPUTemp     float64   `json:"cpu_temp"`     // CPU temperature in Celsius
	GPUTemp     float64   `json:"gpu_temp"`     // GPU temperature in Celsius
	AmbientTemp float64   `json:"ambient_temp"` // Ambient temperature in Celsius
	FanSpeed    uint32    `json:"fan_speed"`    // Current fan speed percentage
	Throttled   bool      `json:"throttled"`    // Whether system is throttled
	UpdatedAt   time.Time `json:"updated_at"`   // Last update timestamp
}

// Config holds thermal monitor configuration
type Config struct {
	GPIO            *gpio.Controller `json:"gpio"`
	MonitorInterval time.Duration    `json:"monitor_interval"`
	CPUTempPath     string          `json:"cpu_temp_path"`    
	GPUTempPath     string          `json:"gpu_temp_path"`    
	AmbientTempPath string          `json:"ambient_temp_path"`
	FanControlPin   string          `json:"fan_control_pin"`  
	ThrottlePin     string          `json:"throttle_pin"`     
	
	// Simple callbacks for hardware events
	OnStateChange   func(ThermalState) `json:"-"`
}

// State is an alias for ThermalState for backward compatibility
type State = ThermalState
