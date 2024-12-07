package power

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/gpio"
)

// Interface defines the required methods for power management
type Interface interface {
	GetState() PowerState
}

// PowerState represents the power state of a device
type PowerState struct {
	State          string                `json:"state"`
	AvailablePower map[PowerSource]bool  `json:"available_power"`
	PowerSource    string                `json:"power_source"`
	Voltage        float64               `json:"voltage"`
	Current        float64               `json:"current"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// PowerSource represents the type of power source
type PowerSource string

const (
	// MainsPower represents main power supply
	MainsPower PowerSource = "mains"
	// BatteryPower represents battery power
	BatteryPower PowerSource = "battery"
	// BackupPower represents backup power supply
	BackupPower PowerSource = "backup"
)

const (
	defaultMonitorInterval = 1 * time.Second
)

// Config represents power management configuration
type Config struct {
	Sources         []PowerSource            `json:"sources"`
	MinVoltage      float64                 `json:"min_voltage"`
	MaxVoltage      float64                 `json:"max_voltage"`
	GPIO            *gpio.Controller         `json:"gpio"`
	PowerPins       map[PowerSource]string   `json:"power_pins"`
	BatteryADCPath  string                  `json:"battery_adc_path"`
	VoltageADCPath  string                  `json:"voltage_adc_path"`
	CurrentADCPath  string                  `json:"current_adc_path"`
	MonitorInterval time.Duration           `json:"monitor_interval"`
	OnPowerChange   func(PowerState)        `json:"-"`
	OnPowerCritical func(PowerState)        `json:"-"`
}