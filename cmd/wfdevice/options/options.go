// Package options provides configuration and initialization for the wfdevice command.
package options

import (
	"fmt"
	"os"

	"github.com/wrale/wrale-fleet/cmd/wfdevice/logger"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/server"
)

// Config holds the command-line options for wfdevice.
type Config struct {
	Port         string
	DataDir      string
	LogLevel     string
	Name         string
	ControlPlane string
	Tags         map[string]string
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{
		Tags: make(map[string]string),
	}
}

// Validate checks that all required configuration options are present
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("port is required")
	}
	if c.DataDir == "" {
		return fmt.Errorf("data directory is required")
	}
	if c.ControlPlane == "" {
		return fmt.Errorf("control plane address is required")
	}
	return nil
}

// NewServer creates and configures the server based on the provided options.
// Instead of accepting raw server.Options, it now accepts a Config to ensure
// proper validation and option creation.
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

	// Now that we have a logger, we can log configuration details
	log.Info("creating server with configuration",
		logger.String("port", cfg.Port),
		logger.String("data_dir", cfg.DataDir),
		logger.String("name", cfg.Name),
		logger.String("control_plane", cfg.ControlPlane),
	)

	// Create server options from validated config
	var opts []server.Option
	opts = append(opts,
		server.WithPort(cfg.Port),
		server.WithDataDir(cfg.DataDir),
		server.WithControlPlane(cfg.ControlPlane),
	)

	// Add optional configurations
	if cfg.Name != "" {
		opts = append(opts, server.WithName(cfg.Name))
	}
	if len(cfg.Tags) > 0 {
		opts = append(opts, server.WithTags(cfg.Tags))
	}

	// Create server with assembled options
	srv, err := server.New(log, opts...)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}

// The following functions are deprecated and will be removed.
// They are kept temporarily for backward compatibility.
// Use NewServer with Config instead.

// WithPort returns a server option for setting the port
// Deprecated: Use NewServer with Config
func WithPort(port string) server.Option {
	return server.WithPort(port)
}

// WithDataDir returns a server option for setting the data directory
// Deprecated: Use NewServer with Config
func WithDataDir(dir string) server.Option {
	return server.WithDataDir(dir)
}

// WithName returns a server option for setting the device name
// Deprecated: Use NewServer with Config
func WithName(name string) server.Option {
	return server.WithName(name)
}

// WithControlPlane returns a server option for setting the control plane address
// Deprecated: Use NewServer with Config
func WithControlPlane(addr string) server.Option {
	return server.WithControlPlane(addr)
}

// WithTags returns a server option for setting device tags
// Deprecated: Use NewServer with Config
func WithTags(tags map[string]string) server.Option {
	return server.WithTags(tags)
}
