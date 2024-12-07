// Package metal provides hardware abstraction and management interfaces
package metal

// GPIO-specific option helpers

// WithPullMode returns an option that sets pin pull mode
func WithPullMode(mode PullMode) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPullMode(PullMode) error }); ok {
            return p.setPullMode(mode)
        }
        return ErrNotSupported
    }
}

// WithPWMFrequency returns an option that sets PWM base frequency
func WithPWMFrequency(freq uint32) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPWMFrequency(uint32) error }); ok {
            return p.setPWMFrequency(freq)
        }
        return ErrNotSupported
    }
}
