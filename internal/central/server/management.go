package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ExposureLevel controls how much information is exposed in management endpoints
type ExposureLevel string

const (
	// ExposureMinimal provides only basic health status
	ExposureMinimal ExposureLevel = "minimal"
	// ExposureStandard includes version and uptime information
	ExposureStandard ExposureLevel = "standard"
	// ExposureFull provides all available health information
	ExposureFull ExposureLevel = "full"
)

// ManagementConfig holds configuration for the management server
type ManagementConfig struct {
	// Port for the management server (must be different from main API port)
	Port string

	// ExposureLevel controls information exposure in health endpoints
	ExposureLevel ExposureLevel

	// Custom health check function
	HealthCheck func(context.Context) error

	// Additional readiness criteria
	ReadinessCheck func(context.Context) error
}

// ManagementServer provides health and readiness endpoints on a separate port.
// It follows security best practices by isolating management functionality and
// supporting configurable information exposure levels.
type ManagementServer struct {
	config     *ManagementConfig
	logger     *zap.Logger
	httpServer *http.Server
	startTime  time.Time

	mu            sync.RWMutex
	isReady       bool
	lastCheck     time.Time
	lastCheckErr  error
	shuttingDown  bool
	healthMetrics map[string]interface{}
}

// newManagementServer creates a new management server instance with proper
// security defaults and validation of required components.
func newManagementServer(cfg *ManagementConfig, logger *zap.Logger) (*ManagementServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("management config is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("management port is required")
	}

	s := &ManagementServer{
		config:        cfg,
		logger:        logger,
		startTime:     time.Now().UTC(),
		healthMetrics: make(map[string]interface{}),
	}

	// Set up HTTP routes with security timeouts
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth())
	mux.HandleFunc("/ready", s.handleReadiness())

	// Configure HTTP server with security defaults
	s.httpServer = &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s, nil
}

// start begins serving management endpoints, launching the HTTP server
// in a separate goroutine to avoid blocking.
func (s *ManagementServer) start() error {
	s.logger.Info("starting management server",
		zap.String("port", s.config.Port),
		zap.String("exposure", string(s.config.ExposureLevel)),
	)

	// Start HTTP server in background
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("management server error", zap.Error(err))
		}
	}()

	return nil
}

// stop gracefully shuts down the management server, ensuring all requests
// complete before shutdown.
func (s *ManagementServer) stop(ctx context.Context) error {
	s.mu.Lock()
	s.shuttingDown = true
	s.mu.Unlock()

	s.logger.Info("stopping management server")
	return s.httpServer.Shutdown(ctx)
}

// setReady marks the server as ready to serve requests, used during startup
// and for maintenance mode transitions.
func (s *ManagementServer) setReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isReady = ready
}

// updateHealthMetric safely updates a named health metric while maintaining
// proper synchronization.
func (s *ManagementServer) updateHealthMetric(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthMetrics[name] = value
}

// handleHealth returns an http.HandlerFunc that performs health checks and
// returns status based on the configured exposure level.
func (s *ManagementServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		s.mu.RLock()
		defer s.mu.RUnlock()

		// Check if we're shutting down
		if s.shuttingDown {
			http.Error(w, "Service shutting down", http.StatusServiceUnavailable)
			return
		}

		// Perform health check if configured
		if s.config.HealthCheck != nil {
			if err := s.config.HealthCheck(ctx); err != nil {
				s.lastCheckErr = err
				http.Error(w, "Service unhealthy: "+err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		// Update last check time
		s.lastCheck = time.Now().UTC()
		s.lastCheckErr = nil

		// Prepare response based on exposure level
		response := make(map[string]interface{})
		response["status"] = "healthy"

		switch s.config.ExposureLevel {
		case ExposureFull:
			response["uptime"] = time.Since(s.startTime).String()
			response["last_check"] = s.lastCheck
			response["metrics"] = s.healthMetrics
			fallthrough
		case ExposureStandard:
			response["ready"] = s.isReady
			fallthrough
		case ExposureMinimal:
			// Already includes status
		}

		// Return health status
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("failed to encode health response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// handleReadiness returns an http.HandlerFunc that checks readiness state
// based on configured criteria.
func (s *ManagementServer) handleReadiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		s.mu.RLock()
		defer s.mu.RUnlock()

		// Check if we're shutting down
		if s.shuttingDown {
			http.Error(w, "Service shutting down", http.StatusServiceUnavailable)
			return
		}

		// Verify readiness if check is configured
		if s.config.ReadinessCheck != nil {
			if err := s.config.ReadinessCheck(ctx); err != nil {
				http.Error(w, "Service not ready: "+err.Error(), http.StatusServiceUnavailable)
				return
			}
		}

		// Prepare response
		response := map[string]interface{}{
			"ready": s.isReady,
		}

		// Add extra info for higher exposure levels
		if s.config.ExposureLevel != ExposureMinimal {
			response["uptime"] = time.Since(s.startTime).String()
		}

		// Return readiness status
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("failed to encode readiness response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
