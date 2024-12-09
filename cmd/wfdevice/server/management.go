package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"go.uber.org/zap"
)

// Version information should be set at build time
var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildTime    = "unknown"
)

// managementServer handles health check and readiness endpoints on a separate port
type managementServer struct {
	server     *Server
	httpServer *http.Server
	logger     *zap.Logger
}

// newManagementServer creates a new management server instance
func newManagementServer(s *Server) *managementServer {
	return &managementServer{
		server: s,
		logger: s.logger.Named("management"),
	}
}

// start begins serving management endpoints
func (m *managementServer) start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", m.handleHealthCheck())
	mux.HandleFunc("/readyz", m.handleReadyCheck())

	addr := ":" + m.server.cfg.ManagementConfig.Port

	m.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		m.logger.Info("starting management server",
			zap.String("addr", addr),
			zap.String("port", m.server.cfg.ManagementConfig.Port),
			zap.String("exposure_level", string(m.server.cfg.ManagementConfig.ExposureLevel)),
			zap.String("healthz_endpoint", "http://localhost"+addr+"/healthz"),
			zap.String("readyz_endpoint", "http://localhost"+addr+"/readyz"),
		)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("management server error", zap.Error(err))
		}
	}()

	return nil
}

// stop performs a graceful shutdown of the management server
func (m *managementServer) stop(ctx context.Context) error {
	if m.httpServer != nil {
		m.logger.Info("stopping management server",
			zap.String("port", m.server.cfg.ManagementConfig.Port),
		)
		return m.httpServer.Shutdown(ctx)
	}
	return nil
}

// filterHealthResponse removes sensitive information based on exposure level
func (m *managementServer) filterHealthResponse(response *health.HealthResponse) {
	switch m.server.cfg.ManagementConfig.ExposureLevel {
	case ExposureMinimal:
		// Provide only basic status
		filtered := &health.HealthResponse{
			Status: response.Status,
			Ready:  response.Ready,
		}
		*response = *filtered

	case ExposureStandard:
		// Include version and uptime, but remove detailed component information
		response.Components = nil
		// Keep only basic version info
		if response.Version != nil {
			response.Version.GitCommit = ""
			response.Version.BuildTime = ""
		}

	case ExposureFull:
		// No filtering, expose all information
	}
}

// handleHealthCheck implements the health check endpoint
func (m *managementServer) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Perform health check with a reasonable timeout
		checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		response, err := m.server.health.CheckHealth(checkCtx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add version and uptime information
		response.Version = &health.Version{
			Version:   buildVersion,
			GitCommit: buildCommit,
			BuildTime: buildTime,
			Stage:     uint8(m.server.stage),
		}
		response.Uptime = time.Since(m.server.GetStartTime())

		// Filter response based on exposure level
		m.filterHealthResponse(response)

		// Return appropriately formatted status information
		w.Header().Set("Content-Type", "application/json")
		if response.Status != health.StatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			m.logger.Error("failed to write health check response",
				zap.Error(err))
		}
	}
}

// handleReadyCheck implements the readiness check endpoint
func (m *managementServer) handleReadyCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if the system is ready to serve requests
		ready, err := m.server.health.IsReady(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create response with exposure-level appropriate information
		response := struct {
			Ready   bool                   `json:"ready"`
			Version *health.Version        `json:"version,omitempty"`
			Uptime  time.Duration          `json:"uptime,omitempty"`
			Status  health.ComponentStatus `json:"status,omitempty"`
		}{
			Ready: ready,
		}

		// Add additional information based on exposure level
		if m.server.cfg.ManagementConfig.ExposureLevel != ExposureMinimal {
			response.Version = &health.Version{
				Version: buildVersion,
			}
			response.Uptime = time.Since(m.server.GetStartTime())

			// Add detailed version info only for full exposure
			if m.server.cfg.ManagementConfig.ExposureLevel == ExposureFull {
				response.Version.GitCommit = buildCommit
				response.Version.BuildTime = buildTime
				response.Version.Stage = uint8(m.server.stage)

				// Get current health status
				healthResponse, err := m.server.health.CheckHealth(ctx)
				if err == nil {
					response.Status = healthResponse.Status
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			m.logger.Error("failed to write readiness check response",
				zap.Error(err))
		}
	}
}
