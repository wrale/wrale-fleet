package metal

import (
    "context"
    "fmt"
    "time"
)

// Common test-related options
var (
    // WithRetries sets the retry count for tests
    WithRetries = func(retries int) Option {
        return func(v interface{}) error {
            if retries < 0 {
                return &ValidationError{Field: "retries", Value: retries, Err: ErrInvalidConfig}
            }
            if d, ok := v.(interface{ setRetries(int) }); ok {
                d.setRetries(retries)
                return nil
            }
            return ErrNotSupported
        }
    }

    // WithTimeout sets the test timeout duration
    WithTimeout = func(timeout time.Duration) Option {
        return func(v interface{}) error {
            if timeout <= 0 {
                return &ValidationError{Field: "timeout", Value: timeout, Err: ErrInvalidConfig}
            }
            if d, ok := v.(interface{ setTimeout(time.Duration) }); ok {
                d.setTimeout(timeout)
                return nil
            }
            return ErrNotSupported
        }
    }

    // WithLoadTest enables extended load testing
    WithLoadTest = func(enable bool) Option {
        return func(v interface{}) error {
            if d, ok := v.(interface{ setLoadTest(bool) }); ok {
                d.setLoadTest(enable)
                return nil
            }
            return ErrNotSupported
        }
    }
)

// NewDiagnosticManager creates a new diagnostic manager instance
func NewDiagnosticManager(cfg DiagnosticManagerConfig, opts ...Option) (DiagnosticManager, error) {
    return NewDiagnosticManagerWithContext(context.Background(), cfg, opts...)
}

// NewDiagnosticManagerWithContext creates a new diagnostic manager with context
func NewDiagnosticManagerWithContext(ctx context.Context, cfg DiagnosticManagerConfig, opts ...Option) (DiagnosticManager, error) {
    return internal.NewDiagnosticManager(ctx, cfg, opts...)
}
