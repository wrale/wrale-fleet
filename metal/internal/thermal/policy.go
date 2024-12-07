package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// PolicyManager handles thermal policy enforcement
type PolicyManager struct {
	sync.RWMutex

	// Core components
	hwMonitor   metal.ThermalManager
	policy      metal.ThermalPolicy
	metrics     ThermalMetrics
	deviceID    string

	// Timing state
	lastFanChange   time.Time
	lastWarning     time.Time
	lastCritical    time.Time
}

// NewPolicyManager creates a new thermal policy manager
func NewPolicyManager(deviceID string, hwMonitor metal.ThermalManager, policy metal.ThermalPolicy) *PolicyManager {
	return &PolicyManager{
		hwMonitor: hwMonitor,
		policy:    policy,
		deviceID:  deviceID,
		metrics: ThermalMetrics{
			CurrentProfile: policy.Profile,
			LastUpdate:     time.Now(),
		},
	}
}

// UpdatePolicy updates the thermal policy
func (p *PolicyManager) UpdatePolicy(policy metal.ThermalPolicy) {
	p.Lock()
	defer p.Unlock()
	p.policy = policy
	p.metrics.CurrentProfile = policy.Profile
}

// GetPolicy returns the current thermal policy
func (p *PolicyManager) GetPolicy() metal.ThermalPolicy {
	p.RLock()
	defer p.RUnlock()
	return p.policy
}

// GetMetrics returns current thermal metrics
func (p *PolicyManager) GetMetrics() ThermalMetrics {
	p.RLock()
	defer p.RUnlock()
	return p.metrics
}

// HandleStateUpdate processes a thermal state update
func (p *PolicyManager) HandleStateUpdate(state metal.ThermalState) error {
	p.Lock()
	defer p.Unlock()

	// Update metrics
	p.updateMetrics(state)

	// Check temperature thresholds
	if err := p.checkThresholds(state); err != nil {
		return fmt.Errorf("threshold check failed: %w", err)
	}

	// Determine required cooling
	if err := p.updateCooling(state); err != nil {
		return fmt.Errorf("cooling update failed: %w", err)
	}

	return nil
}

// checkThresholds verifies temperature limits
func (p *PolicyManager) checkThresholds(state metal.ThermalState) error {
	now := time.Now()

	// Check CPU temperature
	if state.CPUTemp >= p.policy.CPUCritical && time.Since(p.lastCritical) >= p.policy.AlertDelay {
		if p.policy.OnCritical != nil {
			p.policy.OnCritical(metal.ThermalEvent{
				CommonState: metal.CommonState{
					DeviceID:  p.deviceID,
					UpdatedAt: now,
				},
				Zone:        "CPU",
				Type:        "CRITICAL",
				Temperature: state.CPUTemp,
				Threshold:   p.policy.CPUCritical,
			})
		}
		p.lastCritical = now
	} else if state.CPUTemp >= p.policy.CPUWarning && time.Since(p.lastWarning) >= p.policy.WarningDelay {
		if p.policy.OnWarning != nil {
			p.policy.OnWarning(metal.ThermalEvent{
				CommonState: metal.CommonState{
					DeviceID:  p.deviceID,
					UpdatedAt: now,
				},
				Zone:        "CPU",
				Type:        "WARNING",
				Temperature: state.CPUTemp,
				Threshold:   p.policy.CPUWarning,
			})
		}
		p.lastWarning = now
	}

	return nil
}

// updateCooling adjusts cooling based on policy
func (p *PolicyManager) updateCooling(state metal.ThermalState) error {
	if time.Since(p.lastFanChange) < p.policy.MinDelay {
		return nil
	}

	// Get maximum temperature
	maxTemp := state.CPUTemp
	if state.GPUTemp > maxTemp {
		maxTemp = state.GPUTemp
	}

	// Calculate required fan speed based on profile
	var targetSpeed uint32
	switch p.policy.Profile {
	case metal.ProfileQuiet:
		targetSpeed = p.calculateQuietSpeed(maxTemp)
	case metal.ProfileCool:
		targetSpeed = p.calculateCoolSpeed(maxTemp)
	case metal.ProfileMax:
		targetSpeed = 100
	default: // ProfileBalance
		targetSpeed = p.calculateBalanceSpeed(maxTemp)
	}

	// Update fan if speed changed
	if targetSpeed != state.FanSpeed {
		if err := p.hwMonitor.SetFanSpeed(targetSpeed); err != nil {
			return fmt.Errorf("failed to set fan speed: %w", err)
		}
		p.lastFanChange = time.Now()
	}

	return nil
}

func (p *PolicyManager) calculateQuietSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return 0
	}
	speed := uint32((temp - p.policy.FanStartTemp) * 5.0)
	if speed > 60 { // Cap quiet mode at 60%
		speed = 60
	}
	return speed
}

func (p *PolicyManager) calculateCoolSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return 20 // Minimum 20% in cool mode
	}
	speed := uint32((temp - p.policy.FanStartTemp) * 10.0)
	if speed > 100 {
		speed = 100
	}
	return speed
}

func (p *PolicyManager) calculateBalanceSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return 10 // Minimum 10% in balanced mode
	}
	speed := uint32((temp - p.policy.FanStartTemp) * 7.5)
	if speed > 100 {
		speed = 100
	}
	return speed
}

// updateMetrics updates thermal performance metrics
func (p *PolicyManager) updateMetrics(state metal.ThermalState) {
	p.metrics.CPUTemp = state.CPUTemp
	p.metrics.GPUTemp = state.GPUTemp
	p.metrics.AmbientTemp = state.AmbientTemp
	p.metrics.FanSpeed = state.FanSpeed
	p.metrics.LastUpdate = state.UpdatedAt
}

// ThermalMetrics tracks thermal performance data
type ThermalMetrics struct {
	CPUTemp        float64            `json:"cpu_temp"`
	GPUTemp        float64            `json:"gpu_temp"`
	AmbientTemp    float64            `json:"ambient_temp"`
	FanSpeed       uint32             `json:"fan_speed"`
	CurrentProfile metal.ThermalProfile `json:"profile"`
	LastUpdate     time.Time          `json:"last_update"`
}
