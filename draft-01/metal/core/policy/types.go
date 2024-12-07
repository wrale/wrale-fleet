// Package policy defines common policy management interfaces and types
package policy

import "time"

// Manager defines the interface for policy managers
type Manager interface {
	// GetMetrics returns current monitoring metrics
	GetMetrics() interface{}

	// GetPolicy returns the current policy
	GetPolicy() interface{}

	// UpdatePolicy updates the current policy
	UpdatePolicy(policy interface{}) error

	// Start begins policy enforcement
	Start() error

	// Stop halts policy enforcement
	Stop() error
}

// Metrics provides common monitoring metrics
type Metrics struct {
	DeviceID    string    `json:"device_id"`
	UpdatedAt   time.Time `json:"updated_at"`
	Status      string    `json:"status"`
	Warnings    []string  `json:"warnings,omitempty"`
	LastWarning string    `json:"last_warning,omitempty"`
}

// BasePolicy contains common policy settings
type BasePolicy struct {
	Enabled       bool          `json:"enabled"`
	MinDelay      time.Duration `json:"min_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	WarningDelay  time.Duration `json:"warning_delay"`
	AlertDelay    time.Duration `json:"alert_delay"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// State represents a point-in-time system state
type State struct {
	DeviceID  string                 `json:"device_id"`
	Type      string                 `json:"type"`
	Timestamp time.Time             `json:"timestamp"`
	Values    map[string]interface{} `json:"values"`
}