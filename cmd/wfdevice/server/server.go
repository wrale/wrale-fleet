// Package server implements the core wfdevice agent functionality.
package server

import (
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

// DeviceStatus contains the full device status information
type DeviceStatus struct {
	Name            string            `json:"name"`
	Status          device.Status     `json:"status"`
	Tags            map[string]string `json:"tags,omitempty"`
	ControlPlane    string           `json:"control_plane"`
	Registered      bool             `json:"registered"`
	LastHealthCheck time.Time        `json:"last_health_check,omitempty"`
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
