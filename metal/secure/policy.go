package secure

import (
	"fmt"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/secure"
)

// PolicyManager handles security policy enforcement and high-level management
type PolicyManager struct {
	sync.RWMutex

	// Components
	monitor   *Monitor
	deviceID  string
	policy    SecurityPolicy
	metrics   SecurityMetrics
	hwMonitor *HardwareMonitor
}

// HardwareMonitor wraps the low-level security hardware monitor
type HardwareMonitor struct {
	monitor *hw.Manager
}

// NewHardwareMonitor creates a new hardware monitor instance
func NewHardwareMonitor() (*HardwareMonitor, error) {
	monitor, err := hw.New(hw.Config{
		// Configure with default settings
		CaseSensor:     "case_tamper",
		MotionSensor:   "motion_detect",
		VoltageSensor:  "voltage_mon",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create hardware monitor: %w", err)
	}

	return &HardwareMonitor{
		monitor: monitor,
	}, nil
}

// NewPolicyManager creates a new policy manager instance
func NewPolicyManager(deviceID string, hwMonitor *HardwareMonitor, policy SecurityPolicy) *PolicyManager {
	return &PolicyManager{
		deviceID:  deviceID,
		hwMonitor: hwMonitor,
		policy:    policy,
		metrics: SecurityMetrics{
			CurrentLevel: policy.Level,
			UpdatedAt:   time.Now(),
		},
	}
}

// DefaultPolicy returns a sensible default security policy
func DefaultPolicy() SecurityPolicy {
	return SecurityPolicy{
		Level: SecurityMedium,
		
		// Response timing
		MinResponseDelay: defaultMinDelay,
		MaxResponseDelay: defaultMaxDelay,
		
		// Detection settings
		MotionSensitivity: 0.7, // 70% sensitivity
		VoltageThreshold:  4.8, // Minimum 4.8V
		AlertDelay:        defaultAlertDelay,
		
		// Default quiet hours (if needed)
		QuietHours: []TimeWindow{},
	}
}

// GetMetrics returns current security metrics
func (p *PolicyManager) GetMetrics() SecurityMetrics {
	p.RLock()
	defer p.RUnlock()
	return p.metrics
}

// GetPolicy returns the current security policy
func (p *PolicyManager) GetPolicy() SecurityPolicy {
	p.RLock()
	defer p.RUnlock()
	return p.policy
}

// UpdatePolicy updates the current security policy
func (p *PolicyManager) UpdatePolicy(policy SecurityPolicy) {
	p.Lock()
	defer p.Unlock()

	p.policy = policy
	p.metrics.CurrentLevel = policy.Level
	p.metrics.UpdatedAt = time.Now()

	// Update monitor if it exists
	if p.monitor != nil {
		p.monitor.policyManager = p
	}
}