package device

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	EventAuthentication  SecurityEventType = "authentication"
	EventAccess          SecurityEventType = "access"
	EventConfigChange    SecurityEventType = "config_change"
	EventStatusChange    SecurityEventType = "status_change"
	EventNetworkChange   SecurityEventType = "network_change"
	EventComplianceCheck SecurityEventType = "compliance_check"
)

// SecurityEvent represents a security-relevant event
type SecurityEvent struct {
	Type      SecurityEventType `json:"type"`
	DeviceID  string            `json:"device_id"`
	TenantID  string            `json:"tenant_id"`
	Timestamp time.Time         `json:"timestamp"`
	Success   bool              `json:"success"`
	Details   interface{}       `json:"details,omitempty"`
	Actor     string            `json:"actor,omitempty"`
}

// SecurityMonitor tracks and logs security-relevant events
type SecurityMonitor struct {
	logger *zap.Logger
	mu     sync.RWMutex
	events map[string][]SecurityEvent // DeviceID -> Events
}

// NewSecurityMonitor creates a new security monitoring service
func NewSecurityMonitor(logger *zap.Logger) *SecurityMonitor {
	return &SecurityMonitor{
		logger: logger.With(zap.String("component", "security_monitor")),
		events: make(map[string][]SecurityEvent),
	}
}

// RecordEvent logs a security event
func (m *SecurityMonitor) RecordEvent(ctx context.Context, event SecurityEvent) {
	fields := []zap.Field{
		zap.String("event_type", string(event.Type)),
		zap.String("device_id", event.DeviceID),
		zap.String("tenant_id", event.TenantID),
		zap.Time("timestamp", event.Timestamp),
		zap.Bool("success", event.Success),
	}

	if event.Actor != "" {
		fields = append(fields, zap.String("actor", event.Actor))
	}

	if event.Details != nil {
		fields = append(fields, zap.Any("details", event.Details))
	}

	m.logger.Info("security event", fields...)

	// Store event for potential updates
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.events[event.DeviceID]; !exists {
		m.events[event.DeviceID] = make([]SecurityEvent, 0, 10)
	}
	m.events[event.DeviceID] = append(m.events[event.DeviceID], event)
}

// AddEventDetail adds additional details to the most recent event for a device
func (m *SecurityMonitor) AddEventDetail(ctx context.Context, deviceID string, key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	deviceEvents, exists := m.events[deviceID]
	if !exists || len(deviceEvents) == 0 {
		return fmt.Errorf("no events found for device %s", deviceID)
	}

	// Get the most recent event
	lastEvent := &deviceEvents[len(deviceEvents)-1]

	// Convert details to map if necessary
	details, ok := lastEvent.Details.(map[string]string)
	if !ok {
		details = make(map[string]string)
		if lastEvent.Details != nil {
			// If there were existing details of a different type, preserve them
			details["_previous"] = fmt.Sprintf("%v", lastEvent.Details)
		}
	}

	// Add new detail
	details[key] = value
	lastEvent.Details = details

	// Log the addition
	m.logger.Info("added event detail",
		zap.String("device_id", deviceID),
		zap.String("key", key),
		zap.String("value", value))

	return nil
}

// RecordAuthAttempt logs an authentication attempt
func (m *SecurityMonitor) RecordAuthAttempt(ctx context.Context, deviceID, tenantID, actor string, success bool, details interface{}) {
	m.RecordEvent(ctx, SecurityEvent{
		Type:      EventAuthentication,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   success,
		Details:   details,
		Actor:     actor,
	})
}

// RecordConfigChange logs a configuration change
func (m *SecurityMonitor) RecordConfigChange(ctx context.Context, deviceID, tenantID, actor string, details interface{}) {
	m.RecordEvent(ctx, SecurityEvent{
		Type:      EventConfigChange,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   true,
		Details:   details,
		Actor:     actor,
	})
}

// RecordStatusChange logs a device status change
func (m *SecurityMonitor) RecordStatusChange(ctx context.Context, deviceID, tenantID string, oldStatus, newStatus Status) {
	m.RecordEvent(ctx, SecurityEvent{
		Type:      EventStatusChange,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   true,
		Details: map[string]string{
			"old_status": string(oldStatus),
			"new_status": string(newStatus),
		},
	})
}

// RecordNetworkChange logs network configuration changes
func (m *SecurityMonitor) RecordNetworkChange(ctx context.Context, deviceID, tenantID string, oldInfo, newInfo *NetworkInfo) {
	m.RecordEvent(ctx, SecurityEvent{
		Type:      EventNetworkChange,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   true,
		Details: map[string]interface{}{
			"old_network_info": oldInfo,
			"new_network_info": newInfo,
		},
	})
}

// RecordComplianceCheck logs compliance check results
func (m *SecurityMonitor) RecordComplianceCheck(ctx context.Context, deviceID, tenantID string, status *ComplianceStatus) {
	m.RecordEvent(ctx, SecurityEvent{
		Type:      EventComplianceCheck,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   status.IsCompliant,
		Details:   status,
	})
}
