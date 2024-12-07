package thermal

import (
	"fmt"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// Default timing values for thermal management
const (
	// Minimum delay between temperature readings to prevent sensor oversampling
	minResponseDelay = 100 * time.Millisecond

	// Delay before issuing warnings to filter out temporary spikes
	defaultWarningDelay = 5 * time.Second

	// Delay before taking critical action to allow for potential recovery
	defaultCriticalDelay = 1 * time.Second
)

// HardwareMonitor wraps the low-level hardware thermal monitor
type HardwareMonitor struct {
	monitor *hw.Monitor
}

// defaultConfig returns a default hardware configuration
func defaultConfig() hw.Config {
	return hw.Config{
		MonitorInterval: minResponseDelay,
		FanControlPin:  "fan_control",
		ThrottlePin:    "cpu_throttle",
		CPUTempPath:    "/sys/class/thermal/thermal_zone0/temp",
		GPUTempPath:    "/sys/class/thermal/thermal_zone1/temp",
		AmbientTempPath: "/sys/class/thermal/thermal_zone2/temp",
	}
}

// NewHardwareMonitor creates a new hardware monitor instance
func NewHardwareMonitor() (*HardwareMonitor, error) {
	monitor, err := hw.New(defaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware monitor: %w", err)
	}

	return &HardwareMonitor{
		monitor: monitor,
	}, nil
}

// DefaultPolicy returns a sensible default thermal policy
func DefaultPolicy() ThermalPolicy {
	return ThermalPolicy{
		Profile: ProfileBalance,
		
		// Temperature thresholds (in Celsius)
		CPUWarning:  70,
		CPUCritical: 85,
		GPUWarning:  75,
		GPUCritical: 90,
		
		// Fan control
		FanMinSpeed: 20,  // Minimum 20% speed
		FanMaxSpeed: 100, // Maximum 100% speed
		FanStartTemp: 45, // Start ramping up at 45°C
		FanRampRate: 2.5, // Increase by 2.5% per degree
		
		// Timing parameters
		ResponseDelay:  minResponseDelay,
		WarningDelay:  defaultWarningDelay,
		CriticalDelay: defaultCriticalDelay,
		
		// Throttling
		ThrottleTemp: 80, // Begin throttling at 80°C
	}
}

// GetPolicy returns the current thermal policy
func (p *PolicyManager) GetPolicy() ThermalPolicy {
	p.RLock()
	defer p.RUnlock()
	return p.policy
}