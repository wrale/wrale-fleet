// Package types defines the core types and interfaces for the fleet brain component
package types

import (
	"context"
	"time"
)

// DeviceID uniquely identifies a device in the fleet
type DeviceID string

// TaskID uniquely identifies a scheduled task
type TaskID string

// TaskType represents different types of tasks that can be scheduled
type TaskType string

// ResourceType represents different types of resources that can be managed
type ResourceType string

// ThermalProfile defines thermal behavior requirements
type ThermalProfile string

const (
	ResourceCPU    ResourceType = "cpu"
	ResourceMemory ResourceType = "memory"
	ResourcePower  ResourceType = "power"
	ResourceNet    ResourceType = "network"

	// Task types for operations
	TaskUpdateThermalPolicy TaskType = "update_thermal_policy"
	TaskSetThermalProfile  TaskType = "set_thermal_profile"
	TaskSetFanSpeed        TaskType = "set_fan_speed"
	TaskSetCoolingMode     TaskType = "set_cooling_mode"

	// Thermal profiles
	ProfileQuiet   ThermalProfile = "QUIET"   // Prioritize noise reduction
	ProfileBalance ThermalProfile = "BALANCE" // Balance noise and cooling
	ProfileCool    ThermalProfile = "COOL"    // Prioritize cooling
	ProfileMax     ThermalProfile = "MAX"     // Maximum cooling
)

// DeviceState represents the current state of a device
type DeviceState struct {
	ID          DeviceID
	Status      string
	Resources   map[ResourceType]float64
	LastUpdated time.Time
	Location    PhysicalLocation
	Metrics     DeviceMetrics
}

// PhysicalLocation represents the physical placement of a device
type PhysicalLocation struct {
	Rack     string
	Position int
	Zone     string
}

// DeviceMetrics contains real-time metrics from a device
type DeviceMetrics struct {
	Temperature float64
	PowerUsage  float64
	CPULoad     float64
	MemoryUsage float64
	
	// Thermal metrics
	ThermalMetrics *ThermalMetrics
}

// ThermalMetrics contains device-level thermal information
type ThermalMetrics struct {
	// Core temperatures
	CPUTemp     float64
	GPUTemp     float64 // Optional: not all devices have GPUs
	AmbientTemp float64
	
	// Status
	FanSpeed    uint32  // Percentage of max speed
	IsThrottled bool    // Whether device is thermally throttled
	LastUpdate  time.Time
}

// ThermalPolicy defines fleet-level thermal management policy
type ThermalPolicy struct {
	// Management profile
	Profile ThermalProfile
	
	// Temperature thresholds (Celsius)
	CPUWarning     float64 // Alert threshold
	CPUCritical    float64 // Action required
	GPUWarning     float64 // Optional GPU thresholds
	GPUCritical    float64
	AmbientWarning float64 // Environmental threshold
	AmbientCritical float64
	
	// Policy behavior
	MonitoringInterval time.Duration // How often to check temperatures
	AlertInterval      time.Duration // Minimum time between alerts
	AutoThrottle       bool         // Whether to auto-throttle at critical temps
	
	// Fleet-wide settings
	ZonePriority      int     // Higher priority zones throttle later
	MaxDevicesThrottled int   // Max devices that can be throttled per zone
}

// ThermalEvent represents a significant thermal incident
type ThermalEvent struct {
	DeviceID    DeviceID
	Zone        string
	Type        string        // "warning" or "critical"
	Temperature float64
	Threshold   float64       // Which threshold was exceeded
	Throttled   bool         // Whether throttling was applied
	Timestamp   time.Time
}

// Task represents a scheduled operation on one or more devices
type Task struct {
	ID          TaskID
	Type        TaskType
	DeviceIDs   []DeviceID
	Operation   string
	Priority    int
	Deadline    time.Time
	Resources   map[ResourceType]float64
	Status      string
	Payload     interface{}
	CreatedAt   time.Time
	ScheduledAt time.Time
}

