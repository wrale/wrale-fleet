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
	
	// Core managers
	thermalMgr  *core_thermal.PolicyManager
	securityMgr *core_secure.PolicyManager
	
	// State
	mu       sync.RWMutex
	stopping bool
}

// New creates a new server instance
func New(cfg Config) (*Server, error) {
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	// Initialize hardware level components
	hwThermal, err := core_thermal.NewHardwareMonitor()
	if err != nil {
		return nil, fmt.Errorf("failed to create thermal monitor: %w", err)
	}

	hwSecurity, err := core_secure.NewHardwareMonitor()
	if err != nil {
		return nil, fmt.Errorf("failed to create security monitor: %w", err)
	}

	// Initialize core policy managers
	thermalMgr := core_thermal.NewPolicyManager(cfg.DeviceID, hwThermal.Monitor(), core_thermal.DefaultPolicy())
	securityMgr := core_secure.NewPolicyManager(cfg.DeviceID, hwSecurity, core_secure.DefaultPolicy())

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