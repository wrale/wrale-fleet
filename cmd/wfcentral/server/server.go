// Package server implements the core wfcentral server functionality.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap"
)

// Stage represents the server's operational stage/capability level
type Stage uint8

const (
	// Stage1 provides basic device management capabilities
	Stage1 Stage = 1
	// Future stages will be added here
)

const (
	// readHeaderTimeout defines the amount of time allowed to read
	// request headers. This helps prevent Slowloris DoS attacks.
	readHeaderTimeout = 10 * time.Second
)

// Server represents the wfcentral server instance
type Server struct {
	cfg     *Config
	logger  *zap.Logger
	stage   Stage
	device  *device.Service
	httpSrv *http.Server
}

// Config holds the server configuration
type Config struct {
	Port     string
	DataDir  string
	LogLevel string
}

// Option defines a server option
type Option func(*Server) error

// New creates a new server instance with the given options
func New(logger *zap.Logger, opts ...Option) (*Server, error) {
	s := &Server{
		cfg:    &Config{},
		logger: logger,
		stage:  Stage1,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("applying server option: %w", err)
		}
	}

	// Initialize device service
	store := memory.New() // Will be replaced with persistent store
	s.device = device.NewService(store, logger)

	return s, nil
}

// WithPort sets the server port
func WithPort(port string) Option {
	return func(s *Server) error {
		s.cfg.Port = port
		return nil
	}
}

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(s *Server) error {
		s.cfg.DataDir = dir
		return nil
	}
}

// Run starts the server and blocks until the context is canceled
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting wfcentral server",
		zap.String("port", s.cfg.Port),
		zap.String("data_dir", s.cfg.DataDir),
		zap.Uint8("stage", uint8(s.stage)),
	)

	// Initialize HTTP server with security timeouts
	s.httpSrv = &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           s.routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Start HTTP server
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server",
			zap.String("addr", s.httpSrv.Addr),
			zap.Duration("header_timeout", readHeaderTimeout),
		)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		return s.shutdown()
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// shutdown performs a graceful server shutdown
func (s *Server) shutdown() error {
	if err := s.httpSrv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	// Close device store if it implements io.Closer
	if closer, ok := s.device.Store().(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("closing device store: %w", err)
		}
	}

	return nil
}

// routes sets up the HTTP routes
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Stage 1 routes
	mux.HandleFunc("/healthz", s.handleHealth())
	mux.HandleFunc("/api/v1/devices", s.handleDevices())

	return mux
}

// Basic health check handler
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy"}`)
	}
}

// Basic device handler stub
func (s *Server) handleDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement device management endpoints
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, `{"error":"not implemented"}`)
	}
}
