package secure

import (
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal/core/policy"
)

// PolicyManager handles security policy enforcement and high-level management
type PolicyManager struct {
	sync.RWMutex

	// Components
	monitor    *Monitor
	deviceID   string
	policy     SecurityPolicy
	metrics    SecurityMetrics
	hwMonitor  *Manager
}

// NewPolicyManager creates a new policy manager instance
func NewPolicyManager(deviceID string, hwMonitor *Manager, policy SecurityPolicy) *PolicyManager {
	return &PolicyManager{
		deviceID:  deviceID,
		hwMonitor: hwMonitor,
		policy:    policy,
		metrics: SecurityMetrics{
			Metrics: policy.Metrics{
				DeviceID:  deviceID,
				UpdatedAt: time.Now(),
				Status:    "ACTIVE",
			},
			CurrentLevel: policy.Level,
		},
	}
}

// DefaultPolicy returns a sensible default security policy
func DefaultPolicy() SecurityPolicy {
	return SecurityPolicy{
		BasePolicy: policy.BasePolicy{
			Enabled:      true,
			MinDelay:     defaultMinDelay,
			MaxDelay:     defaultMaxDelay,
			AlertDelay:   defaultAlertDelay,
			UpdatedAt:    time.Now(),
		},
		Level: SecurityMedium,
		
		// Detection settings
		MotionSensitivity: 0.7, // 70% sensitivity
		VoltageThreshold:  4.8, // Minimum 4.8V
		
		// Default quiet hours (if needed)
		QuietHours: []TimeWindow{},
	}
}

// Start begins policy enforcement
func (p *PolicyManager) Start() error {
	p.Lock()
	defer p.Unlock()

	// Start monitoring if not already running
	if p.monitor != nil {
		return fmt.Errorf("policy manager already running")
	}

	p.monitor = &Monitor{
		hwManager:     p.hwMonitor,
		policyManager: p,
		deviceID:      p.deviceID,
	}

	return nil
}

// Stop halts policy enforcement
func (p *PolicyManager) Stop() error {
	p.Lock()
	defer p.Unlock()

	if p.monitor == nil {
		return nil
	}

	p.monitor = nil
	return nil
}

// GetMetrics returns current security metrics
func (p *PolicyManager) GetMetrics() interface{} {
	p.RLock()
	defer p.RUnlock()
	return p.metrics
}

// GetPolicy returns the current security policy
func (p *PolicyManager) GetPolicy() interface{} {
	p.RLock()
	defer p.RUnlock()
	return p.policy
}

// UpdatePolicy updates the current security policy
func (p *PolicyManager) UpdatePolicy(newPolicy interface{}) error {
	policy, ok := newPolicy.(SecurityPolicy)
	if !ok {
		return fmt.Errorf("invalid policy type: expected SecurityPolicy")
	}

	p.Lock()
	defer p.Unlock()

	p.policy = policy
	p.metrics.CurrentLevel = policy.Level
	p.metrics.UpdatedAt = time.Now()

	// Update monitor if it exists
	if p.monitor != nil {
		p.monitor.policyManager = p
	}

	return nil
}

// HandleStateUpdate processes a state update from hardware
func (p *PolicyManager) HandleStateUpdate(state TamperState) error {
	p.Lock()
	defer p.Unlock()

	event := TamperEvent{
		DeviceID:  p.deviceID,
		State:     state,
		Timestamp: time.Now(),
	}

	// Process based on current state
	if state.CaseOpen {
		event.Type = "CASE_TAMPER"
		event.Severity = SecurityHigh
		p.metrics.Warnings = append(p.metrics.Warnings, "Case tamper detected")
	}

	if state.MotionDetected {
		event.Type = "MOTION_DETECTED"
		event.Severity = SecurityMedium
		p.metrics.Warnings = append(p.metrics.Warnings, "Motion detected")
	}

	if !state.VoltageNormal {
		event.Type = "VOLTAGE_TAMPER"
		event.Severity = SecurityHigh
		p.metrics.Warnings = append(p.metrics.Warnings, "Voltage tamper detected")
	}

	// Update metrics
	if event.Type != "" {
		p.metrics.DetectionEvents = append(p.metrics.DetectionEvents, event)
		if len(p.metrics.DetectionEvents) > 100 {
			p.metrics.DetectionEvents = p.metrics.DetectionEvents[1:]
		}

		if p.policy.OnTamperDetected != nil {
			p.policy.OnTamperDetected(event)
		}
	}

	// State change notification
	if p.policy.OnStateChange != nil {
		p.policy.OnStateChange(state)
	}

	return nil
}