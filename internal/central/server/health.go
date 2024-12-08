package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// HealthStatus represents component health information
type HealthStatus struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status     string                   `json:"status"`
	Components map[string]*HealthStatus `json:"components,omitempty"`
}

// getComponentHealth checks individual component health
func (s *Server) getComponentHealth(ctx context.Context) map[string]*HealthStatus {
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

// handleHealthResponse writes a JSON health response
func (s *Server) handleHealthResponse(status string, components map[string]*HealthStatus) ([]byte, error) {
	response := HealthResponse{
		Status:     status,
		Components: components,
	}

	return json.Marshal(response)
}
