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

// ResourceType represents different types of resources that can be managed
type ResourceType string

const (
	ResourceCPU    ResourceType = "cpu"
	ResourceMemory ResourceType = "memory"
	ResourcePower  ResourceType = "power"
	ResourceNet    ResourceType = "network"
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
}

// Task represents a scheduled operation on one or more devices
type Task struct {
	ID          TaskID
	DeviceIDs   []DeviceID
	Operation   string
	Priority    int
	Deadline    time.Time
	Resources   map[ResourceType]float64
	Status      string
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

// FleetAnalysis contains analysis results of fleet state
type FleetAnalysis struct {
	TotalDevices     int
	HealthyDevices   int
	ResourceUsage    map[ResourceType]float64
	Alerts           []Alert
	Recommendations  []Recommendation
	AnalyzedAt       time.Time
}

// Alert represents a warning or notification about fleet state
type Alert struct {
	ID        string
	Severity  string
	Message   string
	DeviceID  DeviceID
	CreatedAt time.Time
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