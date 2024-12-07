// Package server implements the metal daemon's HTTP API server
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/wrale/wrale-fleet/metal/core/policy"
)

// Config contains server configuration options
type Config struct {
	DeviceID     string // Unique identifier for this device
	HTTPAddr     string // HTTP API listen address
	ThermalMgr   policy.Manager
	SecurityMgr  policy.Manager
}

// Server implements the metal daemon's HTTP API
type Server struct {
	config     Config
	srv        *http.Server
	
	// Core managers
	thermalMgr  policy.Manager
	securityMgr policy.Manager
	
	// State
	mu       sync.RWMutex
	stopping bool
}

// New creates a new server instance
func New(cfg Config) (*Server, error) {
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	if cfg.ThermalMgr == nil {
		return nil, fmt.Errorf("thermal manager is required")
	}

	if cfg.SecurityMgr == nil {
		return nil, fmt.Errorf("security manager is required")
	}

	s := &Server{
		config:      cfg,
		thermalMgr:  cfg.ThermalMgr,
		securityMgr: cfg.SecurityMgr,
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
	// Start policy managers
	if err := s.thermalMgr.Start(); err != nil {
		return fmt.Errorf("failed to start thermal manager: %w", err)
	}
	defer s.thermalMgr.Stop()

	if err := s.securityMgr.Start(); err != nil {
		return fmt.Errorf("failed to start security manager: %w", err)
	}
	defer s.securityMgr.Stop()

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

		return nil
	}
}