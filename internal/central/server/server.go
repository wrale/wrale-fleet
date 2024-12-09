// Package server implements the core central control plane server functionality
// for the Wrale Fleet Management Platform.
package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// Server represents the central control plane server instance.
// It manages Stage 1 capabilities including core device management,
// monitoring, and configuration operations.
type Server struct {
	cfg       *Config
	logger    *zap.Logger
	stage     Stage
	device    *device.Service
	httpSrv   *http.Server
	health    *healthTracker
	stopOnce  sync.Once
	stopped   chan struct{}
	readyChan chan struct{}
}

// Stage represents the server's operational stage/capability level
type Stage uint8

const (
	// Stage1 provides basic device management capabilities
	Stage1 Stage = 1
	// Future stages will be added here as described in CLI strategy
)

const (
	defaultPort        = "8080"
	defaultDataDir     = "/var/lib/wfcentral"
	defaultLogLevel    = "info"
	shutdownTimeout    = 5 * time.Second
	healthCheckTimeout = 5 * time.Second
	// readHeaderTimeout defines the amount of time allowed to read
	// request headers. This helps prevent Slowloris DoS attacks.
	readHeaderTimeout = 10 * time.Second
)

// New creates a new central control plane server instance.
func New(cfg *Config, logger *zap.Logger) (*Server, error) {
	if cfg == nil {
		cfg = &Config{
			Port:     defaultPort,
			DataDir:  defaultDataDir,
			LogLevel: defaultLogLevel,
		}
	}

	s := &Server{
		cfg:       cfg,
		logger:    logger,
		stage:     Stage1,
		health:    newHealthTracker(),
		stopped:   make(chan struct{}),
		readyChan: make(chan struct{}),
	}

	// Initialize server components
	if err := s.initialize(); err != nil {
		return nil, fmt.Errorf("server initialization failed: %w", err)
	}

	return s, nil
}

// initialize sets up all server components.
func (s *Server) initialize() error {
	s.logger.Info("initializing central control plane server",
		zap.String("port", s.cfg.Port),
		zap.String("data_dir", s.cfg.DataDir),
		zap.Uint8("stage", uint8(s.stage)),
	)

	// Stage1-specific initialization
	if err := s.initStage1(); err != nil {
		return fmt.Errorf("stage 1 initialization failed: %w", err)
	}

	return nil
}

// Start begins serving requests and blocks until stopped.
func (s *Server) Start(ctx context.Context) error {
	// Initialize HTTP server with security timeouts
	s.httpSrv = &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           s.routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Start HTTP server
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server",
			zap.String("addr", s.httpSrv.Addr),
			zap.Duration("header_timeout", readHeaderTimeout),
		)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
			return
		}
		close(errChan)
	}()

	// Start health check monitoring
	go s.runHealthChecks(ctx)

	// Wait for initial health checks to complete
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server startup failed: %w", err)
		}
	case <-time.After(healthCheckTimeout):
		return fmt.Errorf("server failed to become healthy within timeout")
	}

	// Mark server as ready after successful startup
	s.health.setReady()
	close(s.readyChan)

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		return s.Stop()
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}

// Stop performs a graceful server shutdown.
func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		s.logger.Info("stopping central control plane server")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Shutdown HTTP server
		if s.httpSrv != nil {
			if e := s.httpSrv.Shutdown(ctx); e != nil {
				err = fmt.Errorf("http server shutdown: %w", e)
				return
			}
		}

		// Clean up device service
		if e := s.cleanupDeviceService(ctx); e != nil {
			err = fmt.Errorf("device service cleanup: %w", e)
			return
		}

		close(s.stopped)
		s.logger.Info("server stopped successfully")
	})

	return err
}

// Ready returns a channel that will be closed when the server is ready to serve requests.
func (s *Server) Ready() <-chan struct{} {
	return s.readyChan
}

// Status returns the current health status of the server and its components.
func (s *Server) Status(ctx context.Context) (*HealthResponse, error) {
	// Perform health checks on all components
	if err := s.checkComponentHealth(ctx); err != nil {
		s.logger.Error("health check failed", zap.Error(err))
	}

	// Get current health status
	return s.health.getStatus(), nil
}

// runHealthChecks performs periodic health checks on all components.
func (s *Server) runHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.checkComponentHealth(ctx); err != nil {
				s.logger.Error("health check failed", zap.Error(err))
			}
		}
	}
}

// checkComponentHealth verifies the health of all server components.
func (s *Server) checkComponentHealth(ctx context.Context) error {
	// Check device service health
	if err := s.checkDeviceServiceHealth(ctx); err != nil {
		s.health.updateComponent("device_service", err)
		return fmt.Errorf("device service health check failed: %w", err)
	}
	s.health.updateComponent("device_service", nil)

	return nil
}

// routes sets up the HTTP routes based on current stage capabilities.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/healthz", s.handleHealthCheck())
	mux.HandleFunc("/readyz", s.handleReadyCheck())

	// Stage 1 routes
	s.registerStage1Routes(mux)

	return mux
}

// handleHealthCheck implements the health check endpoint.
func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		status, err := s.Status(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := s.handleHealthResponse(w, status); err != nil {
			s.logger.Error("failed to write health check response",
				zap.Error(err),
			)
		}
	}
}

// handleReadyCheck implements the readiness check endpoint.
func (s *Server) handleReadyCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.health.isReady() {
			http.Error(w, "server is not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
