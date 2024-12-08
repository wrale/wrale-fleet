package server

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents component health information
type HealthStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// componentHealth checks individual component health
func (s *Server) componentHealth(ctx context.Context) map[string]*HealthStatus {
	health := make(map[string]*HealthStatus)

	// Check device service health
	if err := s.checkDeviceServiceHealth(ctx); err != nil {
		health["device_service"] = &HealthStatus{
			Status:    "unhealthy",
			Message:   fmt.Sprintf("device service health check failed: %v", err),
			Timestamp: time.Now(),
		}
	} else {
		health["device_service"] = &HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
		}
	}

	// Additional component health checks will be added here
	// as more components are introduced in later stages

	return health
}
