// Package server implements the metal daemon's HTTP API server
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	core_secure "github.com/wrale/wrale-fleet/metal/core/secure"
	core_thermal "github.com/wrale/wrale-fleet/metal/core/thermal"
	hw_secure "github.com/wrale/wrale-fleet/metal/hw/secure"
	hw_thermal "github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// Config contains server configuration options
type Config struct {
	DeviceID string // Unique identifier for this device
	HTTPAddr string // HTTP API listen address
}

// Server implements the metal daemon's HTTP API
type Server struct {
	config Config
	srv    *http.Server
	
	// Managers for different subsystems
	thermalMgr  *hw_thermal.Manager
	securityMgr *hw_secure.Manager
	
	// State
	mu       sync.RWMutex
	stopping bool
}

// New creates a new server instance
func New(cfg Config) (*Server, error) {
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	// Initialize managers
	thermalMgr, err := hw_thermal.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create thermal manager: %w", err)
	}

	securityMgr, err := hw_secure.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}

	s := &Server{
		config:     cfg,
		thermalMgr: thermalMgr,
		securityMgr: securityMgr,
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	
	// Device info endpoints
	mux.HandleFunc("/api/v1/info", s.handleGetInfo)
	
	// Thermal management endpoints
	mux.HandleFunc("/api/v1/thermal/status", s.handleGetThermalStatus)
	mux.HandleFunc("/api/v1/thermal/policy", s.handleThermalPolicy)
	
	// Security endpoints
	mux.HandleFunc("/api/v1/secure/status", s.handleGetSecurityStatus)
	mux.HandleFunc("/api/v1/secure/policy", s.handleSecurityPolicy)

	s.srv = &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	return s, nil
}

// Run starts the server and blocks until context cancellation
func (s *Server) Run(ctx context.Context) error {
	// Start subsystem managers
	if err := s.thermalMgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start thermal manager: %w", err)
	}
	if err := s.securityMgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start security manager: %w", err)
	}

	// Start HTTP server
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting HTTP server on %s", s.config.HTTPAddr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server error: %w", err)
		}
	}()

	// Wait for shutdown
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.mu.Lock()
		s.stopping = true
		s.mu.Unlock()

		// Shutdown HTTP server
		if err := s.srv.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}

		// Stop managers
		s.thermalMgr.Stop()
		s.securityMgr.Stop()

		return nil
	}
}