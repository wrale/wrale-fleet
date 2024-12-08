package device

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Service provides device management operations
type Service struct {
	store   Store
	logger  *zap.Logger
	monitor *SecurityMonitor
}

// NewService creates a new device management service
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:   store,
		logger:  logger,
		monitor: NewSecurityMonitor(logger),
	}
}

// Register creates a new device in the system
func (s *Service) Register(ctx context.Context, tenantID, name string) (*Device, error) {
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

	s.logger.Info("registered new device",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.String("name", device.Name),
	)

	s.monitor.RecordAuthAttempt(ctx, device.ID, device.TenantID, "system", true, map[string]string{
		"action": "register",
		"name":   device.Name,
	})

	return device, nil
}

// Get retrieves a device by ID
func (s *Service) Get(ctx context.Context, tenantID, deviceID string) (*Device, error) {
	device, err := s.store.Get(ctx, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device, nil
}

// Update updates an existing device
func (s *Service) Update(ctx context.Context, device *Device) error {
	if err := device.Validate(); err != nil {
		return fmt.Errorf("invalid device data: %w", err)
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

	s.logger.Info("updated device",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.Time("updated_at", device.UpdatedAt),
	)

	return nil
}

// UpdateStatus changes the device status
func (s *Service) UpdateStatus(ctx context.Context, tenantID, deviceID string, status Status) error {
	device, err := s.store.Get(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	oldStatus := device.Status
	device.SetStatus(status)

	if err := s.store.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	s.monitor.RecordStatusChange(ctx, device.ID, device.TenantID, oldStatus, status)

	s.logger.Info("updated device status",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.String("status", string(status)),
	)

	return nil
}

// List retrieves devices matching the given criteria
func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Device, error) {
	devices, err := s.store.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	return devices, nil
}

// Delete removes a device from the system
func (s *Service) Delete(ctx context.Context, tenantID, deviceID string) error {
	// Record the deletion attempt
	s.monitor.RecordEvent(ctx, SecurityEvent{
		Type:      EventConfigChange,
		DeviceID:  deviceID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Success:   true,
		Details: map[string]string{
			"action": "delete",
		},
	})

	if err := s.store.Delete(ctx, tenantID, deviceID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	s.logger.Info("deleted device",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
	)

	return nil
}
