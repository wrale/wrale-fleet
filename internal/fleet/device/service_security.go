package device

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// recordDeviceAccess logs an access attempt to a device with security context.
// This provides centralized access logging for audit and compliance purposes.
func (s *Service) recordDeviceAccess(ctx context.Context, device *Device, op string, success bool, details map[string]string) {
	ctxTenant, _ := TenantFromContext(ctx)

	s.monitor.RecordEvent(ctx, SecurityEvent{
		Type:      EventAccess,
		DeviceID:  device.ID,
		TenantID:  device.TenantID,
		Timestamp: time.Now().UTC(),
		Success:   success,
		Actor:     ctxTenant,
		Details: map[string]string{
			"operation": op,
			"tenant":    ctxTenant,
		},
	})

	// Add any provided details to the logged event
	for k, v := range details {
		if err := s.monitor.AddEventDetail(ctx, device.ID, k, v); err != nil {
			s.logError("recordDeviceAccess", fmt.Errorf("failed to add event detail: %w", err),
				zap.String("device_id", device.ID),
				zap.String("key", k))
		}
	}
}

// validateTenantOperation performs tenant-level security validation for operations,
// checking both context and explicit tenant parameters.
func (s *Service) validateTenantOperation(ctx context.Context, op string, tenantID string) error {
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError(op, fmt.Errorf("tenant context validation failed: %w", err))
		return err
	}

	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError(op, fmt.Errorf("tenant match validation failed: %w", err),
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID))
		return err
	}

	return nil
}

// recordConfigChange logs configuration changes with security context
// to maintain an audit trail of device modifications.
func (s *Service) recordConfigChange(ctx context.Context, device *Device, oldHash, newHash string) {
	s.monitor.RecordConfigChange(ctx, device.ID, device.TenantID, "system", map[string]interface{}{
		"old_hash":  oldHash,
		"new_hash":  newHash,
		"timestamp": time.Now().UTC(),
	})
}

// validateSecurityContext ensures all required security information
// is present in the context and properly validated.
func (s *Service) validateSecurityContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("nil context provided")
	}

	// Validate tenant presence and format
	if err := EnsureTenant(ctx); err != nil {
		s.logError("validateSecurityContext",
			fmt.Errorf("security context validation failed: %w", err))
		return err
	}

	return nil
}

// recordStatusTransition logs status changes with security context
// to maintain an audit trail of device state changes.
func (s *Service) recordStatusTransition(ctx context.Context, device *Device, oldStatus, newStatus Status) {
	s.monitor.RecordStatusChange(ctx, device.ID, device.TenantID, oldStatus, newStatus)
}
