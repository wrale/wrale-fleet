// Package types defines the API data structures and interfaces
package types

import (
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// API Requests

// DeviceCreateRequest represents a request to register a new device
type DeviceCreateRequest struct {
    ID       types.DeviceID          `json:"id"`
    Location types.PhysicalLocation  `json:"location"`
    Config   map[string]interface{}  `json:"config"`
}

// DeviceUpdateRequest represents a request to update device state
type DeviceUpdateRequest struct {
    Status   string                  `json:"status,omitempty"`
    Location *types.PhysicalLocation `json:"location,omitempty"`
    Config   map[string]interface{}  `json:"config,omitempty"`
}

// DeviceCommandRequest represents a device operation request
type DeviceCommandRequest struct {
    Operation string                 `json:"operation"`
    Params    map[string]interface{} `json:"params,omitempty"`
    Timeout   *time.Duration         `json:"timeout,omitempty"`
}

// FleetCommandRequest represents a fleet-wide operation request
type FleetCommandRequest struct {
    Operation string                 `json:"operation"`
    Devices   []types.DeviceID      `json:"devices"`
    Params    map[string]interface{} `json:"params,omitempty"`
    Timeout   *time.Duration         `json:"timeout,omitempty"`
}

// ConfigUpdateRequest represents a configuration update request
type ConfigUpdateRequest struct {
    Config    map[string]interface{} `json:"config"`
    ValidFrom *time.Time            `json:"valid_from,omitempty"`
    ValidTo   *time.Time            `json:"valid_to,omitempty"`
    Devices   []types.DeviceID      `json:"devices,omitempty"`
}

// API Responses

// APIResponse represents a standard API response
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error response
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// DeviceResponse represents a device state response
type DeviceResponse struct {
    ID         types.DeviceID          `json:"id"`
    Status     string                  `json:"status"`
    Location   types.PhysicalLocation  `json:"location"`
    Metrics    types.DeviceMetrics     `json:"metrics"`
    Config     map[string]interface{}  `json:"config"`
    LastUpdate time.Time               `json:"last_update"`
}

// CommandResponse represents an operation response
type CommandResponse struct {
    ID        string      `json:"id"`
    Status    string      `json:"status"`
    StartTime time.Time   `json:"start_time"`
    EndTime   *time.Time  `json:"end_time,omitempty"`
    Result    interface{} `json:"result,omitempty"`
    Error     string      `json:"error,omitempty"`
}

// WebSocket Messages

// WSMessage represents a websocket message
type WSMessage struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

// WSStateUpdate represents a state update message
type WSStateUpdate struct {
    DeviceID types.DeviceID      `json:"device_id"`
    State    types.DeviceState   `json:"state"`
    Time     time.Time           `json:"time"`
}

// WSMetricsUpdate represents a metrics update message
type WSMetricsUpdate struct {
    DeviceID types.DeviceID       `json:"device_id"`
    Metrics  types.DeviceMetrics  `json:"metrics"`
    Time     time.Time            `json:"time"`
}

// WSAlertMessage represents an alert message
type WSAlertMessage struct {
    DeviceID types.DeviceID `json:"device_id,omitempty"`
    Level    string         `json:"level"`
    Message  string         `json:"message"`
    Time     time.Time      `json:"time"`
}

// Service Interfaces

// DeviceService defines the interface for device operations
type DeviceService interface {
    CreateDevice(req *DeviceCreateRequest) (*DeviceResponse, error)
    GetDevice(id types.DeviceID) (*DeviceResponse, error)
    UpdateDevice(id types.DeviceID, req *DeviceUpdateRequest) (*DeviceResponse, error)
    ListDevices() ([]*DeviceResponse, error)
    DeleteDevice(id types.DeviceID) error
    ExecuteCommand(id types.DeviceID, req *DeviceCommandRequest) (*CommandResponse, error)
}

// FleetService defines the interface for fleet-wide operations
type FleetService interface {
    ExecuteFleetCommand(req *FleetCommandRequest) (*CommandResponse, error)
    GetFleetMetrics() (map[string]interface{}, error)
    UpdateConfig(req *ConfigUpdateRequest) error
    GetConfig(devices []types.DeviceID) (map[types.DeviceID]map[string]interface{}, error)
}

// WebSocketService defines the interface for real-time updates 
type WebSocketService interface {
    Subscribe(deviceIDs []types.DeviceID, updates chan<- *WSMessage) error
    Unsubscribe(updates chan<- *WSMessage) error
    Broadcast(msg *WSMessage) error
}

// AuthService defines the interface for authentication/authorization
type AuthService interface {
    Authenticate(token string) (bool, error)
    Authorize(token string, resource string, action string) (bool, error)
    GenerateToken(userID string, roles []string) (string, error)
}
