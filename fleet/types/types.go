package types

import (
	"context"
	"time"
)

// DeviceID uniquely identifies a device
type DeviceID string

// Base Types
type TaskID string
type ResourceType string

const (
	ResourceCPU    ResourceType = "cpu"
	ResourceMemory ResourceType = "memory"
	ResourcePower  ResourceType = "power"
)

// State and Versioning Types
type VersionedState struct {
	Version   string                 `json:"version"`
	State     map[string]interface{} `json:"state"`
	Timestamp time.Time              `json:"timestamp"`
	UpdatedAt time.Time              `json:"updated_at"`
	UpdatedBy string                 `json:"updated_by"`
	Source    string                 `json:"source"`
}

type StateChange struct {
	DeviceID      DeviceID        `json:"device_id"`
	OldState      *VersionedState `json:"old_state,omitempty"`
	NewState      VersionedState  `json:"new_state"`
	ChangeType    string          `json:"change_type"`
	Timestamp     time.Time       `json:"timestamp"`
	ConflictState bool            `json:"conflict_state"`
	Changes       []string        `json:"changes,omitempty"`
}

// Device Types
type PhysicalLocation struct {
	Rack     string `json:"rack"`
	Position int    `json:"position"`
	Zone     string `json:"zone"`
}

type DeviceState struct {
	ID          DeviceID         `json:"id"`
	Status      string           `json:"status"`
	LastSeen    time.Time        `json:"last_seen"`
	Metrics     DeviceMetrics    `json:"metrics"`
	Location    PhysicalLocation `json:"location"`
	LastUpdated time.Time        `json:"last_updated"`
	Resources   map[ResourceType]float64 `json:"resources"`
}

type DeviceMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIn   uint64    `json:"network_in"`
	NetworkOut  uint64    `json:"network_out"`
	Temperature float64   `json:"temperature"`
	PowerUsage  float64   `json:"power_usage"`
	CPULoad     float64   `json:"cpu_load"`
	UpdatedAt   time.Time `json:"updated_at"`
	Throttled   bool      `json:"throttled"`
}

// Task Types
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

// Monitoring Types
type Alert struct {
	ID        string    `json:"id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	DeviceID  DeviceID  `json:"device_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Recommendation struct {
	ID        string     `json:"id"`
	Priority  int        `json:"priority"`
	Action    string     `json:"action"`
	Reason    string     `json:"reason"`
	DeviceIDs []DeviceID `json:"device_ids"`
	CreatedAt time.Time  `json:"created_at"`
}

// Thermal Management Types
type ThermalEvent struct {
	DeviceID    DeviceID  `json:"device_id"`
	Zone        string    `json:"zone"`
	Type        string    `json:"type"`
	Temperature float64   `json:"temperature"`
	Threshold   float64   `json:"threshold"`
	Throttled   bool      `json:"throttled"`
	Timestamp   time.Time `json:"timestamp"`
}

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

type ZoneThermalMetrics struct {
	Zone              string    `json:"zone"`
	TotalDevices      int       `json:"total_devices"`
	AverageTemp       float64   `json:"average_temp"`
	MaxTemp           float64   `json:"max_temp"`
	MinTemp           float64   `json:"min_temp"`
	DevicesOverTemp   int       `json:"devices_over_temp"`
	DevicesThrottled  int       `json:"devices_throttled"`
	PolicyViolations  []string  `json:"policy_violations"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Analysis Types
type FleetAnalysis struct {
	TotalDevices     int                         `json:"total_devices"`
	HealthyDevices   int                         `json:"healthy_devices"`
	ResourceUsage    map[ResourceType]float64    `json:"resource_usage"`
	Alerts           []Alert                     `json:"alerts"`
	Recommendations  []Recommendation            `json:"recommendations"`
	AnalyzedAt       time.Time                   `json:"analyzed_at"`
}

// Interface Definitions
type StateManager interface {
	GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	UpdateDeviceState(ctx context.Context, state DeviceState) error
	ListDevices(ctx context.Context) ([]DeviceState, error)
	RemoveDevice(ctx context.Context, deviceID DeviceID) error
	AddDevice(ctx context.Context, state DeviceState) error
}

type DeviceManager interface {
	GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	ListDevices(ctx context.Context) ([]DeviceState, error)
	GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
}
