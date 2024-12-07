package types

import (
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Service interfaces

type DeviceService interface {
    CreateDevice(*DeviceCreateRequest) (*DeviceResponse, error)
    GetDevice(types.DeviceID) (*DeviceResponse, error)
    UpdateDevice(types.DeviceID, *DeviceUpdateRequest) (*DeviceResponse, error)
    ListDevices() ([]*DeviceResponse, error)
    DeleteDevice(types.DeviceID) error
    ExecuteCommand(types.DeviceID, *DeviceCommandRequest) (*CommandResponse, error)
}

type FleetService interface {
    GetFleetMetrics() (*FleetMetrics, error)
    ExecuteFleetCommand(*FleetCommandRequest) error
    GetFleetConfig() (map[string]interface{}, error)
    UpdateFleetConfig(map[string]interface{}) error
}

type WebSocketService interface {
    AddClient(*WebSocket.Conn)
    RemoveClient(*WebSocket.Conn)
    GetDeviceUpdates(types.DeviceID) (<-chan interface{}, error)
}

type AuthService interface {
    Authenticate(token string) (bool, error)
    Authorize(token, path, method string) (bool, error)
}

// Request/Response types

type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type DeviceCreateRequest struct {
    ID       types.DeviceID        `json:"id"`
    Location string               `json:"location"`
    Config   map[string]interface{} `json:"config,omitempty"`
}

type DeviceUpdateRequest struct {
    Status   string               `json:"status,omitempty"`
    Location *string             `json:"location,omitempty"`
    Config   map[string]interface{} `json:"config,omitempty"`
}

type DeviceResponse struct {
    ID         types.DeviceID        `json:"id"`
    Status     string               `json:"status"`
    Location   string               `json:"location"`
    Metrics    map[string]float64    `json:"metrics"`
    Config     map[string]interface{} `json:"config"`
    LastUpdate time.Time            `json:"last_update"`
}

type DeviceCommandRequest struct {
    Operation string `json:"operation"`
}

type CommandResponse struct {
    ID        string     `json:"id"`
    Status    string     `json:"status"`
    StartTime time.Time  `json:"start_time"`
    EndTime   *time.Time `json:"end_time,omitempty"`
    Error     string     `json:"error,omitempty"`
}

type FleetMetrics struct {
    TotalDevices   int     `json:"total_devices"`
    ActiveDevices  int     `json:"active_devices"`
    CPUUsage       float64 `json:"cpu_usage"`
    MemoryUsage    float64 `json:"memory_usage"`
    PowerUsage     float64 `json:"power_usage"`
    AverageLatency float64 `json:"average_latency"`
}

type FleetCommandRequest struct {
    Operation      string            `json:"operation"`
    DeviceSelector *DeviceSelector   `json:"device_selector,omitempty"`
}

type DeviceSelector struct {
    Status   []string          `json:"status,omitempty"`
    Location string           `json:"location,omitempty"`
    Metrics  map[string]Range `json:"metrics,omitempty"`
}

type Range struct {
    Min float64 `json:"min"`
    Max float64 `json:"max"`
}