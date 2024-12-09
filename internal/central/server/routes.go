package server

import (
	"net/http"

	"go.uber.org/zap"
)

// routes sets up the HTTP routes based on current stage capabilities.
// It provides a clean separation between routing logic and handler implementations,
// making it easier to manage API versioning and capability stages.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints are always available
	s.logger.Info("registering health check endpoints")
	mux.HandleFunc("/healthz", s.handleHealthCheck())
	mux.HandleFunc("/readyz", s.handleReadyCheck())

	// Stage 1 routes for device management
	s.registerStage1Routes(mux)

	s.logger.Info("HTTP routes registered",
		zap.Uint8("stage", uint8(s.stage)),
	)

	return mux
}

// registerStage1Routes registers HTTP routes for Stage 1 capabilities.
// This includes the basic REST API endpoints for device management and
// sets up the foundation for future stage enhancements.
func (s *Server) registerStage1Routes(mux *http.ServeMux) {
	s.logger.Info("registering Stage 1 API routes")

	// Device management endpoints
	mux.HandleFunc("/api/v1/devices", s.handleDevices())
	mux.HandleFunc("/api/v1/devices/", s.handleDeviceByID())

	s.logger.Debug("Stage 1 routes registered",
		zap.Strings("endpoints", []string{
			"/api/v1/devices",
			"/api/v1/devices/",
		}))
}
