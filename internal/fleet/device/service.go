package device

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Service provides device management operations
type Service struct {
	store  Store
	logger *zap.Logger
}

// NewService creates a new device management service
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// Register creates a new device in the system
func (s *Service) Register(ctx context.Context, tenantID, name string) (*Device, error) {
	device := New(tenantID, name)
	
	if err := device.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device data: %w", err)
	}
	
	if err := s.store.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}
	
	s.logger.Info("registered new device",
		zap.String("device_id", device.ID),
		zap.String("tenant_id", device.TenantID),
		zap.String("name", device.Name),
	)
	
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

// UpdateStatus changes the device status
func (s *Service) UpdateStatus(ctx context.Context, tenantID, deviceID string, status Status) error {
	device, err := s.store.Get(ctx, tenantID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}
	
	device.SetStatus(status)
	
	if err := s.store.Update(ctx, device); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	
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
	if err := s.store.Delete(ctx, tenantID, deviceID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	
	s.logger.Info("deleted device",
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
	)
	
	return nil
}