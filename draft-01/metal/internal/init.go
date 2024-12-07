package internal

import (
	"github.com/wrale/wrale-fleet/metal"
	"github.com/wrale/wrale-fleet/metal/internal/diag"
	"github.com/wrale/wrale-fleet/metal/internal/gpio"
	"github.com/wrale/wrale-fleet/metal/internal/power"
	"github.com/wrale/wrale-fleet/metal/internal/secure"
	"github.com/wrale/wrale-fleet/metal/internal/thermal"
)

// NewGPIO creates a new GPIO controller
func NewGPIO(opts ...metal.Option) (metal.GPIO, error) {
	return gpio.New(opts...)
}

// NewPowerManager creates a new power manager
func NewPowerManager(config metal.PowerManagerConfig, opts ...metal.Option) (metal.PowerManager, error) {
	return power.New(config, opts...)
}

// NewThermalManager creates a new thermal manager
func NewThermalManager(config metal.ThermalManagerConfig, opts ...metal.Option) (metal.ThermalManager, error) {
	return thermal.New(config, opts...)
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config metal.SecurityManagerConfig, opts ...metal.Option) (metal.SecurityManager, error) {
	return secure.New(config, opts...)
}

// NewDiagnosticManager creates a new diagnostic manager
func NewDiagnosticManager(config metal.DiagnosticManagerConfig, opts ...metal.Option) (metal.DiagnosticManager, error) {
	return diag.New(config, opts...)
}
