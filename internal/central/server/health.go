// Package server implements the core central control plane server functionality.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ComponentStatus represents the health status of a server component
type ComponentStatus string

const (
	// StatusHealthy indicates the component is functioning normally
	StatusHealthy ComponentStatus = "healthy"
	// StatusDegraded indicates the component is operating with reduced functionality
	StatusDegraded ComponentStatus = "degraded"
	// StatusUnhealthy indicates the component is not functioning properly
	StatusUnhealthy ComponentStatus = "unhealthy"
	// StatusStarting indicates the component is still initializing
	StatusStarting ComponentStatus = "starting"
)

// HealthChecker defines the interface that components must implement to participate
// in health checking. This enables both connected and airgapped operation modes.
type HealthChecker interface {
	// CheckHealth performs a health check and returns any issues found
	CheckHealth(context.Context) error
}

// HealthStatus represents detailed health information for a component
type HealthStatus struct {
	Status      ComponentStatus `json:"status"`
	Message     string          `json:"message,omitempty"`
	LastChecked time.Time       `json:"last_checked"`
	LastError   string          `json:"last_error,omitempty"`
}

// HealthResponse represents the complete health check response including
// overall system status and individual component details
type HealthResponse struct {
	Status      ComponentStatus          `json:"status"`
	Ready       bool                     `json:"ready"`
	Components  map[string]*HealthStatus `json:"components,omitempty"`
	LastChecked time.Time                `json:"last_checked"`
}

// healthTracker manages component health state and readiness
type healthTracker struct {
	mu         sync.RWMutex
	components map[string]*HealthStatus
	ready      bool
}

// newHealthTracker creates a new health tracking instance
func newHealthTracker() *healthTracker {
	return &healthTracker{
		components: make(map[string]*HealthStatus),
	}
}

// updateComponent updates the health status for a specific component
func (h *healthTracker) updateComponent(name string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	status := &HealthStatus{
		LastChecked: time.Now(),
	}

	if err != nil {
		status.Status = StatusUnhealthy
		status.Message = "Health check failed"
		status.LastError = err.Error()
	} else {
		status.Status = StatusHealthy
		status.Message = "Component operational"
	}

	h.components[name] = status
}

// getStatus returns the current system health status and component details
func (h *healthTracker) getStatus() *HealthResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Deep copy components to avoid map mutations
	components := make(map[string]*HealthStatus, len(h.components))
	for k, v := range h.components {
		copied := *v
		components[k] = &copied
	}

	// Calculate overall status
	status := StatusHealthy
	for _, health := range h.components {
		if health.Status == StatusUnhealthy {
			status = StatusUnhealthy
			break
		} else if health.Status == StatusDegraded {
			status = StatusDegraded
		}
	}

	return &HealthResponse{
		Status:      status,
		Ready:       h.ready,
		Components:  components,
		LastChecked: time.Now(),
	}
}

// setReady marks the system as ready to serve requests
func (h *healthTracker) setReady() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ready = true
}

// isReady returns whether the system is ready to serve requests
func (h *healthTracker) isReady() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ready
}

// handleHealthResponse generates a JSON response for the health check
func (s *Server) handleHealthResponse(w http.ResponseWriter, status *HealthResponse) error {
	response, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("error serializing health response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(response); err != nil {
		return fmt.Errorf("error writing health response: %w", err)
	}

	return nil
}
