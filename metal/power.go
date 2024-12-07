package metal

import "context"

// NewPowerManager creates a new power manager instance
func NewPowerManager(cfg PowerManagerConfig, opts ...Option) (PowerManager, error) {
    return NewPowerManagerWithContext(context.Background(), cfg, opts...)
}

// NewPowerManagerWithContext creates a new power manager with context
func NewPowerManagerWithContext(ctx context.Context, cfg PowerManagerConfig, opts ...Option) (PowerManager, error) {
    return internal.NewPowerManager(ctx, cfg, opts...)
}

// Power Options

// WithBatteryCapacity sets battery capacity in mAh
func WithBatteryCapacity(capacity uint32) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setBatteryCapacity(uint32) error }); ok {
            return p.setBatteryCapacity(capacity)
        }
        return ErrNotSupported
    }
}

// WithPowerSourcePins configures power source GPIO pins
func WithPowerSourcePins(pins map[PowerSource]string) Option {
    return func(v interface{}) error {
        if p, ok := v.(interface{ setPowerSourcePins(map[PowerSource]string) error }); ok {
            return p.setPowerSourcePins(pins)
        }
        return ErrNotSupported
    }
}

// WithPowerLimits sets voltage and current thresholds
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
