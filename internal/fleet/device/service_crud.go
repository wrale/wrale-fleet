package device

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Register creates a new device in the system with proper tenant isolation.
func (s *Service) Register(ctx context.Context, tenantID, name string) (*Device, error) {
	if err := s.validateSecurityContext(ctx); err != nil {
		return nil, err
	}

	if err := s.validateTenantOperation(ctx, "Register", tenantID); err != nil {
		return nil, err
	}

	device := New(tenantID, name)

	if err := s.validateDeviceUpdate(ctx, device); err != nil {
		s.monitor.RecordAuthAttempt(ctx, "", tenantID, "system", false, map[string]string{
			"action": "register",
			"error":  err.Error(),
		})
		return nil, err
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

	s.recordDeviceAccess(ctx, device, "register", true, map[string]string{
		"name": device.Name,
	})

	return device, nil
}

// Get retrieves a device by ID with tenant validation.
func (s *Service) Get(ctx context.Context, tenantID, deviceID string) (*Device, error) {
	device, err := s.validateDeviceOperation(ctx, "Get", tenantID, deviceID)
	if err != nil {
		return nil, err
	}

	s.recordDeviceAccess(ctx, device, "get", true, nil)
	return device, nil
}

// List retrieves devices matching the given criteria with tenant filtering.
func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Device, error) {
	if err := s.validateListOperation(ctx, opts); err != nil {
		return nil, err
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
				s.recordDeviceAccess(ctx, device, "list", true, nil)
			}
		}
		return allowedDevices, nil
	}

	return devices, nil
}

// Delete removes a device from the system with tenant validation.
func (s *Service) Delete(ctx context.Context, tenantID, deviceID string) error {
	device, err := s.validateDeviceOperation(ctx, "Delete", tenantID, deviceID)
	if err != nil {
		return err
	}

	if err := s.store.Delete(ctx, tenantID, deviceID); err != nil {
		s.recordDeviceAccess(ctx, device, "delete", false, map[string]string{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to delete device: %w", err)
	}

	s.recordDeviceAccess(ctx, device, "delete", true, nil)

	s.logInfo("Delete",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID))

	return nil
}
