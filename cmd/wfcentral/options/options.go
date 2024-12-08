// Package options provides configuration and initialization for the wfcentral command.
package options

import (
	"fmt"

	"github.com/joshuapare/wrale-fleet/cmd/wfcentral/logger"
	"github.com/joshuapare/wrale-fleet/cmd/wfcentral/server"
)

// Config holds the command-line options for wfcentral.
type Config struct {
	Port     string
	DataDir  string
	LogLevel string
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{}
}

// NewServer creates and configures the server based on the provided options.
func NewServer(opts ...server.Option) (*server.Server, error) {
	// Initialize logger first
	log, err := logger.New(logger.Config{
		Level: "info", // TODO: Make configurable
	})
	if err != nil {
		return nil, fmt.Errorf("initializing logger: %w", err)
	}

	// Create server with options
	srv, err := server.New(log, opts...)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}

// WithPort returns a server option for setting the port
func WithPort(port string) server.Option {
	return server.WithPort(port)
}

// WithDataDir returns a server option for setting the data directory
func WithDataDir(dir string) server.Option {
	return server.WithDataDir(dir)
}
