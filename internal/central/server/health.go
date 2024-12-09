package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"github.com/wrale/wrale-fleet/internal/fleet/health/store/memory"
	"go.uber.org/zap"
)

// initHealthMonitoring sets up the health monitoring service with proper multi-tenant isolation
// and registers critical system components for monitoring. This provides the foundation for
// both connected and airgapped operational modes.
func (s *Server) initHealthMonitoring() error {
	// Create health service with in-memory store for now
	// In production, this would be replaced with a persistent store implementation
	healthStore := memory.New()
	s.health = health.NewService(healthStore, s.logger)

	// Register core server components that require health monitoring.
	// Each component is registered with metadata that helps determine
	// its importance and impact on overall system health.
	deviceInfo := health.ComponentInfo{
		Name:        "device_service",
		Description: "Device management service",
		Category:    "core",
		Critical:    true,
	}

	if err := s.health.RegisterComponent(s.baseCtx, "device_service", s.device, deviceInfo); err != nil {
		return fmt.Errorf("failed to register device service for health monitoring: %w", err)
	}

	// Start periodic health checks in the background
	go s.runHealthChecks(s.baseCtx)

	return nil
}

// handleHealthCheck implements the health check endpoint that provides detailed
// health status information for the entire system. This endpoint respects tenant
// isolation and provides tenant-specific health information.
func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantID := getTenantFromContext(ctx)

		// Perform health check with a reasonable timeout to prevent long-running checks
		// from impacting system performance. The WithTenant option ensures proper
		// multi-tenant isolation.
		response, err := s.health.CheckHealth(ctx,
			health.WithTimeout(5*time.Second),
			health.WithTenant(tenantID),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return JSON response with proper content type header
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("failed to write health check response",
				"error", err,
				"tenant_id", tenantID,
			)
		}
	}
}

// handleReadyCheck implements the readiness check endpoint that indicates whether
// the server is ready to handle requests. This is particularly important during
// startup and for orchestration systems that need to know when the server is
// fully operational.
func (s *Server) handleReadyCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantID := getTenantFromContext(ctx)

		// Check if the system is ready to serve requests
		ready, err := s.health.IsReady(ctx, health.WithTenant(tenantID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !ready {
			http.Error(w, "server is not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// runHealthChecks performs periodic health checks on all registered components.
// It runs as a background goroutine and continues until the context is canceled.
func (s *Server) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Perform health check with system tenant context
			_, err := s.health.CheckHealth(ctx,
				health.WithTimeout(5*time.Second),
				health.WithTenant("system"),
			)
			if err != nil {
				s.logger.Error("periodic health check failed",
					"error", err,
				)
			}
		}
	}
}

// getTenantFromContext extracts the tenant ID from the request context.
// If no tenant is found, it returns a default system tenant identifier.
func getTenantFromContext(ctx context.Context) string {
	// This should be replaced with proper tenant extraction logic
	// based on your authentication/authorization system
	return "system"
}

// checkDeviceServiceHealth verifies the health of the device management service.
// This is called as part of component health checks and provides detailed
// status information about the device service's operational state.
func (s *Server) checkDeviceServiceHealth(ctx context.Context) error {
	// First verify the device service is initialized
	if s.device == nil {
		s.logger.Error("device service health check failed: service not initialized")
		return fmt.Errorf("device service not initialized")
	}

	// Check if store is accessible by performing a no-op list operation
	if _, err := s.device.List(ctx, device.ListOptions{}); err != nil {
		s.logger.Error("device service health check failed: store access check failed",
			zap.Error(err),
		)
		return fmt.Errorf("device store access check failed: %w", err)
	}

	// Additional health checks can be added here as requirements evolve
	return nil
}
