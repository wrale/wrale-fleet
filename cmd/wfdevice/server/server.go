// Package server implements the core wfdevice agent functionality.
package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/health"
	"go.uber.org/zap"
)

// Stage represents the agent's operational stage/capability level
type Stage uint8

const (
	// Stage1 provides basic device management capabilities
	Stage1 Stage = 1
	// Future stages will be added here
)

// ExposureLevel defines how much information is exposed in health endpoints
type ExposureLevel string

const (
	// ExposureMinimal provides only basic health status
	ExposureMinimal ExposureLevel = "minimal"
	// ExposureStandard includes version and uptime information
	ExposureStandard ExposureLevel = "standard"
	// ExposureFull exposes all available health information
	ExposureFull ExposureLevel = "full"
)

const (
	readHeaderTimeout = 10 * time.Second
	registrationTimeout = 30 * time.Second
	healthCheckInterval = 1 * time.Minute
)

// ManagementConfig holds configuration for the management server
type ManagementConfig struct {
	// Port is the port for health and readiness endpoints
	Port string

	// ExposureLevel controls how much information is exposed in health endpoints
	ExposureLevel ExposureLevel
}

// DeviceStatus contains the full device status information
type DeviceStatus struct {
	Name            string            `json:"name"`
	Status          device.Status     `json:"status"`
	Tags            map[string]string `json:"tags,omitempty"`
	ControlPlane    string           `json:"control_plane"`
	Registered      bool             `json:"registered"`
	LastHealthCheck time.Time        `json:"last_health_check,omitempty"`
}

// Config holds the server configuration
type Config struct {
	// Main server configuration
	Port         string
	DataDir      string
	Name         string
	ControlPlane string
	Tags         map[string]string

	// Management server configuration (for health endpoints)
	ManagementConfig *ManagementConfig
}

// Server represents the wfdevice agent instance
type Server struct {
	cfg     *Config
	logger  *zap.Logger
	stage   Stage
	device  *device.Device
	httpSrv *http.Server
	health  *health.Service
	mgmtServer *managementServer

	// State management
	mu         sync.RWMutex
	registered bool
	stopHealth chan struct{}
	startTime  time.Time
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
		startTime:  time.Now().UTC(),
	}

	// Apply and validate each option
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying server option: %w", err)
		}
	}

	// Initialize health service
	healthSrv, err := health.NewService(health.Config{
		Logger: logger.Named("health"),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing health service: %w", err)
	}
	s.health = healthSrv

	// Register base health checks
	if err := s.registerHealthChecks(); err != nil {
		return nil, fmt.Errorf("registering health checks: %w", err)
	}

	// Create management server for health endpoints
	if s.cfg.ManagementConfig != nil {
		s.mgmtServer = newManagementServer(s)
	}

	// Initialize device state with minimal configuration
	// Full initialization happens during registration
	s.device = &device.Device{
		Status: device.StatusOffline,
		Tags:   s.cfg.Tags,
	}

	// Log configuration for debugging
	logger.Debug("server configuration after applying options",
		zap.String("port", s.cfg.Port),
		zap.String("data_dir", s.cfg.DataDir),
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
		zap.Any("tags", s.cfg.Tags),
	)

	return s, nil
}

// GetStartTime returns the server's start time
func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

// Server options

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

// WithManagementPort sets the management server port
func WithManagementPort(port string) Option {
	return func(s *Server) error {
		if s.cfg.ManagementConfig == nil {
			s.cfg.ManagementConfig = &ManagementConfig{
				ExposureLevel: ExposureStandard, // Default to standard exposure
			}
		}
		s.cfg.ManagementConfig.Port = port
		return nil
	}
}

// WithHealthExposure sets the health endpoint exposure level
func WithHealthExposure(level string) Option {
	return func(s *Server) error {
		if s.cfg.ManagementConfig == nil {
			s.cfg.ManagementConfig = &ManagementConfig{
				ExposureLevel: ExposureStandard,
			}
		}
		s.cfg.ManagementConfig.ExposureLevel = ExposureLevel(level)
		return nil
	}
}
