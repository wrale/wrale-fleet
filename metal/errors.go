package metal

import "fmt"

// Common error types
var (
	ErrNotInitialized = fmt.Errorf("component not initialized")
	ErrInvalidConfig  = fmt.Errorf("invalid configuration")
	ErrInvalidState   = fmt.Errorf("invalid state")
	ErrNotSupported   = fmt.Errorf("operation not supported")
	ErrSimulated      = fmt.Errorf("simulated operation")
)

// HardwareError represents a hardware-specific error
type HardwareError struct {
	Op      string // Operation that failed
	Comp    string // Component that failed
	Err     error  // Underlying error
}

func (e *HardwareError) Error() string {
	return fmt.Sprintf("%s failed on %s: %v", e.Op, e.Comp, e.Err)
}

func (e *HardwareError) Unwrap() error {
	return e.Err
}

// StateError represents a state transition error
type StateError struct {
	Op        string // Operation that failed
	Current   string // Current state
	Requested string // Requested state
	Err       error  // Underlying error
}

func (e *StateError) Error() string {
	return fmt.Sprintf("cannot %s from %s to %s: %v", e.Op, e.Current, e.Requested, e.Err)
}

func (e *StateError) Unwrap() error {
	return e.Err
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field string // Field that failed validation
	Value interface{} // Invalid value
	Err   error // Underlying error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s value %v: %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}
