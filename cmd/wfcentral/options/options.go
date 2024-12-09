// Package options provides configuration and initialization for the wfcentral command.
package options

import (
	"fmt"
	"strconv"

	"github.com/wrale/wrale-fleet/cmd/wfcentral/logger"
	"github.com/wrale/wrale-fleet/internal/central/server"
)

// Config holds the command-line options for wfcentral.
// This separates command-line concerns from the core server configuration.
type Config struct {
	// Port is the main HTTP server port for device management APIs
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
}

// New creates a new Config with sensible default values that prioritize security
// while requiring explicit port configuration. The management port must be
// explicitly set at runtime, so we don't default it here.
func New() *Config {
	return &Config{
		Port:           "8600",               // Default main API port
		DataDir:        "/var/lib/wfcentral", // Default data directory
		LogLevel:       "info",               // Default log level
		HealthExposure: "standard",           // Default to standard health information exposure
	}
}

// NewServer creates and configures a central server instance based on
// the provided configuration options. This method handles the initialization
// of all necessary components including logging, monitoring, and the separate
// management server for health endpoints.
func NewServer(cfg *Config) (*server.Server, error) {
	// Basic validation
	basePort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %s", cfg.Port)
	}

	// Management port must be explicitly configured
	if cfg.ManagementPort == "" {
		return nil, fmt.Errorf("management-port must be specified (use --management-port flag)")
	}

	// Validate management port
	mgmtPort, err := strconv.Atoi(cfg.ManagementPort)
	if err != nil {
		return nil, fmt.Errorf("invalid management port number: %s", cfg.ManagementPort)
	}

	// Ensure ports are different
	if basePort == mgmtPort {
		return nil, fmt.Errorf("management port must be different from main API port")
	}

	// Initialize logger first to ensure proper diagnostics during startup
	log, err := logger.New(logger.Config{
		Level:    cfg.LogLevel,
		FilePath: cfg.LogFile,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing logger: %w", err)
	}

	// Create internal server configuration with management options
	serverConfig := &server.Config{
		Port:     cfg.Port,
		DataDir:  cfg.DataDir,
		LogLevel: cfg.LogLevel,
		ManagementConfig: &server.ManagementConfig{
			Port:          cfg.ManagementPort,
			ExposureLevel: server.ExposureLevel(cfg.HealthExposure),
		},
	}

	// Create and validate server instance
	srv, err := server.New(serverConfig, log)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}

// ValidateHealthExposure checks if the given exposure level is valid.
// This helper function can be used by CLI commands to validate user input
// before attempting server creation.
func ValidateHealthExposure(level string) bool {
	switch server.ExposureLevel(level) {
	case server.ExposureMinimal, server.ExposureStandard, server.ExposureFull:
		return true
	default:
		return false
	}
}
