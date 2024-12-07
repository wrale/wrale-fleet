package metal

import "time"

// Diagnostic-specific option helpers

// WithRetries returns an option that sets the retry count for tests
func WithRetries(retries int) Option {
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

// WithTimeout returns an option that sets the test timeout duration
func WithTimeout(timeout time.Duration) Option {
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

// WithLoadTest returns an option that enables extended load testing
func WithLoadTest(enable bool) Option {
    return func(v interface{}) error {
        if d, ok := v.(interface{ setLoadTest(bool) }); ok {
            d.setLoadTest(enable)
            return nil
        }
        return ErrNotSupported
    }
}
