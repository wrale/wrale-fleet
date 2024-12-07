package types

import (
	"context"
	"time"
)

// DeviceID uniquely identifies a device
type DeviceID string

// TaskID uniquely identifies a task
type TaskID string

// StateManager defines the interface for managing device state
type StateManager interface {
	GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	UpdateDeviceState(ctx context.Context, state DeviceState) error
	ListDevices(ctx context.Context) ([]DeviceState, error)
	RemoveDevice(ctx context.Context, deviceID DeviceID) error
	AddDevice(ctx context.Context, state DeviceState) error
}

// Task represents a scheduled operation
type Task struct {
	ID        TaskID      `json:"id"`
	Type      string      `json:"type"`
	Priority  int         `json:"priority"`
	Deadline  time.Time   `json:"deadline"`
	DeviceIDs []DeviceID  `json:"device_ids"` // Multiple devices
	Operation string      `json:"operation"`   // Operation type
	Device    DeviceID    `json:"device"`     // Single device (legacy)
	Payload   interface{} `json:"payload"`
}

// DeviceState represents current device state
type DeviceState struct {
	ID          DeviceID         `json:"id"`
	Status      string           `json:"status"`
	LastSeen    time.Time        `json:"last_seen"`
	Metrics     DeviceMetrics    `json:"metrics"`
	Location    PhysicalLocation `json:"location"`
	LastUpdated time.Time        `json:"last_updated"`
}

// PhysicalLocation represents a device's physical location
type PhysicalLocation struct {
	Rack     string `json:"rack"`
	Position int    `json:"position"`
	Zone     string `json:"zone"`
}

// DeviceMetrics represents device performance and thermal metrics
type DeviceMetrics struct {
	// System metrics
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIn   uint64    `json:"network_in"`
	NetworkOut  uint64    `json:"network_out"`
	
	// Thermal metrics
	Temperature float64   `json:"temperature"`  // CPU temperature in Celsius
	PowerUsage  float64   `json:"power_usage"` // Power consumption in watts
	CPULoad     float64   `json:"cpu_load"`    // CPU load percentage
	
	// Additional thermal data
	GPUTemp     float64   `json:"gpu_temp"`     // GPU temperature in Celsius
	AmbientTemp float64   `json:"ambient_temp"` // Ambient temperature in Celsius
	FanSpeed    uint32    `json:"fan_speed"`    // Current fan speed percentage
	Throttled   bool      `json:"throttled"`    // Whether system is throttled
	
	UpdatedAt   time.Time `json:"updated_at"`
}

// ThermalMetrics is an alias for DeviceMetrics for backward compatibility
type ThermalMetrics = DeviceMetrics

// ThermalPolicy defines thermal management policy
type ThermalPolicy struct {
	MaxTemp     float64 `json:"max_temp"`      // Maximum temperature threshold
	MinFanSpeed uint32  `json:"min_fan_speed"` // Minimum fan speed percentage
	MaxFanSpeed uint32  `json:"max_fan_speed"` // Maximum fan speed percentage
	ThrottleAt  float64 `json:"throttle_at"`   // Temperature to enable throttling
}