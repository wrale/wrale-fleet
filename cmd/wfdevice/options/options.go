// Package options provides configuration and initialization for the wfdevice command.
package options

import (
	"fmt"

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

// NewServer creates and configures the server based on the provided options.
func NewServer(opts ...server.Option) (*server.Server, error) {
	// Initialize logger first
	log, err := logger.New(logger.Config{
		Level: "info", // TODO: Make configurable
		Stage: 1,      // Start with Stage 1 capabilities
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

// WithName returns a server option for setting the device name
func WithName(name string) server.Option {
	return server.WithName(name)
}

// WithControlPlane returns a server option for setting the control plane address
func WithControlPlane(addr string) server.Option {
	return server.WithControlPlane(addr)
}

// WithTags returns a server option for setting device tags
func WithTags(tags map[string]string) server.Option {
	return server.WithTags(tags)
}
