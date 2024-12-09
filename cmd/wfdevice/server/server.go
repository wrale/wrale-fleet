// Package server implements the device agent server functionality
package server

import (
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"go.uber.org/zap"
)

// Server represents the device agent server instance
type Server struct {
	logger         *zap.Logger
	loggingService *logging.Service
	// other fields...
}

// Option is a functional option for configuring the server
type Option func(*Server) error

// WithLogging sets the logging service
func WithLogging(svc *logging.Service) Option {
	return func(s *Server) error {
		s.loggingService = svc
		return nil
	}
}

// New creates a new server instance
func New(logger *zap.Logger, opts ...Option) (*Server, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	s := &Server{
		logger: logger,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	// Validate required components
	if s.loggingService == nil {
		return nil, fmt.Errorf("logging service is required")
	}

	return s, nil
}
