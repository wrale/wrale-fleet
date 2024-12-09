package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"go.uber.org/zap"
)

var (
	// Version information should be set at build time
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildTime    = "unknown"
)

// ServerHealth implements health.HealthChecker for the server itself.
// It encapsulates server-specific health checking logic to ensure the
// core server functionality is operating correctly.
type ServerHealth struct {
	server *Server
}

func newServerHealth(s *Server) *ServerHealth {
	return &ServerHealth{
		server: s,
	}
}

// CheckHealth implements health.HealthChecker by verifying core server
// functionality is working correctly. This provides the base health status
// that other components build upon.
func (h *ServerHealth) CheckHealth(ctx context.Context) error {
	// For now we just verify the server is running
	// Future enhancements will add more sophisticated checks
	return nil
}

// registerHealthChecks registers all components that need health monitoring.
// This establishes the foundation for comprehensive system health tracking,
// focusing on critical components first.
func (s *Server) registerHealthChecks() error {
	s.logger.Info("registering component health checks")

	// Register server itself as the foundational component
	serverInfo := health.ComponentInfo{
		Name:        "server",
		Description: "Central control plane server",
		Category:    "core",
		Critical:    true,
	}

	serverHealth := newServerHealth(s)
	if err := s.health.RegisterComponent(s.baseCtx, "server", serverHealth, serverInfo); err != nil {
		return fmt.Errorf("failed to register server health monitoring: %w", err)
	}

	// Register device service as a critical operational component
	deviceInfo := health.ComponentInfo{
		Name:        "device_service",
		Description: "Device management service",
		Category:    "core",
		Critical:    true,
	}

	if err := s.health.RegisterComponent(s.baseCtx, "device_service", s.device, deviceInfo); err != nil {
		return fmt.Errorf("failed to register device service for health monitoring: %w", err)
	}

	return nil
}

// runHealthChecks performs periodic health checks on all registered components.
// It runs as a background goroutine and continues until the context is canceled,
// providing continuous monitoring of system health.
func (s *Server) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Perform health check with system tenant context and reasonable timeout
			_, err := s.health.CheckHealth(ctx,
				health.WithTimeout(5*time.Second),
				health.WithTenant("system"),
			)
			if err != nil {
				s.logger.Error("periodic health check failed",
					zap.Error(err),
				)
			}
		}
	}
}

// handleHealthCheck implements the health check endpoint that provides detailed
// health status information for the entire system. This endpoint respects tenant
// isolation and provides tenant-specific health information.
func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantID := getTenantFromContext(ctx)

		// Add version information to provide deployment context
		version := &health.Version{
			Version:   buildVersion,
			GitCommit: buildCommit,
			BuildTime: buildTime,
			Stage:     uint8(s.stage),
		}

		// Calculate uptime since server start
		uptime := time.Since(s.GetStartTime())

		// Perform health check with a reasonable timeout
		response, err := s.health.CheckHealth(ctx,
			health.WithTimeout(5*time.Second),
			health.WithTenant(tenantID),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Enhance response with version and uptime information
		response.Version = version
		response.Uptime = uptime

		// Return appropriately formatted status information
		w.Header().Set("Content-Type", "application/json")
		if response.Status != health.StatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("failed to write health check response",
				zap.Error(err),
				zap.String("tenant_id", tenantID),
			)
		}
	}
}

// handleReadyCheck implements the readiness check endpoint that indicates whether
// the server is ready to handle requests. This is particularly important during
// startup and for orchestration systems.
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

		// Create response with version and uptime information
		version := &health.Version{
			Version:   buildVersion,
			GitCommit: buildCommit,
			BuildTime: buildTime,
			Stage:     uint8(s.stage),
		}

		response := struct {
			Ready    bool            `json:"ready"`
			Version  *health.Version `json:"version"`
			Uptime   time.Duration   `json:"uptime"`
			TenantID string          `json:"tenant_id,omitempty"`
		}{
			Ready:    ready,
			Version:  version,
			Uptime:   time.Since(s.GetStartTime()),
			TenantID: tenantID,
		}

		w.Header().Set("Content-Type", "application/json")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("failed to write readiness check response",
				zap.Error(err),
				zap.String("tenant_id", tenantID),
			)
		}
	}
}

// getTenantFromContext extracts the tenant ID from the request context.
// If no tenant is found, it returns a default system tenant identifier.
// This method ensures proper multi-tenant isolation in health reporting.
func getTenantFromContext(ctx context.Context) string {
	// TODO: Replace with proper tenant extraction once auth system is in place
	return "system"
}
