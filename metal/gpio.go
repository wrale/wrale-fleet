package metal

import "context"

// NewGPIO creates a new GPIO controller instance
func NewGPIO(opts ...Option) (GPIO, error) {
    return NewGPIOWithContext(context.Background(), opts...)
}

// NewGPIOWithContext creates a new GPIO controller with context
func NewGPIOWithContext(ctx context.Context, opts ...Option) (GPIO, error) {
    return internal.NewGPIO(ctx, opts...)
}

// GPIO Options

// WithSimulation enables GPIO simulation mode
func WithSimulation(enabled bool) Option {
    return func(v interface{}) error {
        if g, ok := v.(interface{ SetSimulated(bool) }); ok {
            g.SetSimulated(enabled)
            return nil
        }
        return ErrNotSupported
    }
}

// WithPullMode sets pin pull mode
func WithPullMode(mode PullMode) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPullMode(PullMode) error }); ok {
            return p.setPullMode(mode)
        }
        return ErrNotSupported
    }
}

// WithPWMFrequency sets PWM base frequency
func WithPWMFrequency(freq uint32) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPWMFrequency(uint32) error }); ok {
            return p.setPWMFrequency(freq)
        }
        return ErrNotSupported
    }
}
