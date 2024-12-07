package thermal

import (
	"fmt"
	"time"
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

// HardwareMonitor handles low-level thermal monitoring
type HardwareMonitor struct {
	monitor *Monitor
}

// defaultConfig returns a default hardware configuration
func defaultConfig() Config {
	return Config{
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
	monitor, err := NewMonitor(defaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware monitor: %w", err)
	}

	return &HardwareMonitor{
		monitor: monitor,
	}, nil
}

// Monitor returns the underlying hardware monitor
func (h *HardwareMonitor) Monitor() *Monitor {
	return h.monitor
}

// NewMonitor creates a new thermal monitor
func NewMonitor(cfg Config) (*Monitor, error) {
	return &Monitor{
		config:     cfg,
		fanControl: cfg.FanControlPin,
		throttle:   cfg.ThrottlePin,
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