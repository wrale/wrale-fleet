// Package options provides configuration and initialization for the wfdevice command.
package options

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/cmd/wfdevice/logger"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/server"
)

var (
	// Global server instance for the running device agent
	globalServer     *server.Server
	globalServerLock sync.RWMutex
)

// Config holds the command-line options for wfdevice.
// This aligns with wfcentral's configuration structure while maintaining
// device-specific features.
type Config struct {
	// Port is the main HTTP server port for device operations
	Port string

	// DataDir is the path for persistent storage
	DataDir string

	// LogLevel controls logging verbosity
	LogLevel string

	// LogFile specifies the path for log output (empty for stdout)
	LogFile string

	// ManagementPort is the port for health and readiness endpoints
	// This must be explicitly configured for proper security setup
	ManagementPort string

	// HealthExposure controls how much information is exposed in health endpoints
	// Valid values are: "minimal", "standard", "full"
	// - minimal: Only basic health status
	// - standard: Includes version and uptime (default)
	// - full: All available health information
	HealthExposure string

	// Device-specific configurations
	Name         string            // Device identifier
	ControlPlane string            // Control plane address
	Tags         map[string]string // Device metadata tags
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		Port:           "9090",              // Default main API port
		DataDir:        "/var/lib/wfdevice", // Default data directory
		LogLevel:       "info",              // Default log level
		HealthExposure: "standard",          // Default to standard health information exposure
		Tags:           make(map[string]string),
	}
}

// Validate performs comprehensive configuration validation
func (c *Config) Validate() error {
	// Validate required fields
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	if c.DataDir == "" {
		return fmt.Errorf("data directory is required")
	}

	// Validate port numbers
	basePort, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", c.Port)
	}

	// Management port must be explicitly configured
	if c.ManagementPort == "" {
		return fmt.Errorf("management-port must be specified (use --management-port flag)")
	}

	// Validate management port
	mgmtPort, err := strconv.Atoi(c.ManagementPort)
	if err != nil {
		return fmt.Errorf("invalid management port number: %s", c.ManagementPort)
	}

	// Ensure ports are different
	if basePort == mgmtPort {
		return fmt.Errorf("management port must be different from main API port")
	}

	// Validate health exposure level
	switch c.HealthExposure {
	case "minimal", "standard", "full":
		// Valid values
	default:
		return fmt.Errorf("invalid health exposure level: %s (must be minimal, standard, or full)", c.HealthExposure)
	}

	return nil
}

// NewServer creates and configures a new server instance.
func NewServer(cfg *Config) (*server.Server, error) {
	// First validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set logging environment variables before initializing logger
	if os.Getenv("LOG_LEVEL") == "" {
		logLevel := cfg.LogLevel
		if logLevel == "" {
			logLevel = "info"
		}
		if err := os.Setenv("LOG_LEVEL", logLevel); err != nil {
			return nil, fmt.Errorf("setting default log level: %w", err)
		}
	}

	// Initialize logger with environment-based configuration
	log, err := logger.New(logger.Config{
		LogLevel: cfg.LogLevel,
		Stage:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing logger: %w", err)
	}

	// Create server options from validated config
	var opts []server.Option
	opts = append(opts,
		server.WithPort(cfg.Port),
		server.WithDataDir(cfg.DataDir),
		server.WithManagementPort(cfg.ManagementPort),
		server.WithHealthExposure(cfg.HealthExposure),
	)

	// Add optional device-specific configurations
	if cfg.Name != "" {
		opts = append(opts, server.WithName(cfg.Name))
	}
	if cfg.ControlPlane != "" {
		opts = append(opts, server.WithControlPlane(cfg.ControlPlane))
	}
	if len(cfg.Tags) > 0 {
		opts = append(opts, server.WithTags(cfg.Tags))
	}

	// Create server instance
	srv, err := server.New(log, opts...)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}

// GetRunningServer returns the currently running server instance.
// Returns an error if no server is running.
func GetRunningServer() (*server.Server, error) {
	globalServerLock.RLock()
	defer globalServerLock.RUnlock()

	if globalServer == nil {
		return nil, fmt.Errorf("no server is currently running")
	}
	return globalServer, nil
}

// SetRunningServer sets the global server instance.
func SetRunningServer(srv *server.Server) {
	globalServerLock.Lock()
	globalServer = srv
	globalServerLock.Unlock()
}

// ClearRunningServer clears the global server instance.
func ClearRunningServer() {
	globalServerLock.Lock()
	globalServer = nil
	globalServerLock.Unlock()
}

// NewRegistrationClient creates a new client for device registration.
// Implements registration with the control plane.
func NewRegistrationClient(controlPlane string) (*RegistrationClient, error) {
	if controlPlane == "" {
		return nil, fmt.Errorf("control plane address is required")
	}

	return &RegistrationClient{
		controlPlane: controlPlane,
		timeout:      30 * time.Second,
	}, nil
}

// RegistrationClient handles device registration with the control plane.
type RegistrationClient struct {
	controlPlane string
	timeout      time.Duration
}

// Register registers a device with the control plane.
func (c *RegistrationClient) Register(ctx context.Context, name string, tags map[string]string) error {
	// Validate registration parameters
	if name == "" {
		return fmt.Errorf("device name is required")
	}

	// Create a server instance for registration
	cfg := &Config{
		Name:         name,
		ControlPlane: c.controlPlane,
		Tags:         tags,
	}

	// Set required fields with defaults for registration
	cfg.Port = "9090"
	cfg.ManagementPort = "9091"
	cfg.HealthExposure = "minimal" // Use minimal exposure during registration
	cfg.DataDir = "/var/lib/wfdevice"

	srv, err := NewServer(cfg)
	if err != nil {
		return fmt.Errorf("creating server for registration: %w", err)
	}

	// Set as the running server
	SetRunningServer(srv)

	// Run the server to complete registration
	if err := srv.Run(ctx); err != nil {
		ClearRunningServer()
		return fmt.Errorf("running server for registration: %w", err)
	}

	return nil
}
