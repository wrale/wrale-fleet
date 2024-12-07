package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// PolicyManager handles thermal policy enforcement
type PolicyManager struct {
	sync.RWMutex

	// Core components
	hwMonitor   *hw.Monitor
	policy      ThermalPolicy
	metrics     ThermalMetrics
	deviceID    string

	// Cooling curve
	curve *CoolingCurve

	// Timing state
	lastFanChange   time.Time
	lastWarning     time.Time
	lastCritical    time.Time
}

// NewPolicyManager creates a new thermal policy manager
func NewPolicyManager(deviceID string, hwMonitor *hw.Monitor, policy ThermalPolicy) *PolicyManager {
	return &PolicyManager{
		hwMonitor: hwMonitor,
		policy:    policy,
		deviceID:  deviceID,
		metrics: ThermalMetrics{
			CurrentProfile: policy.Profile,
			LastUpdate:     time.Now(),
		},
		curve: calculateCoolingCurve(policy),
	}
}

// UpdatePolicy updates the thermal policy
func (p *PolicyManager) UpdatePolicy(policy ThermalPolicy) {
	p.Lock()
	defer p.Unlock()

	p.policy = policy
	p.curve = calculateCoolingCurve(policy)
	p.metrics.CurrentProfile = policy.Profile
}

// HandleStateUpdate processes a thermal state update
func (p *PolicyManager) HandleStateUpdate(ctx context.Context, state hw.ThermalState) error {
	p.Lock()
	defer p.Unlock()

	// Update metrics
	p.updateMetrics(state)

	// Check temperature thresholds
	if err := p.checkThresholds(ctx, state); err != nil {
		return fmt.Errorf("threshold check failed: %w", err)
	}

	// Determine required cooling
	if err := p.updateCooling(state); err != nil {
		return fmt.Errorf("cooling update failed: %w", err)
	}

	// Notify of state change if configured
	if p.policy.OnStateChange != nil {
		p.policy.OnStateChange(state)
	}

	return nil
}

// checkThresholds verifies temperature limits
func (p *PolicyManager) checkThresholds(ctx context.Context, state hw.ThermalState) error {
	now := time.Now()

	// Check CPU temperature
	if state.CPUTemp >= p.policy.CPUCritical {
		if now.Sub(p.lastCritical) >= p.policy.CriticalDelay {
			event := ThermalEvent{
				DeviceID:    p.deviceID,
				Zone:        "CPU",
				Type:        "CRITICAL_TEMPERATURE",
				Temperature: state.CPUTemp,
				Threshold:   p.policy.CPUCritical,
				State:      state,
				Timestamp:   now,
				Details: map[string]interface{}{
					"profile": p.policy.Profile,
					"fan_speed": state.FanSpeed,
				},
			}
			if p.policy.OnCritical != nil {
				p.policy.OnCritical(event)
			}
			p.lastCritical = now
		}
	} else if state.CPUTemp >= p.policy.CPUWarning {
		if now.Sub(p.lastWarning) >= p.policy.WarningDelay {
			event := ThermalEvent{
				DeviceID:    p.deviceID,
				Zone:        "CPU",
				Type:        "WARNING_TEMPERATURE",
				Temperature: state.CPUTemp,
				Threshold:   p.policy.CPUWarning,
				State:      state,
				Timestamp:   now,
			}
			if p.policy.OnWarning != nil {
				p.policy.OnWarning(event)
			}
			p.lastWarning = now
		}
	}

	// Similar checks for GPU
	if state.GPUTemp >= p.policy.GPUCritical {
		if now.Sub(p.lastCritical) >= p.policy.CriticalDelay {
			event := ThermalEvent{
				DeviceID:    p.deviceID,
				Zone:        "GPU",
				Type:        "CRITICAL_TEMPERATURE",
				Temperature: state.GPUTemp,
				Threshold:   p.policy.GPUCritical,
				State:      state,
				Timestamp:   now,
			}
			if p.policy.OnCritical != nil {
				p.policy.OnCritical(event)
			}
			p.lastCritical = now
		}
	} else if state.GPUTemp >= p.policy.GPUWarning {
		if now.Sub(p.lastWarning) >= p.policy.WarningDelay {
			event := ThermalEvent{
				DeviceID:    p.deviceID,
				Zone:        "GPU",
				Type:        "WARNING_TEMPERATURE",
				Temperature: state.GPUTemp,
				Threshold:   p.policy.GPUWarning,
				State:      state,
				Timestamp:   now,
			}
			if p.policy.OnWarning != nil {
				p.policy.OnWarning(event)
			}
			p.lastWarning = now
		}
	}

	return nil
}

