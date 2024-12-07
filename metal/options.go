package metal

// Common options for all hardware components

// WithSimulation returns an option that enables simulation mode
func WithSimulation() Option {
	return func(v interface{}) error {
		if c, ok := v.(interface{ setSimulation(bool) }); ok {
			c.setSimulation(true)
			return nil
		}
		return ErrNotSupported
	}
}

// WithDeviceID returns an option that sets the device ID
func WithDeviceID(id string) Option {
	return func(v interface{}) error {
		if c, ok := v.(interface{ setDeviceID(string) }); ok {
			if id == "" {
				return &ValidationError{Field: "device_id", Value: id, Err: ErrInvalidConfig}
			}
			c.setDeviceID(id)
			return nil
		}
		return ErrNotSupported
	}
}

// WithMonitorInterval returns an option that sets the monitoring interval
func WithMonitorInterval(interval time.Duration) Option {
	return func(v interface{}) error {
		if c, ok := v.(interface{ setMonitorInterval(time.Duration) }); ok {
			if interval <= 0 {
				return &ValidationError{Field: "monitor_interval", Value: interval, Err: ErrInvalidConfig}
			}
			c.setMonitorInterval(interval)
			return nil
		}
		return ErrNotSupported
	}
}
