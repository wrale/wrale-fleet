package config

import "fmt"

// ErrorCode represents a specific error condition
type ErrorCode string

const (
	ErrInvalidTemplate    ErrorCode = "invalid_template"
	ErrInvalidVersion     ErrorCode = "invalid_version"
	ErrInvalidDeployment  ErrorCode = "invalid_deployment"
	ErrTemplateNotFound   ErrorCode = "template_not_found"
	ErrVersionNotFound    ErrorCode = "version_not_found"
	ErrDeploymentNotFound ErrorCode = "deployment_not_found"
	ErrValidationFailed   ErrorCode = "validation_failed"
	ErrStoreOperation     ErrorCode = "store_operation_failed"
)

// Error represents a configuration management error
type Error struct {
	Op      string    // Operation where the error occurred
	Code    ErrorCode // Machine-readable error code
	Message string    // Human-readable error message
	Err     error     // Underlying error if any
}

// NewError creates a new configuration error
func NewError(op string, code ErrorCode, message string) *Error {
	return &Error{
		Op:      op,
		Code:    code,
		Message: message,
	}
}

// Error returns a string representation of the error
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying error if any
func (e *Error) Unwrap() error {
	return e.Err
}
