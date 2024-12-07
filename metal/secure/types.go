// Package secure provides security management and policy enforcement
package secure

import (
	"context"
	"time"
)

// TamperState represents the current tamper detection status
type TamperState struct {
	CaseOpen       bool
	MotionDetected bool
	VoltageNormal  bool
	LastCheck      time.Time
}

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
	OnStateChange    func(TamperState)
}

// TimeWindow represents a time period
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// SecurityMetrics provides monitoring statistics
type SecurityMetrics struct {
	CurrentLevel     SecurityLevel     `json:"current_level"`
	TamperState     TamperState       `json:"tamper_state"`
	DetectionEvents  []TamperEvent    `json:"detection_events"`
	VoltageLevel     float64         `json:"voltage_level"`
	MotionDetected   bool           `json:"motion_detected"`
	LastTamperEvent  *TamperEvent   `json:"last_tamper_event,omitempty"`
	PolicyViolations []string       `json:"policy_violations,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// TamperEvent represents a security violation
type TamperEvent struct {
	DeviceID    string
	Type        string
	Severity    SecurityLevel
	Description string
	State       TamperState
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

// SecurityEvent represents a specific security-related occurrence
type SecurityEvent struct {
	Timestamp time.Time
	Type      string
	Source    string
	Severity  string
	State     TamperState
	Context   map[string]interface{}
}

// TamperAttempt represents a detected pattern of potentially malicious activity
type TamperAttempt struct {
	StartTime  time.Time
	EndTime    time.Time
	EventCount int
	Pattern    string
	Severity   string
}

// StateTransition records a change in the system's security state
type StateTransition struct {
	Timestamp time.Time
	FromState TamperState
	ToState   TamperState
	Trigger   string
	Context   map[string]interface{}
}

// StateStore defines the interface for persisting security state
type StateStore interface {
	// SaveState persists the current security state
	SaveState(ctx context.Context, deviceID string, state TamperState) error

	// LoadState retrieves the last known security state
	LoadState(ctx context.Context, deviceID string) (TamperState, error)

	// LogEvent records a security event
	LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}

// Config holds the configuration for the security manager
type Config struct {
	GPIO          GPIOController
	CaseSensor    string
	MotionSensor  string
	VoltageSensor string
	DeviceID      string
	StateStore    StateStore
	OnTamper      func(TamperState)
}

// GPIOController defines the interface required for GPIO operations
type GPIOController interface {
	ConfigurePin(name string, pin uint, mode string) error
	GetPinState(name string) (bool, error)
	SetPinState(name string, state bool) error
	MonitorPin(name string) (<-chan bool, error)
}