package types

import (
	"context"
	"time"
)

// DeviceID uniquely identifies a device
type DeviceID string

// TaskID uniquely identifies a task
type TaskID string

// ResourceType represents a type of resource
type ResourceType string

const (
	ResourceCPU    ResourceType = "cpu"
	ResourceMemory ResourceType = "memory"
	ResourcePower  ResourceType = "power"
)

// PhysicalLocation represents a device's physical location
type PhysicalLocation struct {
	Rack     string `json:"rack"`
	Position int    `json:"position"`
	Zone     string `json:"zone"`
}

// Task represents a scheduled operation
type Task struct {
	ID        TaskID                    `json:"id"`
	Type      string                    `json:"type"`
	Priority  int                       `json:"priority"`
	Deadline  time.Time                 `json:"deadline"`
	DeviceIDs []DeviceID               `json:"device_ids"`
	Operation string                    `json:"operation"`
	Device    DeviceID                 `json:"device"` // Legacy
	Payload   interface{}              `json:"payload"`
	Resources map[ResourceType]float64 `json:"resources"`
}

// DeviceState represents current device state
type DeviceState struct {
	ID          DeviceID         `json:"id"`
	Status      string           `json:"status"`
	LastSeen    time.Time        `json:"last_seen"`
	Metrics     DeviceMetrics    `json:"metrics"`
	Location    PhysicalLocation `json:"location"`
	LastUpdated time.Time        `json:"last_updated"`
	Resources   map[ResourceType]float64 `json:"resources"`
}

// DeviceMetrics represents device performance metrics
type DeviceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIn   uint64  `json:"network_in"`
	NetworkOut  uint64  `json:"network_out"`
	Temperature float64 `json:"temperature"`
	PowerUsage  float64 `json:"power_usage"`
	CPULoad     float64 `json:"cpu_load"`
	UpdatedAt   time.Time `json:"updated_at"`
	Throttled   bool     `json:"throttled"`
}

// Alert represents a system alert
type Alert struct {
	ID        string    `json:"id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	DeviceID  DeviceID  `json:"device_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Recommendation represents a suggested action
type Recommendation struct {
	ID        string     `json:"id"`
	Priority  int        `json:"priority"`
	Action    string     `json:"action"`
	Reason    string     `json:"reason"`
	DeviceIDs []DeviceID `json:"device_ids"`
	CreatedAt time.Time  `json:"created_at"`
}

// ThermalEvent represents a thermal-related event
type ThermalEvent struct {
	DeviceID    DeviceID  `json:"device_id"`
	Zone        string    `json:"zone"`
	Type        string    `json:"type"`
	Temperature float64   `json:"temperature"`
	Threshold   float64   `json:"threshold"`
	Throttled   bool      `json:"throttled"`
	Timestamp   time.Time `json:"timestamp"`
}

// ThermalPolicy defines thermal management policy
type ThermalPolicy struct {
	MaxTemp            float64 `json:"max_temp"`
	MinFanSpeed        uint32  `json:"min_fan_speed"`
	MaxFanSpeed        uint32  `json:"max_fan_speed"`
	ThrottleAt         float64 `json:"throttle_at"`
	CPUWarning         float64 `json:"cpu_warning"`
	CPUCritical        float64 `json:"cpu_critical"`
	AutoThrottle       bool    `json:"auto_throttle"`
	MaxDevicesThrottled int    `json:"max_devices_throttled"`
}

// ZoneThermalMetrics represents thermal metrics for a cooling zone
type ZoneThermalMetrics struct {
	Zone              string   `json:"zone"`
	TotalDevices      int      `json:"total_devices"`
	AverageTemp       float64  `json:"average_temp"`
	MaxTemp           float64  `json:"max_temp"`
	MinTemp           float64  `json:"min_temp"`
	DevicesOverTemp   int      `json:"devices_over_temp"`
	DevicesThrottled  int      `json:"devices_throttled"`
	PolicyViolations  []string `json:"policy_violations"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// FleetAnalysis represents fleet-wide analysis results
type FleetAnalysis struct {
	TotalDevices     int                         `json:"total_devices"`
	HealthyDevices   int                         `json:"healthy_devices"`
	ResourceUsage    map[ResourceType]float64    `json:"resource_usage"`
	Alerts           []Alert                     `json:"alerts"`
	Recommendations  []Recommendation            `json:"recommendations"`
	AnalyzedAt       time.Time                   `json:"analyzed_at"`
}

// StateManager defines the interface for managing device state
type StateManager interface {
	GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	UpdateDeviceState(ctx context.Context, state DeviceState) error
	ListDevices(ctx context.Context) ([]DeviceState, error)
	RemoveDevice(ctx context.Context, deviceID DeviceID) error
	AddDevice(ctx context.Context, state DeviceState) error
}

// DeviceManager defines the interface for accessing device info
type DeviceManager interface {
	GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	ListDevices(ctx context.Context) ([]DeviceState, error)
	GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
}