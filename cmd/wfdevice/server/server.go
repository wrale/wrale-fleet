// Package server implements the core wfdevice agent functionality.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
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

	// serverPIDFile is the name of the file storing the running server's PID
	serverPIDFile = "wfdevice.pid"
)

// DeviceStatus contains the full device status information
type DeviceStatus struct {
	Name            string            `json:"name"`
	Status          device.Status     `json:"status"`
	Tags            map[string]string `json:"tags,omitempty"`
	ControlPlane    string            `json:"control_plane"`
	Registered      bool              `json:"registered"`
	LastHealthCheck time.Time         `json:"last_health_check,omitempty"`
}

// Server represents the wfdevice agent instance
type Server struct {
	cfg     *Config
	logger  *zap.Logger
	stage   Stage
	device  *device.Device
	httpSrv *http.Server

	// State management
	mu         sync.RWMutex
	registered bool
	stopHealth chan struct{}
}

// Config holds the server configuration
type Config struct {
	Port         string
	DataDir      string
	Name         string // Name is optional at startup, required for registration
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

	// Apply and validate each option
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying server option: %w", err)
		}
	}

	// Log configuration for debugging
	logger.Debug("server configuration after applying options",
		zap.String("port", s.cfg.Port),
		zap.String("data_dir", s.cfg.DataDir),
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
		zap.Any("tags", s.cfg.Tags),
	)

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(s.cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	// Initialize device state with minimal configuration
	// Full initialization happens during registration
	s.device = &device.Device{
		Status: device.StatusOffline,
		Tags:   s.cfg.Tags,
	}

	return s, nil
}

// WithPort sets the server port
func WithPort(port string) Option {
	return func(s *Server) error {
		s.cfg.Port = port
		s.logger.Debug("setting server port", zap.String("port", port))
		return nil
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(s *Server) error {
		s.cfg.DataDir = dir
		s.logger.Debug("setting data directory", zap.String("dir", dir))
		return nil
	}
}

// WithName sets the device name
func WithName(name string) Option {
	return func(s *Server) error {
		s.cfg.Name = name
		s.logger.Debug("setting device name", zap.String("name", name))
		return nil
	}
}

// WithControlPlane sets the control plane address
func WithControlPlane(addr string) Option {
	return func(s *Server) error {
		s.logger.Debug("setting control plane address", zap.String("addr", addr))
		s.cfg.ControlPlane = addr
		return nil
	}
}

// WithTags sets the device tags
func WithTags(tags map[string]string) Option {
	return func(s *Server) error {
		s.cfg.Tags = tags
		s.logger.Debug("setting device tags", zap.Any("tags", tags))
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

	// Write PID file
	if err := s.writePIDFile(); err != nil {
		return fmt.Errorf("writing pid file: %w", err)
	}
	defer s.removePIDFile()

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

	// Register with control plane if name is provided
	if s.cfg.Name != "" {
		regCtx, cancel := context.WithTimeout(ctx, registrationTimeout)
		defer cancel()

		if err := s.register(regCtx); err != nil {
			return fmt.Errorf("device registration failed: %w", err)
		}

		// Start health reporting after successful registration
		s.startHealthReporting()
	} else {
		s.logger.Info("device name not provided, skipping registration",
			zap.String("status", string(s.device.Status)))
	}

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down agent")
		return s.shutdown()
	case err := <-errChan:
		return fmt.Errorf("agent error: %w", err)
	}
}

// Stop initiates a graceful shutdown of the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping server")
	return s.shutdown()
}

// Status returns the current device status
func (s *Server) Status(ctx context.Context) (*DeviceStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &DeviceStatus{
		Name:         s.cfg.Name,
		Status:       s.device.Status,
		Tags:         s.device.Tags,
		ControlPlane: s.cfg.ControlPlane,
		Registered:   s.registered,
	}, nil
}

// NotifyShutdown informs the control plane of a planned shutdown
func (s *Server) NotifyShutdown(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.registered {
		return fmt.Errorf("device not registered with control plane")
	}

	s.notifyShutdown()
	return nil
}

// GetRunningPID returns the PID of the running server, if any
func GetRunningPID(dataDir string) (int, error) {
	pidFile := filepath.Join(dataDir, serverPIDFile)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, fmt.Errorf("invalid pid file content: %w", err)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, nil
	}

	// Check if the process is actually running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return 0, nil
	}

	return pid, nil
}

// pidFilePath returns the path to the PID file
func (s *Server) pidFilePath() string {
	return filepath.Join(s.cfg.DataDir, serverPIDFile)
}

// writePIDFile writes the current process ID to the PID file
func (s *Server) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(s.pidFilePath(), []byte(fmt.Sprintf("%d", pid)), 0644)
}

// removePIDFile removes the PID file
func (s *Server) removePIDFile() {
	if err := os.Remove(s.pidFilePath()); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("failed to remove pid file", zap.Error(err))
	}
}

// shutdown performs a graceful server shutdown
func (s *Server) shutdown() error {
	// Stop health reporting
	if s.stopHealth != nil {
		close(s.stopHealth)
	}

	// Notify control plane of shutdown if registered
	s.mu.RLock()
	if s.registered {
		s.notifyShutdown()
	}
	s.mu.RUnlock()

	// Shutdown HTTP server
	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
	}

	// Remove PID file
	s.removePIDFile()

	return nil
}

// register handles device registration with the control plane
func (s *Server) register(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate registration requirements
	if s.cfg.Name == "" {
		return fmt.Errorf("device name is required for registration")
	}

	s.logger.Info("registering device with control plane",
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
	)

	// Update device identity now that we have the name
	s.device.Name = s.cfg.Name

	// TODO: Implement actual registration logic with control plane
	time.Sleep(time.Second)

	s.registered = true
	s.device.Status = device.StatusOnline

	s.logger.Info("device registration successful")
	return nil
}

// notifyShutdown informs the control plane of planned shutdown
func (s *Server) notifyShutdown() {
	s.logger.Info("notifying control plane of shutdown")
	// TODO: Implement shutdown notification to control plane
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
	// TODO: Implement health report submission to control plane
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

// handleHealth handles basic health check requests
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}
}

// handleStatus handles status check requests
func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		status, err := s.Status(r.Context())
		s.mu.RUnlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
