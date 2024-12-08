package device

import (
	"fmt"
)

// ErrorCode represents a specific type of error
type ErrorCode string

const (
	// ErrCodeNotFound indicates the requested device doesn't exist
	ErrCodeNotFound ErrorCode = "DEVICE_NOT_FOUND"
	
	// ErrCodeAlreadyExists indicates a device with the same ID already exists
	ErrCodeAlreadyExists ErrorCode = "DEVICE_ALREADY_EXISTS"
	
	// ErrCodeInvalidData indicates invalid device data
	ErrCodeInvalidData ErrorCode = "DEVICE_INVALID_DATA"
	
	// ErrCodeUnauthorized indicates lack of permission to access the device
	ErrCodeUnauthorized ErrorCode = "DEVICE_UNAUTHORIZED"
	
	// ErrCodeInternal indicates an internal server error
	ErrCodeInternal ErrorCode = "DEVICE_INTERNAL_ERROR"
)

// Error represents a domain-specific error with additional context
type Error struct {
	// Code is a machine-readable error code
	Code ErrorCode
	
	// Message is a human-readable error description
	Message string
	
	// Op is the operation being performed
	Op string
	
	// Err is the underlying error (if any)
	Err error
	
	// Fields contains additional error context
	Fields map[string]interface{}
}

// Error implements the error interface
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

// WithField adds context to the error
func (e *Error) WithField(key string, value interface{}) *Error {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// NewError creates a new domain error
func NewError(code ErrorCode, message string, op string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Fields:  make(map[string]interface{}),
	}
}

// WrapError wraps an existing error with domain context
func WrapError(err error, code ErrorCode, message string, op string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// Common error instances
var (
	ErrDeviceNotFound = NewError(ErrCodeNotFound, "device not found", "device.Get")
	ErrDeviceExists   = NewError(ErrCodeAlreadyExists, "device already exists", "device.Create")
)
