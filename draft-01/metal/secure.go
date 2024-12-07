package metal

// Security-specific option helpers

// WithSecurityLevel returns an option that sets the default security level
func WithSecurityLevel(level SecurityLevel) Option {
    return func(v interface{}) error {
        if s, ok := v.(interface{ setSecurityLevel(SecurityLevel) error }); ok {
            return s.setSecurityLevel(level)
        }
        return ErrNotSupported
    }
}

// WithQuietHours returns an option that sets quiet hours windows
func WithQuietHours(windows []TimeWindow) Option {
    return func(v interface{}) error {
        if s, ok := v.(interface{ setQuietHours([]TimeWindow) error }); ok {
            return s.setQuietHours(windows)
        }
        return ErrNotSupported
    }
}

// WithMotionSensitivity returns an option that sets motion detection sensitivity
func WithMotionSensitivity(sensitivity float64) Option {
    return func(v interface{}) error {
        if s, ok := v.(interface{ setMotionSensitivity(float64) error }); ok {
            return s.setMotionSensitivity(sensitivity)
        }
        return ErrNotSupported
    }
}

// WithVoltageThreshold returns an option that sets voltage tamper threshold
func WithVoltageThreshold(min float64) Option {
    return func(v interface{}) error {
        if s, ok := v.(interface{ setVoltageThreshold(float64) error }); ok {
            return s.setVoltageThreshold(min)
        }
        return ErrNotSupported
    }
}
