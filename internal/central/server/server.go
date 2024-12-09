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
	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"go.uber.org/zap"
)

// Server represents the central control plane server instance.
// It manages Stage 1 capabilities including core device management,
// monitoring, and configuration operations.
type Server struct {
	cfg        *Config
	logger     *zap.Logger
	stage      Stage
	device     *device.Service
	httpSrv    *http.Server
	health     *health.Service
	baseCtx    context.Context
	baseCancel context.CancelFunc
	stopOnce   sync.Once
	stopped    chan struct{}
	readyChan  chan struct{}
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

	// Create base context for server lifetime
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		cfg:        cfg,
		logger:     logger,
		stage:      Stage1,
		baseCtx:    ctx,
		baseCancel: cancel,
		stopped:    make(chan struct{}),
		readyChan:  make(chan struct{}),
	}

	// Initialize server components
	if err := s.initialize(); err != nil {
		cancel() // Clean up context if initialization fails
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

	// Initialize health monitoring first to track component health during startup
	if err := s.initHealthMonitoring(); err != nil {
		return fmt.Errorf("health monitoring initialization failed: %w", err)
	}

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

	// Wait for initial health checks to complete
	checkCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	response, err := s.health.CheckHealth(checkCtx, health.WithTimeout(healthCheckTimeout))
	if err != nil {
		return fmt.Errorf("initial health check failed: %w", err)
	}

	if response.Status != health.StatusHealthy {
		return fmt.Errorf("server failed to become healthy: %s", response.Status)
	}

	// Mark server as ready
	if err := s.health.SetReady(ctx, true); err != nil {
		return fmt.Errorf("failed to set ready status: %w", err)
	}
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

		// Mark server as not ready
		if e := s.health.SetReady(ctx, false); e != nil {
			s.logger.Error("failed to update ready status during shutdown", zap.Error(e))
		}

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

		// Cancel base context and close channels
		s.baseCancel()
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
func (s *Server) Status(ctx context.Context) (*health.HealthResponse, error) {
	return s.health.CheckHealth(ctx, health.WithTimeout(healthCheckTimeout))
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
