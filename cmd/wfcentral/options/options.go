// Package options provides configuration and initialization for the wfcentral command.
package options

import (
	"fmt"

	"github.com/wrale/wrale-fleet/cmd/wfcentral/logger"
	"github.com/wrale/wrale-fleet/internal/central/server"
)

// Config holds the command-line options for wfcentral.
// This separates command-line concerns from the core server configuration.
type Config struct {
	// Port is the main HTTP server port
	Port string

	// DataDir is the path for persistent storage
	DataDir string

	// LogLevel controls logging verbosity
	LogLevel string

	// ManagementPort is the port for health and readiness endpoints
	// If not specified, defaults to Port + 1
	ManagementPort string

	// HealthExposure controls how much information is exposed in health endpoints
	// Valid values are: "minimal", "standard", "full"
	// - minimal: Only basic health status
	// - standard: Includes version and uptime (default)
	// - full: All available health information
	HealthExposure string
}

// New creates a new Config with sensible default values that prioritize security
// while maintaining backward compatibility. The defaults are chosen to align with
// enterprise deployment patterns where health endpoints are typically accessed
// through internal networks or orchestration systems.
func New() *Config {
	return &Config{
		Port:           "8080",               // Default main server port
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
	// Initialize logger first to ensure proper diagnostics during startup
	log, err := logger.New(logger.Config{
		Level: cfg.LogLevel,
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
			Port:          cfg.ManagementPort, // Will be defaulted if empty
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

// GetDefaultManagementPort returns the default management port for a given main port.
// This is useful for CLI help text and documentation to show users the default value
// that will be used if they don't specify a management port.
func GetDefaultManagementPort(mainPort string) string {
	// The actual defaulting logic is handled in server.Config.Validate()
	// This is just for user information
	return fmt.Sprintf("%s [Main port + 1]", mainPort)
}
