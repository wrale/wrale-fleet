package power

import "time"

// PowerState represents the power state of a device
type PowerState string

const (
	// PowerOn represents the powered on state
	PowerOn PowerState = "on"
	// PowerOff represents the powered off state
	PowerOff PowerState = "off"
	// PowerStandby represents the standby state
	PowerStandby PowerState = "standby"
)

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

// Config represents power management configuration
type Config struct {
	Sources         []PowerSource     `json:"sources"`
	MinVoltage      float64          `json:"min_voltage"`
	MaxVoltage      float64          `json:"max_voltage"`
	GPIO           string           `json:"gpio"`
	PowerPins      []int            `json:"power_pins"`
	BatteryADCPath string           `json:"battery_adc_path"`
	VoltageADCPath string           `json:"voltage_adc_path"`
	CurrentADCPath string           `json:"current_adc_path"`
	MonitorInterval time.Duration    `json:"monitor_interval"`
	OnPowerCritical func()          `json:"-"`
}