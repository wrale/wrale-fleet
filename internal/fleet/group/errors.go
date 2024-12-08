package group

import "fmt"

// Error codes for the group package
const (
	ErrCodeInvalidGroup     = "INVALID_GROUP"
	ErrCodeGroupExists      = "GROUP_EXISTS"
	ErrCodeGroupNotFound    = "GROUP_NOT_FOUND"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
	ErrCodeCyclicDependency = "CYCLIC_DEPENDENCY"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeStoreOperation   = "STORE_OPERATION"
)

// Common error field names for consistent error annotation
const (
	FieldGroupID  = "group_id"
	FieldTenantID = "tenant_id"
)

// Error represents a group management error
type Error struct {
	Code    string
	Message string
	Op      string
	Err     error
	Fields  map[string]interface{}
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

// WithField adds a field to the error
func (e *Error) WithField(key string, value interface{}) *Error {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// WithGroupID adds the group ID field to the error for better error context
func (e *Error) WithGroupID(groupID string) *Error {
	return e.WithField(FieldGroupID, groupID)
}

// WithTenantID adds the tenant ID field to the error for better error context
func (e *Error) WithTenantID(tenantID string) *Error {
	return e.WithField(FieldTenantID, tenantID)
}

// E creates a new Error
func E(op, code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}

// Common error variables
var (
	ErrGroupExists      = E("group", ErrCodeGroupExists, "group already exists", nil)
	ErrGroupNotFound    = E("group", ErrCodeGroupNotFound, "group not found", nil)
	ErrCyclicDependency = E("group", ErrCodeCyclicDependency, "cyclic dependency detected in group hierarchy", nil)
)