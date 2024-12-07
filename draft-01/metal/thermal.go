package metal

// Thermal-specific option helpers

// WithThermalProfile returns an option that sets the default thermal profile
func WithThermalProfile(profile ThermalProfile) Option {
    return func(v interface{}) error {
        if t, ok := v.(interface{ setThermalProfile(ThermalProfile) error }); ok {
            return t.setThermalProfile(profile)
        }
        return ErrNotSupported
    }
}

// WithCoolingCurve returns an option that configures the cooling curve
func WithCoolingCurve(curve *CoolingCurve) Option {
    return func(v interface{}) error {
        if t, ok := v.(interface{ setCoolingCurve(*CoolingCurve) error }); ok {
            return t.setCoolingCurve(curve)
        }
        return ErrNotSupported
    }
}

// WithThermalZone returns an option that adds a thermal zone
func WithThermalZone(zone ThermalZone) Option {
    return func(v interface{}) error {
        if t, ok := v.(interface{ addThermalZone(ThermalZone) error }); ok {
            return t.addThermalZone(zone)
        }
        return ErrNotSupported
    }
}
