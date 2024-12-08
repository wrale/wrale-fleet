package tenant

import "fmt"

// Error codes for tenant operations
const (
	ErrCodeInvalidTenant    = "INVALID_TENANT"
	ErrCodeTenantNotFound   = "TENANT_NOT_FOUND"
	ErrCodeDuplicateTenant  = "DUPLICATE_TENANT"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
	ErrCodeQuotaExceeded    = "QUOTA_EXCEEDED"
)

// Error represents a tenant operation error
type Error struct {
	Code    string
	Message string
	Op      string
	Err     error
}

// Error returns the string representation of the error
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// E creates a new Error instance
func E(op, code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}
