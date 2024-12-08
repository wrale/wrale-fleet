package device

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// validateDeviceOperation performs common validation for device operations,
// providing a central point for tenant access control and logging.
func (s *Service) validateDeviceOperation(ctx context.Context, op string, tenantID, deviceID string) (*Device, error) {
	// First validate the tenant context matches the requested tenant
	ctxTenant, err := TenantFromContext(ctx)
	if err != nil {
		s.logError(op, err)
		return nil, err
	}

	// Ensure tenant IDs match, providing detailed context for any mismatch
	if err := ValidateTenantMatch(ctxTenant, tenantID); err != nil {
		s.logError(op, err,
			zap.String("context_tenant", ctxTenant),
			zap.String("requested_tenant", tenantID),
			zap.String("device_id", deviceID))
		return nil, err
	}

	// Retrieve the device to validate its tenant ownership
	device, err := s.store.Get(ctx, tenantID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device for validation: %w", err)
	}

	// Perform a final validation that the device belongs to the tenant
	if err := ValidateTenantAccess(ctx, device); err != nil {
		s.logError(op, err,
			zap.String("device_tenant", device.TenantID),
			zap.String("context_tenant", ctxTenant))
		return nil, err
	}

	return device, nil
}

// validateDeviceUpdate performs validation specific to device updates,
// including both tenant validation and entity validation.
func (s *Service) validateDeviceUpdate(ctx context.Context, device *Device) error {
	// Validate the device entity itself
	if err := device.Validate(); err != nil {
		return fmt.Errorf("invalid device data: %w", err)
	}

	// Validate tenant access permissions
	if err := ValidateTenantAccess(ctx, device); err != nil {
		s.logError("validateDeviceUpdate", err,
			zap.String("device_id", device.ID),
			zap.String("device_tenant", device.TenantID))
		return err
	}

	return nil
}

// validateListOperation ensures list operations maintain tenant isolation
// by validating tenant context and options.
func (s *Service) validateListOperation(ctx context.Context, opts ListOptions) error {
	// If tenant ID is specified in options, validate it matches context
	if opts.TenantID != "" {
		ctxTenant, err := TenantFromContext(ctx)
		if err != nil {
			s.logError("validateListOperation", err)
			return err
		}
		if err := ValidateTenantMatch(ctxTenant, opts.TenantID); err != nil {
			s.logError("validateListOperation", err,
				zap.String("context_tenant", ctxTenant),
				zap.String("requested_tenant", opts.TenantID))
			return err
		}
	}

	return nil
}

// validateConfig checks if a device's configuration update is valid
// within the context of its tenant.
func (s *Service) validateConfig(ctx context.Context, device *Device, config interface{}) error {
	if err := ValidateTenantAccess(ctx, device); err != nil {
		s.logError("validateConfig", err,
			zap.String("device_id", device.ID),
			zap.String("device_tenant", device.TenantID))
		return err
	}

	// Add any additional configuration validation logic here
	return nil
}

// validateStatus ensures a status transition is valid and authorized
// within the tenant context.
func (s *Service) validateStatus(ctx context.Context, device *Device, newStatus Status) error {
	if err := ValidateTenantAccess(ctx, device); err != nil {
		s.logError("validateStatus", err,
			zap.String("device_id", device.ID),
			zap.String("device_tenant", device.TenantID),
			zap.String("new_status", string(newStatus)))
		return err
	}

	// Add any status transition validation logic here
	return nil
}
