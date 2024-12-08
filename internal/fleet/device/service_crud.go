package device

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Register creates a new device in the system with proper tenant isolation.
func (s *Service) Register(ctx context.Context, tenantID, name string) (*Device, error) {
	// Validate tenant context
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError("Register", err)
		return nil, err
	}
	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError("Register", err,
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID))
		return nil, err
	}

	device := New(tenantID, name)

	if err := device.Validate(); err != nil {
		s.monitor.RecordAuthAttempt(ctx, "", tenantID, "system", false, map[string]string{
			"action": "register",
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("invalid device data: %w", err)
	}

	if err := s.store.Create(ctx, device); err != nil {
		s.monitor.RecordAuthAttempt(ctx, device.ID, tenantID, "system", false, map[string]string{
			"action": "register",
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	s.logInfo("Register",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.String("name", device.Name))

	s.monitor.RecordAuthAttempt(ctx, device.ID, device.TenantID, "system", true, map[string]string{
		"action": "register",
		"name":   device.Name,
	})

	return device, nil
}

// Get retrieves a device by ID with tenant validation.
func (s *Service) Get(ctx context.Context, tenantID, deviceID string) (*Device, error) {
	// Validate tenant context
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError("Get", err)
		return nil, err
	}
	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError("Get", err,
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID),
			zap.String("device_id", deviceID))
		return nil, err
	}

	device, err := s.store.Get(ctx, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if err := ValidateTenantAccess(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

// List retrieves devices matching the given criteria with tenant filtering.
func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Device, error) {
	// Validate tenant context matches list options
	if opts.TenantID != "" {
		ctxTenant, err := TenantFromContext(ctx)
		if err != nil {
			s.logError("List", err)
			return nil, err
		}
		if err := ValidateTenantMatch(ctxTenant, opts.TenantID); err != nil {
			s.logError("List", err,
				zap.String("context_tenant", ctxTenant),
				zap.String("requested_tenant", opts.TenantID))
			return nil, err
		}
	}

	devices, err := s.store.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Filter devices by tenant if context tenant is set
	if ctxTenant, err := TenantFromContext(ctx); err == nil {
		var allowedDevices []*Device
		for _, device := range devices {
			if device.TenantID == ctxTenant {
				allowedDevices = append(allowedDevices, device)
			}
		}
		return allowedDevices, nil
	}

	return devices, nil
}

// Delete removes a device from the system with tenant validation.
func (s *Service) Delete(ctx context.Context, tenantID, deviceID string) error {
	// Validate tenant context
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError("Delete", err)
		return err
	}
	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError("Delete", err,
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID),
			zap.String("device_id", deviceID))
		return err
	}

	// Verify device exists and belongs to tenant
	device, err := s.Get(ctx, tenantID, deviceID)
	if err != nil {
		return err
	}

	if err := s.store.Delete(ctx, tenantID, deviceID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	s.monitor.RecordEvent(ctx, SecurityEvent{
		Type:      EventConfigChange,
		DeviceID:  device.ID,
		TenantID:  device.TenantID,
		Timestamp: device.UpdatedAt,
		Success:   true,
		Details: map[string]string{
			"action": "delete",
		},
	})

	s.logInfo("Delete",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))

	return nil
}