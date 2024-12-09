// Package server implements the device agent server functionality
package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"go.uber.org/zap"
)

const (
	// Default timeouts and intervals
	registrationTimeout = 30 * time.Second
	readHeaderTimeout   = 10 * time.Second
	healthCheckInterval = 60 * time.Second
)

// Server represents the device agent server instance
type Server struct {
	// Core components
	logger         *zap.Logger
	loggingService *logging.Service
	device         *device.Device
	health         *health.Service

	// Synchronization
	mu sync.RWMutex

	// HTTP servers
	httpSrv    *http.Server
	mgmtServer *ManagementServer

	// Configuration
	cfg     *Config
	stage   int
	pidFile string

	// State
	startTime    time.Time
	registered   bool
	stopHealth   chan struct{}
	shuttingDown bool
}

// Config holds server configuration options
type Config struct {
	Name         string
	Port         string
	DataDir      string
	LogLevel     string
	ControlPlane string
	Stage        int
}

// Option is a functional option for configuring the server
type Option func(*Server) error

// WithLogging sets the logging service
func WithLogging(svc *logging.Service) Option {
	return func(s *Server) error {
		s.loggingService = svc
		return nil
	}
}

// WithHealth sets the health service
func WithHealth(svc *health.Service) Option {
	return func(s *Server) error {
		s.health = svc
		return nil
	}
}

// WithStage sets the capability stage
func WithStage(stage int) Option {
	return func(s *Server) error {
		if stage < 1 || stage > 6 {
			return fmt.Errorf("invalid stage: %d (must be 1-6)", stage)
		}
		s.stage = stage
		return nil
	}
}

// WithPIDFile sets the PID file path
func WithPIDFile(path string) Option {
	return func(s *Server) error {
		s.pidFile = path
		return nil
	}
}

// New creates a new server instance with the provided configuration and options
func New(cfg *Config, logger *zap.Logger, opts ...Option) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	s := &Server{
		logger:    logger,
		cfg:       cfg,
		stage:     1, // Default to Stage 1
		startTime: time.Now().UTC(),
		device:    device.New("", cfg.Name), // TenantID will be set during registration
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	// Validate required components
	if s.loggingService == nil {
		return nil, fmt.Errorf("logging service is required")
	}
	if s.health == nil {
		return nil, fmt.Errorf("health service is required")
	}

	return s, nil
}

// Config returns a copy of the current server configuration
func (s *Server) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.cfg
}

// Stage returns the current capability stage
func (s *Server) Stage() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stage
}

// IsRegistered returns whether the device is registered with the control plane
func (s *Server) IsRegistered() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registered
}

// IsShuttingDown returns whether the server is in the process of shutting down
func (s *Server) IsShuttingDown() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shuttingDown
}
