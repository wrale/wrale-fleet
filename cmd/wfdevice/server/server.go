// Package server implements the core wfdevice agent functionality.
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

// Stage represents the agent's operational stage/capability level
type Stage uint8

const (
	// Stage1 provides basic device management capabilities
	Stage1 Stage = 1
	// Future stages will be added here
)

const (
	// readHeaderTimeout defines the amount of time allowed to read
	// request headers. This helps prevent Slowloris DoS attacks.
	readHeaderTimeout = 10 * time.Second

	// registrationTimeout is the maximum time allowed for initial registration
	registrationTimeout = 30 * time.Second

	// healthCheckInterval is the time between health report submissions
	healthCheckInterval = 1 * time.Minute
)

// Server represents the wfdevice agent instance
type Server struct {
	cfg     *Config
	logger  *zap.Logger
	stage   Stage
	device  *device.Device
	httpSrv *http.Server

	// State management
	mu          sync.RWMutex
	registered  bool
	healthTimer *time.Timer
	stopHealth  chan struct{}
}

// Config holds the server configuration
type Config struct {
	Port         string
	DataDir      string
	Name         string
	ControlPlane string
	Tags         map[string]string
}

// Option defines a server option
type Option func(*Server) error

// New creates a new server instance with the given options
func New(logger *zap.Logger, opts ...Option) (*Server, error) {
	s := &Server{
		cfg:        &Config{},
		logger:     logger,
		stage:      Stage1,
		stopHealth: make(chan struct{}),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying server option: %w", err)
		}
	}

	// Validate required configuration
	if s.cfg.Name == "" {
		return nil, fmt.Errorf("device name is required")
	}
	if s.cfg.ControlPlane == "" {
		return nil, fmt.Errorf("control plane address is required")
	}

	// Initialize device state
	s.device = &device.Device{
		Name:   s.cfg.Name,
		Tags:   s.cfg.Tags,
		Status: device.StatusOffline,
	}

	return s, nil
}

// WithPort sets the server port
func WithPort(port string) Option {
	return func(s *Server) error {
		s.cfg.Port = port
		return nil
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(s *Server) error {
		s.cfg.DataDir = dir
		return nil
	}
}

// WithName sets the device name
func WithName(name string) Option {
	return func(s *Server) error {
		s.cfg.Name = name
		return nil
	}
}

// WithControlPlane sets the control plane address
func WithControlPlane(addr string) Option {
	return func(s *Server) error {
		s.cfg.ControlPlane = addr
		return nil
	}
}

// WithTags sets the device tags
func WithTags(tags map[string]string) Option {
	return func(s *Server) error {
		s.cfg.Tags = tags
		return nil
	}
}

// Run starts the server and blocks until the context is canceled
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting wfdevice agent",
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
		zap.Uint8("stage", uint8(s.stage)),
	)

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
		}
	}()

	// Register with control plane
	regCtx, cancel := context.WithTimeout(ctx, registrationTimeout)
	defer cancel()

	if err := s.register(regCtx); err != nil {
		return fmt.Errorf("device registration failed: %w", err)
	}

	// Start health reporting
	s.startHealthReporting()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down agent")
		return s.shutdown()
	case err := <-errChan:
		return fmt.Errorf("agent error: %w", err)
	}
}

// shutdown performs a graceful server shutdown
func (s *Server) shutdown() error {
	// Stop health reporting
	if s.stopHealth != nil {
		close(s.stopHealth)
	}

	// Notify control plane of shutdown
	s.notifyShutdown()

	// Shutdown HTTP server
	if err := s.httpSrv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return nil
}

// register handles device registration with the control plane
func (s *Server) register(ctx context.Context) error {
	s.logger.Info("registering device with control plane",
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
	)

	// TODO: Implement actual registration logic
	// For now, we'll simulate successful registration
	time.Sleep(time.Second)

	s.mu.Lock()
	s.registered = true
	s.device.Status = device.StatusOnline
	s.mu.Unlock()

	s.logger.Info("device registration successful")
	return nil
}

// notifyShutdown informs the control plane of planned shutdown
func (s *Server) notifyShutdown() {
	s.logger.Info("notifying control plane of shutdown")
	// TODO: Implement shutdown notification
}

// startHealthReporting begins periodic health check submissions
func (s *Server) startHealthReporting() {
	go func() {
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.submitHealthReport(); err != nil {
					s.logger.Error("failed to submit health report", zap.Error(err))
				}
			case <-s.stopHealth:
				return
			}
		}
	}()
}

// submitHealthReport sends a health report to the control plane
func (s *Server) submitHealthReport() error {
	s.logger.Debug("submitting health report")
	// TODO: Implement health report submission
	return nil
}

// routes sets up the HTTP routes
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Stage 1 routes
	mux.HandleFunc("/healthz", s.handleHealth())
	mux.HandleFunc("/api/v1/status", s.handleStatus())

	return mux
}

// Basic health check handler
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	}
}

// Status handler
func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		status := s.device.Status
		s.mu.RUnlock()

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"%s"}`, status)
	}
}
