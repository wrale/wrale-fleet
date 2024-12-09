// Package server implements the central control plane server functionality
package server

import (
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"go.uber.org/zap"
)

// Server represents the control plane server instance
type Server struct {
	logger         *zap.Logger
	loggingService *logging.Service
	// other fields...
}

// Config holds server configuration
type Config struct {
	Port             string
	DataDir          string
	LogLevel         string
	ManagementPort   string
	LoggingService   *logging.Service
	ManagementConfig *ManagementConfig
}

// New creates a new server instance
func New(cfg *Config, logger *zap.Logger) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if cfg.LoggingService == nil {
		return nil, fmt.Errorf("logging service is required")
	}

	s := &Server{
		logger:         logger,
		loggingService: cfg.LoggingService,
	}

	return s, nil
}
