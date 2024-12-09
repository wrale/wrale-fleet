package server

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/device/store/memory"
)

// initStage1 initializes Stage 1 capabilities.
func (s *Server) initStage1() error {
	s.logger.Info("initializing Stage 1 capabilities")

	// Initialize device service
	if err := s.initDeviceService(); err != nil {
		return fmt.Errorf("device service initialization failed: %w", err)
	}

	// Perform initial health check
	if err := s.checkDeviceServiceHealth(context.Background()); err != nil {
		return fmt.Errorf("initial health check failed: %w", err)
	}

	return nil
}

// initDeviceService initializes the device management service.
func (s *Server) initDeviceService() error {
	store := memory.New()
	s.device = device.NewService(store, s.logger)
	return nil
}

// cleanupDeviceService performs cleanup of device service resources.
func (s *Server) cleanupDeviceService(ctx context.Context) error {
	if s.device == nil {
		return nil
	}

	// Cleanup device store if it implements cleanup
	if closer, ok := s.device.Store().(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("closing device store: %w", err)
		}
	}

	return nil
}

// checkDeviceServiceHealth verifies device service health.
func (s *Server) checkDeviceServiceHealth(ctx context.Context) error {
	if s.device == nil {
		return fmt.Errorf("device service not initialized")
	}

	// Check if store is accessible by performing a no-op list
	if _, err := s.device.List(ctx, device.ListOptions{}); err != nil {
		return fmt.Errorf("device store access check failed: %w", err)
	}

	// Additional health checks specific to Stage 1 can be added here
	// as requirements evolve

	return nil
}

// registerStage1Routes registers HTTP routes for Stage 1 capabilities.
func (s *Server) registerStage1Routes(mux *http.ServeMux) {
	// Device management endpoints
	mux.HandleFunc("/api/v1/devices", s.handleDevices())
	mux.HandleFunc("/api/v1/devices/", s.handleDeviceByID())
}

// handleDevices handles device list and creation requests.
func (s *Server) handleDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// TODO: Implement device listing
			http.Error(w, "not implemented", http.StatusNotImplemented)
		case http.MethodPost:
			// TODO: Implement device creation
			http.Error(w, "not implemented", http.StatusNotImplemented)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleDeviceByID handles requests for specific devices.
func (s *Server) handleDeviceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// TODO: Implement device retrieval
			http.Error(w, "not implemented", http.StatusNotImplemented)
		case http.MethodPut:
			// TODO: Implement device update
			http.Error(w, "not implemented", http.StatusNotImplemented)
		case http.MethodDelete:
			// TODO: Implement device deletion
			http.Error(w, "not implemented", http.StatusNotImplemented)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
