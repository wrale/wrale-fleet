package internal

import (
	"github.com/wrale/wrale-fleet/metal"
	"github.com/wrale/wrale-fleet/metal/internal/diag"
	"github.com/wrale/wrale-fleet/metal/internal/gpio"
	"github.com/wrale/wrale-fleet/metal/internal/power"
	"github.com/wrale/wrale-fleet/metal/internal/secure"
	"github.com/wrale/wrale-fleet/metal/internal/thermal"
)

func init() {
	// Register GPIO implementation
	metal.RegisterGPIOFactory(func(opts ...metal.Option) (metal.GPIO, error) {
		return gpio.New(opts...)
	})

	// Register power manager implementation
	metal.RegisterPowerFactory(func(config metal.PowerManagerConfig, opts ...metal.Option) (metal.PowerManager, error) {
		return power.New(config, opts...)
	})

	// Register thermal manager implementation
	metal.RegisterThermalFactory(func(config metal.ThermalManagerConfig, opts ...metal.Option) (metal.ThermalManager, error) {
		return thermal.New(config, opts...)
	})

	// Register security manager implementation
	metal.RegisterSecureFactory(func(config metal.SecurityManagerConfig, opts ...metal.Option) (metal.SecurityManager, error) {
		return secure.New(config, opts...)
	})

	// Register diagnostic manager implementation
	metal.RegisterDiagnosticFactory(func(config metal.DiagnosticManagerConfig, opts ...metal.Option) (metal.DiagnosticManager, error) {
		return diag.New(config, opts...)
	})
}
