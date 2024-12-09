package offline

// ErrorCode represents types of offline capability errors
type ErrorCode string

const (
	// ErrInvalidConfig indicates invalid configuration
	ErrInvalidConfig ErrorCode = "INVALID_CONFIG"

	// ErrInvalidOperation indicates an unsupported offline operation
	ErrInvalidOperation ErrorCode = "INVALID_OPERATION"

	// ErrSyncFailed indicates a synchronization failure
	ErrSyncFailed ErrorCode = "SYNC_FAILED"

	// ErrBufferFull indicates the local buffer is at capacity
	ErrBufferFull ErrorCode = "BUFFER_FULL"

	// ErrNotSupported indicates a requested operation is not supported
	ErrNotSupported ErrorCode = "NOT_SUPPORTED"
)

// Error represents an offline capabilities error
type Error struct {
	Code    ErrorCode
	Message string
}

// Error implements the error interface
func (e *Error) Error() string {
	return string(e.Code) + ": " + e.Message
}

// newError creates a new Error with the given code and message
func newError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}
