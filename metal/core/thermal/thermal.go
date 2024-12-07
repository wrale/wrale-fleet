package thermal

import (
	"fmt"

	hw "github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// HardwareMonitor wraps the low-level hardware thermal monitor
type HardwareMonitor struct {
	monitor *hw.Monitor
}

// NewHardwareMonitor creates a new hardware monitor instance
func NewHardwareMonitor() (*HardwareMonitor, error) {
	monitor, err := hw.New(hw.DefaultConfig())
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
