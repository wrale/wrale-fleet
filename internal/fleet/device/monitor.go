package device

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// SecurityEventType represents different types of security events
type SecurityEventType string

const (
	EventAuthentication  SecurityEventType = "authentication"
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
}

// NewSecurityMonitor creates a new security monitoring service
func NewSecurityMonitor(logger *zap.Logger) *SecurityMonitor {
	return &SecurityMonitor{
		logger: logger.With(zap.String("component", "security_monitor")),
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
