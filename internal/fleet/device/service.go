// Package device provides core device management functionality for the fleet management system.
package device

import (
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
