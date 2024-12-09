package logging

import "errors"

// Common error definitions for the logging domain
var (
	// ErrMissingTenant indicates a tenant ID was not provided
	ErrMissingTenant = errors.New("tenant ID is required")

	// ErrMissingMessage indicates an event message was not provided
	ErrMissingMessage = errors.New("event message is required")

	// ErrInvalidLevel indicates an invalid log level was specified
	ErrInvalidLevel = errors.New("invalid log level")

	// ErrInvalidRetention indicates an invalid retention policy was specified
	ErrInvalidRetention = errors.New("invalid retention policy")

	// ErrStoreNotInitialized indicates the store was not properly initialized
	ErrStoreNotInitialized = errors.New("store not initialized")

	// ErrEventNotFound indicates a requested event does not exist
	ErrEventNotFound = errors.New("event not found")
)

// DomainError represents a domain-specific error with context
type DomainError struct {
	Op      string // Operation that failed
	Code    string // Error code for categorization
	Message string // Human-readable error message
	Err     error  // Underlying error if any
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Op + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Message
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// E creates a new DomainError
func E(op, code, message string, err error) error {
	return &DomainError{
		Op:      op,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes for consistent error handling
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeStoreFailure     = "STORE_FAILURE"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeValidation       = "VALIDATION_FAILED"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
)
