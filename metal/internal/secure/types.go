// Package secure provides security management and policy enforcement
package secure

import (
	"context"
	"time"

	"github.com/wrale/wrale-fleet/metal/core/policy"
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

	// Default timing values
	defaultMinDelay   = 100 * time.Millisecond
	defaultMaxDelay   = 500 * time.Millisecond
	defaultAlertDelay = 5 * time.Minute
)

// SecurityPolicy defines security requirements and responses
type SecurityPolicy struct {
	policy.BasePolicy

	// Required security level
	Level SecurityLevel

	// Tamper detection settings
	MotionSensitivity float64     // 0.0-1.0
	VoltageThreshold  float64     // Minimum acceptable voltage
	QuietHours       []TimeWindow // Time windows where motion is expected to be quiet
	
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
	policy.Metrics
	CurrentLevel    SecurityLevel `json:"current_level"`
	DetectionEvents []TamperEvent `json:"detection_events"`
	VoltageLevel    float64      `json:"voltage_level"`
	MotionDetected  bool         `json:"motion_detected"`
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

// StateStore defines the interface for persisting security state
type StateStore interface {
	SaveState(ctx context.Context, deviceID string, state TamperState) error
	LoadState(ctx context.Context, deviceID string) (TamperState, error)
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