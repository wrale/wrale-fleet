package metal

import "github.com/wrale/wrale-fleet/metal/internal"

// NewGPIO creates a new GPIO controller
func NewGPIO(opts ...Option) (GPIO, error) {
	return internal.NewGPIO(opts...)
}

// NewPowerManager creates a new power manager
func NewPowerManager(config PowerManagerConfig, opts ...Option) (PowerManager, error) {
	return internal.NewPowerManager(config, opts...)
}

// NewThermalManager creates a new thermal manager
func NewThermalManager(config ThermalManagerConfig, opts ...Option) (ThermalManager, error) {
	return internal.NewThermalManager(config, opts...)
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config SecurityManagerConfig, opts ...Option) (SecurityManager, error) {
	return internal.NewSecurityManager(config, opts...)
}

// NewDiagnosticManager creates a new diagnostic manager
func NewDiagnosticManager(config DiagnosticManagerConfig, opts ...Option) (DiagnosticManager, error) {
	return internal.NewDiagnosticManager(config, opts...)
}
