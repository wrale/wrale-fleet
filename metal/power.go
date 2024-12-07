package metal

// Power-specific option helpers

// WithBatteryCapacity returns an option that sets battery capacity
func WithBatteryCapacity(capacity uint32) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setBatteryCapacity(uint32) error }); ok {
            return p.setBatteryCapacity(capacity)
        }
        return ErrNotSupported
    }
}

// WithPowerSourcePins returns an option that configures power source pins
func WithPowerSourcePins(pins map[PowerSource]string) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPowerSourcePins(map[PowerSource]string) error }); ok {
            return p.setPowerSourcePins(pins)
        }
        return ErrNotSupported
    }
}

// WithPowerLimits returns an option that sets voltage and current thresholds
func WithPowerLimits(voltageMin, voltageMax, currentMin, currentMax float64) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface {
            setPowerLimits(float64, float64, float64, float64) error
        }); ok {
            return p.setPowerLimits(voltageMin, voltageMax, currentMin, currentMax)
        }
        return ErrNotSupported
    }
}
