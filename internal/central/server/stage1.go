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
	s.logger.Debug("initializing device service",
		zap.String("store_type", "memory"))

	store := memory.New()
	s.device = device.NewService(store, s.logger)

	s.logger.Info("device service initialized successfully")
	return nil
}

// cleanupDeviceService performs cleanup of device service resources.
func (s *Server) cleanupDeviceService(ctx context.Context) error {
	if s.device == nil {
		s.logger.Debug("no device service to clean up")
		return nil
	}

	s.logger.Info("cleaning up device service resources")

	// Cleanup device store if it implements cleanup
	if closer, ok := s.device.Store().(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.logger.Error("failed to close device store", zap.Error(err))
			return fmt.Errorf("closing device store: %w", err)
		}
	}

	s.logger.Info("device service cleanup completed")
	return nil
}

// registerStage1Routes registers HTTP routes for Stage 1 capabilities.
func (s *Server) registerStage1Routes(mux *http.ServeMux) {
	s.logger.Info("registering Stage 1 API routes")

	// Device management endpoints
	mux.HandleFunc("/api/v1/devices", s.handleDevices())
	mux.HandleFunc("/api/v1/devices/", s.handleDeviceByID())

	s.logger.Debug("Stage 1 routes registered",
		zap.Strings("endpoints", []string{
			"/api/v1/devices",
			"/api/v1/devices/",
		}))
}

// handleDevices handles device list and creation requests.
func (s *Server) handleDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.logger.Debug("handling device list request",
				zap.String("remote_addr", r.RemoteAddr))
			// TODO: Implement device listing
			http.Error(w, "not implemented", http.StatusNotImplemented)
		case http.MethodPost:
			s.logger.Debug("handling device creation request",
				zap.String("remote_addr", r.RemoteAddr))
			// TODO: Implement device creation
			http.Error(w, "not implemented", http.StatusNotImplemented)
		default:
			s.logger.Warn("invalid method for devices endpoint",
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleDeviceByID handles requests for specific devices.
func (s *Server) handleDeviceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID := r.URL.Path[len("/api/v1/devices/"):]

		s.logger.Debug("handling device-specific request",
			zap.String("device_id", deviceID),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr))

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
			s.logger.Warn("invalid method for device-specific endpoint",
				zap.String("method", r.Method),
				zap.String("device_id", deviceID),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
