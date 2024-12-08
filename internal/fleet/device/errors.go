package device

// Error represents a device management error
type Error struct {
	Code    string
	Message string
	Op      string
	Err     error
	Fields  map[string]interface{}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Op + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Message
}

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

// Common error codes
const (
	ErrCodeInvalidDevice    = "INVALID_DEVICE"
	ErrCodeDeviceNotFound   = "DEVICE_NOT_FOUND"
	ErrCodeDeviceExists     = "DEVICE_EXISTS"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
	ErrCodeStorageError     = "STORAGE_ERROR"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
)

// E creates a new Error
func E(op string, code string, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}