// updateCooling adjusts cooling based on policy
func (p *PolicyManager) updateCooling(state hw.ThermalState) error {
	now := time.Now()

	// Check if we can change fan speed yet
	if now.Sub(p.lastFanChange) < p.policy.ResponseDelay {
		return nil
	}

	// Get maximum temperature
	maxTemp := state.CPUTemp
	if state.GPUTemp > maxTemp {
		maxTemp = state.GPUTemp
	}

	// Calculate required fan speed
	var targetSpeed uint32
	switch p.policy.Profile {
	case ProfileQuiet:
		targetSpeed = p.calculateQuietSpeed(maxTemp)
	case ProfileCool:
		targetSpeed = p.calculateCoolSpeed(maxTemp)
	case ProfileMax:
		targetSpeed = p.policy.FanMaxSpeed
	default: // ProfileBalance
		targetSpeed = p.calculateBalancedSpeed(maxTemp)
	}

	// Apply fan speed limits
	if targetSpeed < p.policy.FanMinSpeed {
		targetSpeed = p.policy.FanMinSpeed
	}
	if targetSpeed > p.policy.FanMaxSpeed {
		targetSpeed = p.policy.FanMaxSpeed
	}

	// Update fan if speed changed
	if targetSpeed != state.FanSpeed {
		if err := p.hwMonitor.SetFanSpeed(targetSpeed); err != nil {
			return fmt.Errorf("failed to set fan speed: %w", err)
		}
		p.lastFanChange = now
	}

	// Handle throttling
	if maxTemp >= p.policy.ThrottleTemp && !state.Throttled {
		if err := p.hwMonitor.SetThrottling(true); err != nil {
			return fmt.Errorf("failed to enable throttling: %w", err)
		}
		p.metrics.ThrottleCount++
		p.metrics.LastThrottle = now
	} else if maxTemp < p.policy.ThrottleTemp && state.Throttled {
		if err := p.hwMonitor.SetThrottling(false); err != nil {
			return fmt.Errorf("failed to disable throttling: %w", err)
		}
	}

	return nil
}

// Helper functions for different cooling profiles

func (p *PolicyManager) calculateQuietSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return 0
	}
	// Gentle fan curve for quiet operation
	delta := temp - p.policy.FanStartTemp
	speed := uint32(float64(p.policy.FanMinSpeed) + (delta * p.policy.FanRampRate * 0.5))
	return speed
}

func (p *PolicyManager) calculateCoolSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return p.policy.FanMinSpeed
	}
	// Aggressive fan curve for cooling
	delta := temp - p.policy.FanStartTemp
	speed := uint32(float64(p.policy.FanMinSpeed) + (delta * p.policy.FanRampRate * 2.0))
	return speed
}

func (p *PolicyManager) calculateBalancedSpeed(temp float64) uint32 {
	if temp < p.policy.FanStartTemp {
		return p.policy.FanMinSpeed
	}
	// Standard fan curve
	delta := temp - p.policy.FanStartTemp
	speed := uint32(float64(p.policy.FanMinSpeed) + (delta * p.policy.FanRampRate))
	return speed
}

// updateMetrics updates thermal performance metrics
func (p *PolicyManager) updateMetrics(state hw.ThermalState) {
	p.metrics.CPUTemp = state.CPUTemp
	p.metrics.GPUTemp = state.GPUTemp
	p.metrics.AmbientTemp = state.AmbientTemp
	p.metrics.FanSpeed = state.FanSpeed
	p.metrics.LastUpdate = state.UpdatedAt
}

// GetMetrics returns current thermal metrics
func (p *PolicyManager) GetMetrics() ThermalMetrics {
	p.RLock()
	defer p.RUnlock()
	return p.metrics
}

// calculateCoolingCurve creates a cooling curve from policy
func calculateCoolingCurve(policy ThermalPolicy) *CoolingCurve {
	// Create standard cooling curve points
	points := []float64{
		policy.FanStartTemp,
		policy.CPUWarning,
		policy.CPUCritical,
	}

	// Calculate speeds for each point based on profile
	speeds := make([]uint32, len(points))
	switch policy.Profile {
	case ProfileQuiet:
		speeds[0] = policy.FanMinSpeed
		speeds[1] = (policy.FanMaxSpeed + policy.FanMinSpeed) / 2
		speeds[2] = policy.FanMaxSpeed
	case ProfileCool:
		speeds[0] = policy.FanMinSpeed + ((policy.FanMaxSpeed - policy.FanMinSpeed) / 4)
		speeds[1] = policy.FanMinSpeed + ((policy.FanMaxSpeed - policy.FanMinSpeed) * 3 / 4)
		speeds[2] = policy.FanMaxSpeed
	case ProfileMax:
		speeds[0] = policy.FanMinSpeed + ((policy.FanMaxSpeed - policy.FanMinSpeed) / 2)
		speeds[1] = policy.FanMaxSpeed
		speeds[2] = policy.FanMaxSpeed
	default: // ProfileBalance
		speeds[0] = policy.FanMinSpeed
		speeds[1] = policy.FanMinSpeed + ((policy.FanMaxSpeed - policy.FanMinSpeed) * 2 / 3)
		speeds[2] = policy.FanMaxSpeed
	}

	return &CoolingCurve{
		Points:      points,
		Speeds:      speeds,
		Hysteresis:  2.0,  // 2°C hysteresis
		SmoothSteps: 5,    // 5 steps for speed changes
		RampTime:    time.Second, // 1 second ramp time
	}
}