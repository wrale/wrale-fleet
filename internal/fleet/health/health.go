// Package health provides core health monitoring functionality for the fleet management system.
package health

import (
	"context"
	"time"
)

// ComponentStatus represents the health status of a system component
type ComponentStatus string

const (
	// StatusHealthy indicates the component is functioning normally
	StatusHealthy ComponentStatus = "healthy"
	// StatusDegraded indicates the component is operating with reduced functionality
	StatusDegraded ComponentStatus = "degraded"
	// StatusUnhealthy indicates the component is not functioning properly
	StatusUnhealthy ComponentStatus = "unhealthy"
	// StatusStarting indicates the component is still initializing
	StatusStarting ComponentStatus = "starting"
)

// HealthChecker defines the interface that components must implement to participate
// in health checking. This enables both connected and airgapped operation modes.
type HealthChecker interface {
	// CheckHealth performs a health check and returns any issues found
	CheckHealth(context.Context) error
}

// Version represents the version information for a component or system
type Version struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit,omitempty"`
	BuildTime string `json:"build_time,omitempty"`
	Stage     uint8  `json:"stage"`
}

// ConnectionStats provides information about component connectivity
type ConnectionStats struct {
	TotalConnections   int64     `json:"total_connections"`
	ActiveConnections  int64     `json:"active_connections"`
	LastConnection     time.Time `json:"last_connection,omitempty"`
	ConnectionFailures int64     `json:"connection_failures"`
	LastFailure        time.Time `json:"last_failure,omitempty"`
	LastFailureMessage string    `json:"last_failure_message,omitempty"`
}

// HealthStatus represents detailed health information for a component
type HealthStatus struct {
	TenantID    string           `json:"tenant_id,omitempty"`
	Status      ComponentStatus  `json:"status"`
	Message     string           `json:"message,omitempty"`
	LastChecked time.Time        `json:"last_checked"`
	LastError   string           `json:"last_error,omitempty"`
	Version     *Version         `json:"version,omitempty"`
	Stats       *ConnectionStats `json:"stats,omitempty"`
	Uptime      time.Duration    `json:"uptime,omitempty"`
}

// HealthResponse represents the complete health check response including
// overall system status and individual component details
type HealthResponse struct {
	TenantID    string                   `json:"tenant_id,omitempty"`
	Status      ComponentStatus          `json:"status"`
	Ready       bool                     `json:"ready"`
	Components  map[string]*HealthStatus `json:"components,omitempty"`
	LastChecked time.Time                `json:"last_checked"`
	Version     *Version                 `json:"version,omitempty"`
	Uptime      time.Duration            `json:"uptime,omitempty"`
}

// ComponentInfo contains metadata about a monitored component
type ComponentInfo struct {
	Name        string
	Description string
	Category    string
	Critical    bool
}

// HealthCheckResult represents the outcome of a component health check
type HealthCheckResult struct {
	Status    ComponentStatus
	Message   string
	Error     error
	Stats     *ConnectionStats
	Timestamp time.Time
}

// Option defines functional options for configuring health checks
type Option func(*options) error

type options struct {
	tenantID string
	timeout  time.Duration
}

// WithTenant sets the tenant context for health operations
func WithTenant(tenantID string) Option {
	return func(o *options) error {
		o.tenantID = tenantID
		return nil
	}
}

// WithTimeout sets a timeout for health check operations
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		o.timeout = timeout
		return nil
	}
}
