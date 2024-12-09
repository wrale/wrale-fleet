// Package server implements the wfcentral command server initialization.
package server

import (
	"fmt"

	"github.com/wrale/wrale-fleet/internal/central/server"
	"go.uber.org/zap"
)

// Config holds the server configuration parsed from command-line flags
type Config struct {
	Port     string
	DataDir  string
	LogLevel string
}

// Option defines a server option
type Option func(*Config)

// WithPort sets the server port
func WithPort(port string) Option {
	return func(cfg *Config) {
		cfg.Port = port
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(cfg *Config) {
		cfg.DataDir = dir
	}
}

// WithLogLevel sets the logging level
func WithLogLevel(level string) Option {
	return func(cfg *Config) {
		cfg.LogLevel = level
	}
}

// New creates a new server instance with the given options
func New(logger *zap.Logger, opts ...Option) (*server.Server, error) {
	// Apply options to config
	cfg := &Config{
		Port:     "8080",
		DataDir:  "/var/lib/wfcentral",
		LogLevel: "info",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Create internal server config
	serverConfig := &server.Config{
		Port:     cfg.Port,
		DataDir:  cfg.DataDir,
		LogLevel: cfg.LogLevel,
	}

	// Create and return internal server instance
	srv, err := server.New(serverConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("creating internal server: %w", err)
	}

	return srv, nil
}