// Scheduler handles task scheduling and resource allocation
type Scheduler interface {
	Schedule(ctx context.Context, task Task) error
	Cancel(ctx context.Context, taskID TaskID) error
	GetTask(ctx context.Context, taskID TaskID) (*Task, error)
	ListTasks(ctx context.Context) ([]Task, error)
}

// Orchestrator manages fleet-wide operations and coordination
type Orchestrator interface {
	ExecuteTask(ctx context.Context, task Task) error
	GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	UpdateDeviceState(ctx context.Context, state DeviceState) error
	ListDevices(ctx context.Context) ([]DeviceState, error)
}

// DeviceInventory manages device tracking and state
type DeviceInventory interface {
	AddDevice(ctx context.Context, state DeviceState) error
	RemoveDevice(ctx context.Context, deviceID DeviceID) error
	GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
	ListDevices(ctx context.Context) ([]DeviceState, error)
	UpdateState(ctx context.Context, state DeviceState) error
}

// TopologyManager handles physical layout and relationships
type TopologyManager interface {
	GetLocation(ctx context.Context, deviceID DeviceID) (*PhysicalLocation, error)
	UpdateLocation(ctx context.Context, deviceID DeviceID, location PhysicalLocation) error
	GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
	GetDevicesInRack(ctx context.Context, rack string) ([]DeviceState, error)
}

// ResourceOptimizer handles resource allocation and optimization
type ResourceOptimizer interface {
	OptimizeResources(ctx context.Context, devices []DeviceState) ([]DeviceState, error)
	GetResourceUtilization(ctx context.Context) (map[ResourceType]float64, error)
	SuggestPlacements(ctx context.Context, task Task) ([]DeviceID, error)
}

// SituationAnalyzer analyzes fleet state and provides insights
type SituationAnalyzer interface {
	AnalyzeState(ctx context.Context) (*FleetAnalysis, error)
	GetAlerts(ctx context.Context) ([]Alert, error)
	GetRecommendations(ctx context.Context) ([]Recommendation, error)
}

// ThermalManager coordinates thermal management across the fleet
type ThermalManager interface {
	// State management
	UpdateDeviceThermal(ctx context.Context, deviceID DeviceID, metrics *ThermalMetrics) error
	GetDeviceThermal(ctx context.Context, deviceID DeviceID) (*ThermalMetrics, error)
	
	// Policy management
	SetDevicePolicy(ctx context.Context, deviceID DeviceID, policy *ThermalPolicy) error
	GetDevicePolicy(ctx context.Context, deviceID DeviceID) (*ThermalPolicy, error)
	
	// Zone management
	SetZonePolicy(ctx context.Context, zone string, policy *ThermalPolicy) error
	GetZonePolicy(ctx context.Context, zone string) (*ThermalPolicy, error)
	
	// Monitoring and analysis
	GetZoneMetrics(ctx context.Context, zone string) (*ZoneThermalMetrics, error)
	GetThermalEvents(ctx context.Context) ([]ThermalEvent, error)
}

// ZoneThermalMetrics provides thermal analysis for a zone
type ZoneThermalMetrics struct {
	Zone             string
	AverageTemp      float64
	MaxTemp          float64
	MinTemp          float64
	DevicesOverTemp  int
	DevicesThrottled int
	PolicyViolations []string
	TotalDevices     int
	UpdatedAt        time.Time
}

// FleetAnalysis contains analysis results of fleet state
type FleetAnalysis struct {
	TotalDevices     int
	HealthyDevices   int
	ResourceUsage    map[ResourceType]float64
	Alerts           []Alert
	Recommendations  []Recommendation
	AnalyzedAt       time.Time
	
	// Thermal analysis
	ZoneMetrics      map[string]*ZoneThermalMetrics
	ThermalEvents    []ThermalEvent
}

// Alert represents a warning or notification about fleet state
type Alert struct {
	ID        string
	Severity  string
	Message   string
	DeviceID  DeviceID
	CreatedAt time.Time
	Type      string
}

// Recommendation represents a suggested action based on analysis
type Recommendation struct {
	ID          string
	Priority    int
	Action      string
	Reason      string
	DeviceIDs   []DeviceID
	CreatedAt   time.Time
}