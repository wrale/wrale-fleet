package metal

import (
	"context"
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// PowerSource identifies a power supply source
type PowerSource string

const (
	MainPower    PowerSource = "MAIN"
	BatteryPower PowerSource = "BATTERY"
	SolarPower   PowerSource = "SOLAR"
)

// PowerState represents the current power system state
type PowerState struct {
	types.CommonState
	BatteryLevel     float64                `json:"battery_level"`      // 0-100 percent
	Charging         bool                    `json:"charging"`
	Voltage          float64                `json:"voltage"`           // Current voltage
	CurrentDraw      float64                `json:"current_draw"`      // Current amperage
	CurrentSource    PowerSource            `json:"current_source"`
	AvailablePower   map[PowerSource]bool   `json:"available_power"`
	PowerConsumption float64                `json:"power_consumption"` // Watts
	Warnings         []string               `json:"warnings,omitempty"`
}

// PowerManager defines the interface for power management
type PowerManager interface {
	types.Monitor

	// State Management
	GetState() (PowerState, error)
	GetSource() (PowerSource, error)
	GetVoltage() (float64, error)
	GetCurrent() (float64, error)
	
	// Power Control
	SetPowerMode(source PowerSource) error
	EnableCharging(enable bool) error
	
	// Monitoring
	WatchPower(ctx context.Context) (<-chan PowerState, error)
	WatchSource(ctx context.Context) (<-chan PowerSource, error)
	
	// Thresholds
	SetVoltageThresholds(min, max float64) error
	SetCurrentThresholds(min, max float64) error
	
	// Events
	OnCritical(func(PowerState))
	OnWarning(func(PowerState))
	
	// Configuration
	ConfigurePowerSource(source PowerSource, pin string) error
	EnablePowerSource(source PowerSource, enable bool) error
}

// PowerEvent represents a power system event
type PowerEvent struct {
	types.CommonState
	Source    PowerSource `json:"source"`
	Type      string     `json:"type"`
	Reading   float64    `json:"reading"`
	Threshold float64    `json:"threshold"`
	Message   string     `json:"message,omitempty"`
}

// PowerManagerConfig holds configuration for power management
type PowerManagerConfig struct {
	GPIO            types.GPIOController
	MonitorInterval time.Duration
	PowerPins       map[PowerSource]string
	VoltageMin      float64
	VoltageMax      float64
	CurrentMin      float64
	CurrentMax      float64
	OnCritical      func(PowerState)
	OnWarning       func(PowerState)
}