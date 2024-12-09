// Package device provides core device management functionality for the fleet management system.
package device

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

// Service provides device management operations with multi-tenant isolation.
// It handles device lifecycle management, status updates, and security validation
// while ensuring strict tenant boundaries are maintained.
type Service struct {
	store   Store
	logger  *zap.Logger
	monitor *SecurityMonitor
}

// NewService creates a new device management service with the provided
// storage backend and logger. It initializes security monitoring for
// audit and compliance tracking.
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:   store,
		logger:  logger,
		monitor: NewSecurityMonitor(logger),
	}
}

// Store returns the device store instance.
// This provides controlled access to the underlying storage implementation
// while maintaining proper encapsulation.
func (s *Service) Store() Store {
	return s.store
}

// CheckHealth performs health validation of the device service and its dependencies.
// It implements the health.HealthChecker interface to participate in system-wide
// health monitoring. This enables both connected and airgapped operation modes to
// verify service health.
func (s *Service) CheckHealth(ctx context.Context) error {
	const op = "Service.CheckHealth"

	// Verify service initialization
	if s.store == nil {
		s.logError(op, fmt.Errorf("store not initialized"))
		return fmt.Errorf("device service store not initialized")
	}
	if s.monitor == nil {
		s.logError(op, fmt.Errorf("security monitor not initialized"))
		return fmt.Errorf("device service security monitor not initialized")
	}

	// Check store accessibility with a no-op list operation
	if _, err := s.store.List(ctx, ListOptions{}); err != nil {
		s.logError(op, fmt.Errorf("store health check failed: %w", err))
		return fmt.Errorf("device store health check failed: %w", err)
	}

	// Verify security monitor by recording a health check event
	s.monitor.RecordEvent(ctx, SecurityEvent{
		Type:     EventComplianceCheck,
		DeviceID: "system",
		TenantID: "system",
		Success:  true,
		Details:  "health check validation",
	})

	s.logInfo(op, zap.String("status", "healthy"))
	return nil
}

// logError logs an error with contextual information
func (s *Service) logError(op string, err error, fields ...zap.Field) {
	fields = append(fields, zap.String("operation", op))
	fields = append(fields, zap.Error(err))
	s.logger.Error("device operation failed", fields...)
}

// logInfo logs an informational message with contextual information
func (s *Service) logInfo(op string, fields ...zap.Field) {
	fields = append(fields, zap.String("operation", op))
	s.logger.Info("device operation completed", fields...)
}
