package power

import (
	"time"

	"github.com/wrale/wrale-fleet/metal"
)

// Interface defines the required methods for power management
type Interface interface {
	GetState() metal.PowerState
}

// PowerState represents internal power state
type PowerState struct {
	State          string                     `json:"state"`
	AvailablePower map[metal.PowerSource]bool `json:"available_power"`
	PowerSource    string                     `json:"power_source"`
	Voltage        float64                    `json:"voltage"`
	Current        float64                    `json:"current"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

const (
	defaultMonitorInterval = 1 * time.Second
)

// Config represents power management configuration
type Config struct {
	Sources         []metal.PowerSource          `json:"sources"`
	MinVoltage      float64                      `json:"min_voltage"`
	MaxVoltage      float64                      `json:"max_voltage"`
	GPIO            metal.GPIO                   `json:"gpio"`
	PowerPins       map[metal.PowerSource]string `json:"power_pins"`
	BatteryADCPath  string                       `json:"battery_adc_path"`
	VoltageADCPath  string                       `json:"voltage_adc_path"`
	CurrentADCPath  string                       `json:"current_adc_path"`
	MonitorInterval time.Duration                `json:"monitor_interval"`
	OnPowerChange   func(metal.PowerState)       `json:"-"`
	OnPowerCritical func(metal.PowerState)       `json:"-"`
}
