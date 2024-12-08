package device

// Error represents a device management error
type Error struct {
	Code    string
	Message string
	Op      string
	Err     error
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

// Common error codes
const (
	ErrCodeInvalidDevice    = "INVALID_DEVICE"
	ErrCodeDeviceNotFound   = "DEVICE_NOT_FOUND"
	ErrCodeDeviceExists     = "DEVICE_EXISTS"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
	ErrCodeStorageError     = "STORAGE_ERROR"
)

// E creates a new Error
func E(op string, code string, message string, err error) error {
	return &Error{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}
