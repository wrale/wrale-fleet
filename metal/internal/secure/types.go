package secure

import (
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// TamperState represents the current tamper detection status
type TamperState struct {
	metal.CommonState
	CaseOpen       bool      `json:"case_open"`
	MotionDetected bool      `json:"motion_detected"`
	VoltageNormal  bool      `json:"voltage_normal"`
	LastCheck      time.Time `json:"last_check"`
}

// TamperEvent represents a security violation event
type TamperEvent struct {
	metal.CommonState
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	State       TamperState `json:"state"`
	Details     interface{} `json:"details,omitempty"`
}

// Config defines security manager configuration
type Config struct {
	GPIO          metal.GPIO
	CaseSensor    string
	MotionSensor  string
	VoltageSensor string
	DeviceID      string
	OnTamper      func(TamperEvent)
}
