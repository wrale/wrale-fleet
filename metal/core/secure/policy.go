package secure

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/secure"
)

// PolicyManager handles security policy enforcement
type PolicyManager struct {
	sync.RWMutex
	policy     SecurityPolicy
	stateStore StateStore
	deviceID   string

	// Track alert timing
	lastAlert time.Time
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(deviceID string, policy SecurityPolicy, store StateStore) *PolicyManager {
	return &PolicyManager{
		policy:     policy,
		stateStore: store,
		deviceID:   deviceID,
	}
}

// UpdatePolicy updates the security policy
func (p *PolicyManager) UpdatePolicy(policy SecurityPolicy) {
	p.Lock()
	defer p.Unlock()
	p.policy = policy
}

// HandleStateUpdate processes a new state update according to policy
func (p *PolicyManager) HandleStateUpdate(ctx context.Context, state hw.TamperState) error {
	p.Lock()
	defer p.Unlock()

	// Always persist state updates
	if p.stateStore != nil {
		if err := p.stateStore.SaveState(ctx, p.deviceID, state); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	// Check for tamper conditions
	if p.detectTamper(state) {
		// Ensure minimum time between alerts
		if time.Since(p.lastAlert) < p.policy.AlertDelay {
			return nil
		}

		// Create tamper event
		event := TamperEvent{
			DeviceID:    p.deviceID,
			Type:        determineTamperType(state),
			Severity:    determineSeverity(state, p.policy),
			Description: describeTamperState(state),
			State:       state,
			Timestamp:   time.Now(),
			Details:     createEventDetails(state),
		}

		// Add random delay to prevent timing attacks
		if p.policy.MaxResponseDelay > 0 {
			delay := p.policy.MinResponseDelay
			if delta := p.policy.MaxResponseDelay - p.policy.MinResponseDelay; delta > 0 {
				delay += time.Duration(rand.Int63n(int64(delta)))
			}
			time.Sleep(delay)
		}

		// Handle tamper event
		if err := p.handleTamperEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to handle tamper event: %w", err)
		}

		p.lastAlert = time.Now()
	}

	// Notify of state change
	if p.policy.OnStateChange != nil {
		p.policy.OnStateChange(state)
	}

	return nil
}

// detectTamper determines if the current state represents a tamper condition
func (p *PolicyManager) detectTamper(state hw.TamperState) bool {
	// Basic tamper detection
	if state.CaseOpen {
		return true
	}

	// Check voltage against policy threshold
	if !state.VoltageNormal && p.policy.VoltageThreshold > 0 {
		return true
	}

	// Motion detection with policy sensitivity and quiet hours
	if state.MotionDetected && p.policy.MotionSensitivity > 0 {
		now := time.Now()
		// Check if current time is in quiet hours
		for _, window := range p.policy.QuietHours {
			if isTimeInWindow(now, window) {
				return true
			}
		}
	}

	return false
}

// handleTamperEvent processes a detected tamper event
func (p *PolicyManager) handleTamperEvent(ctx context.Context, event TamperEvent) error {
	// Log event if store is available
	if p.stateStore != nil {
		if err := p.stateStore.LogEvent(ctx, p.deviceID, event.Type, event); err != nil {
			return fmt.Errorf("failed to log tamper event: %w", err)
		}
	}

	// Call tamper callback if configured
	if p.policy.OnTamperDetected != nil {
		p.policy.OnTamperDetected(event)
	}

	return nil
}

// Helper functions

func determineTamperType(state hw.TamperState) string {
	switch {
	case state.CaseOpen:
		return "case_intrusion"
	case !state.VoltageNormal:
		return "voltage_tamper"
	case state.MotionDetected:
		return "motion_detected"
	default:
		return "unknown_tamper"
	}
}

func determineSeverity(state hw.TamperState, policy SecurityPolicy) SecurityLevel {
	switch {
	case state.CaseOpen:
		return SecurityHigh
	case !state.VoltageNormal:
		return SecurityHigh
	case state.MotionDetected:
		return SecurityMedium
	default:
		return SecurityLow
	}
}

func describeTamperState(state hw.TamperState) string {
	switch {
	case state.CaseOpen:
		return "Case intrusion detected"
	case !state.VoltageNormal:
		return "Voltage tampering detected"
	case state.MotionDetected:
		return "Unexpected motion detected"
	default:
		return "Unknown tamper condition"
	}
}

func createEventDetails(state hw.TamperState) map[string]interface{} {
	return map[string]interface{}{
		"case_open":       state.CaseOpen,
		"motion_detected": state.MotionDetected,
		"voltage_normal":  state.VoltageNormal,
		"timestamp":       state.LastCheck,
	}
}

func isTimeInWindow(t time.Time, window TimeWindow) bool {
	// Handle window crossing midnight
	if window.Start.After(window.End) {
		return !t.After(window.End) || !t.Before(window.Start)
	}
	return !t.Before(window.Start) && !t.After(window.End)
}
