// Package options provides configuration and initialization for the wfdevice command.
package options

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/cmd/wfdevice/server"
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	// Logging configuration
	LogLevel string // Logging level (debug, info, warn, error)
	LogFile  string // Log file path (empty for stdout)
	LogJSON  bool   // Enable JSON log format
	LogStage int    // Stage-aware logging (1-6)

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
		LogStage:       1,                   // Default to Stage 1 capabilities
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

	// Validate logging configuration
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid values
	default:
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	if c.LogStage < 1 || c.LogStage > 6 {
		return fmt.Errorf("invalid log stage: %d (must be between 1 and 6)", c.LogStage)
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

	// Initialize the logging service
	loggingStore := memory.New()
	loggingService, err := logging.NewService(loggingStore, nil,
		logging.WithRetentionPolicy(logging.EventSystem, 7*24*time.Hour),    // 7 days for system events
		logging.WithRetentionPolicy(logging.EventSecurity, 30*24*time.Hour), // 30 days for security events
	)
	if err != nil {
		return nil, fmt.Errorf("initializing logging service: %w", err)
	}

	// Convert log level
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder based on format preference
	var encoder zapcore.Encoder
	if cfg.LogJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var output zapcore.WriteSyncer
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
		output = zapcore.AddSync(f)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create the logger
	core := zapcore.NewCore(encoder, output, level)
	logger := zap.New(core,
		zap.AddCaller(),
		zap.Fields(
			zap.Int("stage", cfg.LogStage),
			zap.String("component", "wfdevice"),
		),
	)

	// Create server options from validated config
	var opts []server.Option
	opts = append(opts,
		server.WithPort(cfg.Port),
		server.WithDataDir(cfg.DataDir),
		server.WithManagementPort(cfg.ManagementPort),
		server.WithHealthExposure(cfg.HealthExposure),
		server.WithLogging(loggingService),
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
	srv, err := server.New(logger, opts...)
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
