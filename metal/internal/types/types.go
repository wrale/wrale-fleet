package types

import (
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Config defines hardware diagnostics configuration
type Config struct {
	GPIO          metal.GPIO
	Power         metal.PowerManager
	Thermal       metal.ThermalManager
	Security      metal.SecurityManager
	GPIOPins      map[string]int
	RetryAttempts int
	LoadTestTime  time.Duration
	MinVoltage    float64
	TempRange     [2]float64
	OnTestComplete func(metal.TestResult)
}

// Constants moved to public metal package

// Common file operations interface
type FileStore interface {
	SaveState(deviceID string, state interface{}) error
	LoadState(deviceID string) (interface{}, error)
}
