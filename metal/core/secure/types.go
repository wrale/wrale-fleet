// Package secure provides security management and policy enforcement
package secure

import (
	"context"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/secure"
)

// Default timing values
const (
	defaultMinDelay   = 100 * time.Millisecond
	defaultMaxDelay   = 500 * time.Millisecond
	defaultAlertDelay = 5 * time.Minute
)

// SecurityLevel indicates the required security posture
type SecurityLevel string

const (
	SecurityLow    SecurityLevel = "LOW"
	SecurityMedium SecurityLevel = "MEDIUM"
	SecurityHigh   SecurityLevel = "HIGH"
)

// SecurityPolicy defines security requirements and responses
type SecurityPolicy struct {
	// Required security level
	Level SecurityLevel

	// Response delays to prevent timing attacks
	MinResponseDelay time.Duration
	MaxResponseDelay time.Duration

	// Tamper detection settings
	MotionSensitivity float64        // 0.0-1.0
	VoltageThreshold  float64        // Minimum acceptable voltage
	AlertDelay        time.Duration  // Minimum time between alerts
	QuietHours       []TimeWindow    // Time windows where motion is expected to be quiet
	
	// Callbacks
	OnTamperDetected func(TamperEvent)
	OnStateChange    func(hw.TamperState)
}

// TimeWindow represents a time period
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// SecurityMetrics provides monitoring statistics
type SecurityMetrics struct {
	CurrentLevel     SecurityLevel        `json:"current_level"`
	TamperState     hw.TamperState       `json:"tamper_state"`
	DetectionEvents  []TamperEvent       `json:"detection_events"`
	VoltageLevel     float64             `json:"voltage_level"`
	MotionDetected   bool                `json:"motion_detected"`
	LastTamperEvent  *TamperEvent        `json:"last_tamper_event,omitempty"`
	PolicyViolations []string            `json:"policy_violations,omitempty"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// TamperEvent represents a security violation
type TamperEvent struct {
	DeviceID    string
	Type        string
	Severity    SecurityLevel
	Description string
	State       hw.TamperState
	Timestamp   time.Time
	Details     interface{}
}

// Event represents a general security-related incident for logging
type Event struct {
	DeviceID  string      `json:"device_id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Details   interface{} `json:"details"`
}

// StateStore defines the interface for persisting security state
type StateStore interface {
	// SaveState persists the current security state
	SaveState(ctx context.Context, deviceID string, state hw.TamperState) error

	// LoadState retrieves the last known security state
	LoadState(ctx context.Context, deviceID string) (hw.TamperState, error)

	// LogEvent records a security event
	LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}