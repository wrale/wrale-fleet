package device

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// UpdateStatus changes the device status with tenant validation.
func (s *Service) UpdateStatus(ctx context.Context, tenantID, deviceID string, status Status) error {
	// Validate tenant context before any operations
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError("UpdateStatus", err)
		return err
	}
	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError("UpdateStatus", err,
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID),
			zap.String("device_id", deviceID))
		return err
	}

	// Get device with tenant validation
	device, err := s.Get(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	oldStatus := device.Status
	device.SetStatus(status)

	// Update with full tenant validation
	if err := s.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}

	s.monitor.RecordStatusChange(ctx, device.ID, device.TenantID, oldStatus, status)

	s.logInfo("UpdateStatus",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.String("old_status", string(oldStatus)),
		zap.String("new_status", string(status)))

	return nil
}

// Update updates an existing device with full tenant validation.
func (s *Service) Update(ctx context.Context, device *Device) error {
	if err := device.Validate(); err != nil {
		s.logError("Update", fmt.Errorf("invalid device data: %w", err))
		return fmt.Errorf("invalid device data: %w", err)
	}

	// Validate tenant access
	if err := ValidateTenantAccess(ctx, device); err != nil {
		s.logError("Update", err,
			zap.String("device_id", device.ID),
			zap.String("device_tenant", device.TenantID))
		return err
	}

	// Get existing device for comparison
	existing, err := s.store.Get(ctx, device.TenantID, device.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing device: %w", err)
	}

	// Track security-relevant changes
	if existing.NetworkInfo != device.NetworkInfo {
		s.monitor.RecordNetworkChange(ctx, device.ID, device.TenantID, existing.NetworkInfo, device.NetworkInfo)
	}

	if existing.Status != device.Status {
		s.monitor.RecordStatusChange(ctx, device.ID, device.TenantID, existing.Status, device.Status)
	}

	if existing.LastConfigHash != device.LastConfigHash {
		s.monitor.RecordConfigChange(ctx, device.ID, device.TenantID, "system", map[string]interface{}{
			"old_hash": existing.LastConfigHash,
			"new_hash": device.LastConfigHash,
		})
	}

	if err := s.store.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	s.logInfo("Update",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.Time("updated_at", device.UpdatedAt))

	return nil
}
