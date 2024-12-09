// Package offline provides interfaces and implementations for managing device
// offline capabilities and airgapped operations in the fleet management system.
package offline

import (
	"context"
	"encoding/json"
	"time"
)

// Operation represents a type of operation that can be performed offline
type Operation string

// Supported offline operations
const (
	OpStatusUpdate     Operation = "status_update"
	OpMetricCollection Operation = "metric_collection"
	OpLogCollection    Operation = "log_collection"
	OpConfigValidation Operation = "config_validation"
	OpHealthCheck      Operation = "health_check"
)

// SyncStatus represents the current synchronization state
type SyncStatus struct {
	LastSuccess time.Time     `json:"last_success"`
	LastAttempt time.Time     `json:"last_attempt"`
	LastError   string        `json:"last_error,omitempty"`
	NextSync    time.Time     `json:"next_sync"`
	Interval    time.Duration `json:"interval"`
}

// BufferStats provides information about the local storage buffer
type BufferStats struct {
	TotalSize     int64 `json:"total_size"`
	UsedSize      int64 `json:"used_size"`
	AvailableSize int64 `json:"available_size"`
	ItemCount     int   `json:"item_count"`
}

// Capabilities defines the offline operational capabilities of a device
type Capabilities struct {
	// Core capabilities
	SupportsAirgap      bool              `json:"supports_airgap"`
	SupportedOperations []Operation       `json:"supported_operations,omitempty"`
	LocalBufferSize     int64             `json:"local_buffer_size,omitempty"`
	SyncInterval        time.Duration     `json:"sync_interval,omitempty"`
	SyncSchedule        map[string]string `json:"sync_schedule,omitempty"`
	CustomConfig        json.RawMessage   `json:"custom_config,omitempty"`

	// Current state
	SyncStatus  *SyncStatus  `json:"sync_status,omitempty"`
	BufferStats *BufferStats `json:"buffer_stats,omitempty"`
}

// Validate checks if the capabilities configuration is valid
func (c *Capabilities) Validate() error {
	// Basic validation
	if c.LocalBufferSize < 0 {
		return newError(ErrInvalidConfig, "local buffer size cannot be negative")
	}
	if c.SyncInterval < 0 {
		return newError(ErrInvalidConfig, "sync interval cannot be negative")
	}

	// Operation validation
	for _, op := range c.SupportedOperations {
		switch op {
		case OpStatusUpdate, OpMetricCollection, OpLogCollection,
			OpConfigValidation, OpHealthCheck:
			// Valid operations
		default:
			return newError(ErrInvalidOperation, "unsupported operation: "+string(op))
		}
	}

	// Schedule validation
	for day, schedule := range c.SyncSchedule {
		if !isValidDayOfWeek(day) || !isValidTimeRange(schedule) {
			return newError(ErrInvalidConfig, "invalid sync schedule")
		}
	}

	return nil
}

// Manager defines the interface for managing device offline capabilities
type Manager interface {
	// GetCapabilities retrieves the current offline capabilities
	GetCapabilities(ctx context.Context) (*Capabilities, error)

	// UpdateCapabilities updates the offline capabilities configuration
	UpdateCapabilities(ctx context.Context, caps *Capabilities) error

	// IsSyncDue checks if synchronization is due based on schedule
	IsSyncDue(ctx context.Context) (bool, error)

	// Sync performs a synchronization operation
	Sync(ctx context.Context) error

	// UpdateBufferStats updates the local buffer statistics
	UpdateBufferStats(ctx context.Context, stats *BufferStats) error

	// IsOperationSupported checks if an operation can be performed offline
	IsOperationSupported(ctx context.Context, op Operation) (bool, error)
}

// isValidDayOfWeek checks if the given string is a valid day of week
func isValidDayOfWeek(day string) bool {
	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true,
		"thursday": true, "friday": true, "saturday": true, "sunday": true,
	}
	return validDays[day]
}

// isValidTimeRange checks if the given string is a valid time range (HH:MM-HH:MM)
func isValidTimeRange(timeRange string) bool {
	// Basic implementation - should be expanded for proper time validation
	return len(timeRange) > 0
}
