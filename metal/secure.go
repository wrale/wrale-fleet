package metal

import (
	"context"
	"time"
)

// SecurityLevel indicates the required security posture
type SecurityLevel string

const (
	SecurityLow    SecurityLevel = "LOW"
	SecurityMedium SecurityLevel = "MEDIUM"
	SecurityHigh   SecurityLevel = "HIGH"
)

// TamperState represents the current tamper detection status
type TamperState struct {
	CommonState
	CaseOpen       bool          `json:"case_open"`
	MotionDetected bool          `json:"motion_detected"`
	VoltageNormal  bool          `json:"voltage_normal"`
	SecurityLevel  SecurityLevel `json:"security_level"`
	Violations     []string      `json:"violations,omitempty"`
}

// TimeWindow represents a time period
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SecurityManager defines the interface for security management
type SecurityManager interface {
	Monitor

	// State Management
	GetState() (TamperState, error)
	GetSecurityLevel() (SecurityLevel, error)
	ValidateState() error
	
	// Security Control
	SetSecurityLevel(level SecurityLevel) error
	ClearViolations() error
	ResetTamperState() error
	
	// Monitoring
	WatchState(ctx context.Context) (<-chan TamperState, error)
	WatchSensor(ctx context.Context, name string) (<-chan bool, error)
	
	// Policy Management
	SetQuietHours(windows []TimeWindow) error
	SetMotionSensitivity(sensitivity float64) error
	SetVoltageThreshold(min float64) error
	
	// Events
	OnTamper(func(TamperEvent))
	OnViolation(func(TamperEvent))
}

// TamperEvent represents a security violation
type TamperEvent struct {
	CommonState
	Type        string        `json:"type"`
	Severity    SecurityLevel `json:"severity"`
	Description string        `json:"description"`
	State       TamperState   `json:"state"`
	Details     interface{}   `json:"details,omitempty"`
}

// StateStore defines the interface for persisting security state
type StateStore interface {
	SaveState(ctx context.Context, deviceID string, state TamperState) error
	LoadState(ctx context.Context, deviceID string) (TamperState, error)
	LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error
}

// SecurityManagerConfig holds configuration for security management
type SecurityManagerConfig struct {
	GPIO            GPIOController
	StateStore      StateStore
	CaseSensor      string
	MotionSensor    string
	VoltageSensor   string
	DefaultLevel    SecurityLevel
	QuietHours      []TimeWindow
	VoltageMin      float64
	Sensitivity     float64
	OnTamper        func(TamperEvent)
	OnViolation     func(TamperEvent)
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityManagerConfig, opts ...Option) (SecurityManager, error) {
	return internal.NewSecurityManager(config, opts...)
}
