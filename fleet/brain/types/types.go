package types

import (
	"time"
)

// DeviceID uniquely identifies a device
type DeviceID string

// DeviceState represents current device state
type DeviceState struct {
	ID       DeviceID  `json:"id"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// ThermalState represents current thermal conditions - mirrors metal/hw/thermal.ThermalState
type ThermalState struct {
	CPUTemp     float64   `json:"cpu_temp"`     // CPU temperature in Celsius
	GPUTemp     float64   `json:"gpu_temp"`     // GPU temperature in Celsius
	AmbientTemp float64   `json:"ambient_temp"` // Ambient temperature in Celsius
	FanSpeed    uint32    `json:"fan_speed"`    // Current fan speed percentage
	Throttled   bool      `json:"throttled"`    // Whether system is throttled
	UpdatedAt   time.Time `json:"updated_at"`   // Last update timestamp
}

// DeviceMetrics represents device performance metrics
type DeviceMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIn   uint64    `json:"network_in"`
	NetworkOut  uint64    `json:"network_out"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ThermalPolicy defines thermal management policy
type ThermalPolicy struct {
	MaxTemp     float64 `json:"max_temp"`      // Maximum temperature threshold
	MinFanSpeed uint32  `json:"min_fan_speed"` // Minimum fan speed percentage
	MaxFanSpeed uint32  `json:"max_fan_speed"` // Maximum fan speed percentage
	ThrottleAt  float64 `json:"throttle_at"`   // Temperature to enable throttling
}