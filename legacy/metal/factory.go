package metal

import (
	"fmt"
	"sync"
)

var (
	gpioFactory      func(...Option) (GPIO, error)
	gpioFactoryMu    sync.RWMutex
	powerFactory     func(PowerManagerConfig, ...Option) (PowerManager, error)
	powerFactoryMu   sync.RWMutex
	thermalFactory   func(ThermalManagerConfig, ...Option) (ThermalManager, error)
	thermalFactoryMu sync.RWMutex
	secureFactory    func(SecurityManagerConfig, ...Option) (SecurityManager, error)
	secureFactoryMu  sync.RWMutex
	diagFactory      func(DiagnosticManagerConfig, ...Option) (DiagnosticManager, error)
	diagFactoryMu    sync.RWMutex
)

// RegisterGPIOFactory registers the implementation for creating GPIO controllers
func RegisterGPIOFactory(factory func(...Option) (GPIO, error)) {
	gpioFactoryMu.Lock()
	defer gpioFactoryMu.Unlock()
	gpioFactory = factory
}

// RegisterPowerFactory registers the implementation for creating power managers
func RegisterPowerFactory(factory func(PowerManagerConfig, ...Option) (PowerManager, error)) {
	powerFactoryMu.Lock()
	defer powerFactoryMu.Unlock()
	powerFactory = factory
}

// RegisterThermalFactory registers the implementation for creating thermal managers
func RegisterThermalFactory(factory func(ThermalManagerConfig, ...Option) (ThermalManager, error)) {
	thermalFactoryMu.Lock()
	defer thermalFactoryMu.Unlock()
	thermalFactory = factory
}

// RegisterSecureFactory registers the implementation for creating security managers
func RegisterSecureFactory(factory func(SecurityManagerConfig, ...Option) (SecurityManager, error)) {
	secureFactoryMu.Lock()
	defer secureFactoryMu.Unlock()
	secureFactory = factory
}

// RegisterDiagnosticFactory registers the implementation for creating diagnostic managers
func RegisterDiagnosticFactory(factory func(DiagnosticManagerConfig, ...Option) (DiagnosticManager, error)) {
	diagFactoryMu.Lock()
	defer diagFactoryMu.Unlock()
	diagFactory = factory
}

// NewGPIO creates a new GPIO controller using the registered factory
func NewGPIO(opts ...Option) (GPIO, error) {
	gpioFactoryMu.RLock()
	factory := gpioFactory
	gpioFactoryMu.RUnlock()
	
	if factory == nil {
		return nil, fmt.Errorf("no GPIO implementation registered")
	}
	return factory(opts...)
}

// NewPowerManager creates a new power manager using the registered factory
func NewPowerManager(config PowerManagerConfig, opts ...Option) (PowerManager, error) {
	powerFactoryMu.RLock()
	factory := powerFactory
	powerFactoryMu.RUnlock()
	
	if factory == nil {
		return nil, fmt.Errorf("no power manager implementation registered")
	}
	return factory(config, opts...)
}

// NewThermalManager creates a new thermal manager using the registered factory
func NewThermalManager(config ThermalManagerConfig, opts ...Option) (ThermalManager, error) {
	thermalFactoryMu.RLock()
	factory := thermalFactory
	thermalFactoryMu.RUnlock()
	
	if factory == nil {
		return nil, fmt.Errorf("no thermal manager implementation registered")
	}
	return factory(config, opts...)
}

// NewSecurityManager creates a new security manager using the registered factory
func NewSecurityManager(config SecurityManagerConfig, opts ...Option) (SecurityManager, error) {
	secureFactoryMu.RLock()
	factory := secureFactory
	secureFactoryMu.RUnlock()
	
	if factory == nil {
		return nil, fmt.Errorf("no security manager implementation registered")
	}
	return factory(config, opts...)
}

// NewDiagnosticManager creates a new diagnostic manager using the registered factory
func NewDiagnosticManager(config DiagnosticManagerConfig, opts ...Option) (DiagnosticManager, error) {
	diagFactoryMu.RLock()
	factory := diagFactory
	diagFactoryMu.RUnlock()
	
	if factory == nil {
		return nil, fmt.Errorf("no diagnostic manager implementation registered")
	}
	return factory(config, opts...)
}
