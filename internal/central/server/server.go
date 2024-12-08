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
	cfg      *Config
	logger   *zap.Logger
	stage    Stage
	device   *device.Service
	httpSrv  *http.Server
	stopOnce sync.Once
	stopped  chan struct{}
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
		cfg:     cfg,
		logger:  logger,
		stage:   Stage1,
		stopped: make(chan struct{}),
	}

	// Initialize device service and other components
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

	// Initialize device service
	if err := s.initDeviceService(); err != nil {
		return fmt.Errorf("device service initialization failed: %w", err)
	}

	// Stage1-specific initialization
	if err := s.initStage1(); err != nil {
		return fmt.Errorf("stage 1 initialization failed: %w", err)
	}

	return nil
}

// Start begins serving requests and blocks until stopped.
func (s *Server) Start(ctx context.Context) error {
	// Initialize HTTP server
	s.httpSrv = &http.Server{
		Addr:    ":" + s.cfg.Port,
		Handler: s.routes(),
	}

	// Start HTTP server
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server", zap.String("addr", s.httpSrv.Addr))
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		return s.Stop()
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
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

// Status returns the current server health status.
func (s *Server) Status(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	// Perform health checks
	if err := s.checkHealth(ctx); err != nil {
		return "unhealthy", err
	}

	return "healthy", nil
}

// routes sets up the HTTP routes based on current stage capabilities.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Stage 1 routes
	s.registerStage1Routes(mux)

	return mux
}

// checkHealth performs comprehensive health checks.
func (s *Server) checkHealth(ctx context.Context) error {
	// Basic connectivity check
	if s.httpSrv == nil {
		return fmt.Errorf("http server not initialized")
	}

	// Device service health check
	if err := s.checkDeviceServiceHealth(ctx); err != nil {
		return fmt.Errorf("device service health check failed: %w", err)
	}

	return nil
}
