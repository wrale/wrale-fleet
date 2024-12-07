package types

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// APIResponse wraps all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Device represents a managed device
type Device struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	Metadata    Metadata  `json:"metadata"`
}

// DeviceCommand represents a command sent to a device
type DeviceCommand struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// FleetMetrics contains fleet-wide metrics
type FleetMetrics struct {
	TotalDevices      int       `json:"total_devices"`
	ActiveDevices     int       `json:"active_devices"`
	AverageUptime     float64   `json:"average_uptime"`
	TotalErrors       int       `json:"total_errors"`
	LastUpdated       time.Time `json:"last_updated"`
}

// FleetConfig contains fleet-wide configuration
type FleetConfig struct {
	UpdateInterval  time.Duration          `json:"update_interval"`
	PolicyDefaults map[string]interface{} `json:"policy_defaults"`
}

// FleetCommand represents a fleet-wide command
type FleetCommand struct {
	Type    string                 `json:"type"`
	Targets []string              `json:"targets"`
	Payload map[string]interface{} `json:"payload"`
}

// Metadata contains device metadata
type Metadata map[string]interface{}

// Service interfaces

// DeviceService handles device operations
type DeviceService interface {
	List(ctx context.Context) ([]Device, error)
	Get(ctx context.Context, id string) (*Device, error)
	Create(ctx context.Context, device *Device) error
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, id string) error
	SendCommand(ctx context.Context, id string, cmd *DeviceCommand) error
}

// FleetService handles fleet-wide operations
type FleetService interface {
	GetMetrics(ctx context.Context) (*FleetMetrics, error)
	GetConfig(ctx context.Context) (*FleetConfig, error)
	UpdateConfig(ctx context.Context, config *FleetConfig) error
	SendCommand(ctx context.Context, cmd *FleetCommand) error
}

// WebSocketService handles real-time updates
type WebSocketService interface {
	HandleConnection(conn *websocket.Conn) error
}

// AuthService handles authentication and authorization
type AuthService interface {
	Authenticate(token string) (bool, error)
	Authorize(token string, path string, method string) (bool, error)
}