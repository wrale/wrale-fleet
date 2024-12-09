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
	Port     string // HTTP server port
	DataDir  string // Data directory path
	LogLevel string // Logging level (debug, info, warn, error)
}

// New creates a new Config with default values.
// This provides sensible defaults for command-line flags.
func New() *Config {
	return &Config{
		Port:     "8080",               // Default port for wfcentral
		DataDir:  "/var/lib/wfcentral", // Default data directory
		LogLevel: "info",               // Default log level
	}
}

// NewServer creates and configures a central server instance based on
// the provided configuration options. It handles the initialization of
// all necessary components including logging and monitoring.
func NewServer(cfg *Config) (*server.Server, error) {
	// Initialize logger first with configured level
	log, err := logger.New(logger.Config{
		Level: cfg.LogLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("initializing logger: %w", err)
	}

	// Create internal server configuration
	serverConfig := &server.Config{
		Port:     cfg.Port,
		DataDir:  cfg.DataDir,
		LogLevel: cfg.LogLevel,
	}

	// Create server instance with configured options
	srv, err := server.New(serverConfig, log)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}
