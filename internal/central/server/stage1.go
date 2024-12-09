package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// initStage1 initializes Stage 1 capabilities.
func (s *Server) initStage1() error {
	s.logger.Info("initializing Stage 1 capabilities")

	// Verify core services are ready
	if s.device == nil {
		return fmt.Errorf("device service not initialized")
	}

	// Stage-specific initialization can be added here
	return nil
}

// cleanupDeviceService performs cleanup of device service resources.
// This is called during server shutdown to ensure proper resource release.
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
// This sets up the basic REST API endpoints for device management.
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
// This implements the collection endpoints for device management:
// - GET: List devices with optional filtering
// - POST: Create new devices
func (s *Server) handleDevices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract tenant ID from context
		tenantID, err := device.TenantFromContext(ctx)
		if err != nil {
			s.logger.Error("failed to get tenant from context",
				zap.Error(err),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.logger.Debug("handling device list request",
				zap.String("tenant_id", tenantID),
				zap.String("remote_addr", r.RemoteAddr))

			devices, err := s.device.List(ctx, device.ListOptions{
				TenantID: tenantID,
			})
			if err != nil {
				s.logger.Error("failed to list devices",
					zap.Error(err),
					zap.String("tenant_id", tenantID),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// Return devices as JSON response
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"devices": devices,
			}); err != nil {
				s.logger.Error("failed to encode device list response",
					zap.Error(err),
					zap.String("tenant_id", tenantID),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

		case http.MethodPost:
			s.logger.Debug("handling device creation request",
				zap.String("tenant_id", tenantID),
				zap.String("remote_addr", r.RemoteAddr))

			// TODO: Parse device creation request
			// TODO: Validate device data
			// TODO: Create device in store
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
// This implements the instance endpoints for device management:
// - GET: Retrieve device details
// - PUT: Update device configuration
// - DELETE: Remove device from management
func (s *Server) handleDeviceByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deviceID := r.URL.Path[len("/api/v1/devices/"):]

		// Extract tenant ID from context
		tenantID, err := device.TenantFromContext(ctx)
		if err != nil {
			s.logger.Error("failed to get tenant from context",
				zap.Error(err),
				zap.String("device_id", deviceID),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		s.logger.Debug("handling device-specific request",
			zap.String("device_id", deviceID),
			zap.String("tenant_id", tenantID),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr))

		switch r.Method {
		case http.MethodGet:
			dev, err := s.device.Get(ctx, tenantID, deviceID)
			if err != nil {
				s.logger.Error("failed to get device",
					zap.Error(err),
					zap.String("device_id", deviceID),
					zap.String("tenant_id", tenantID),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// Return device as JSON response
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"device": dev,
			}); err != nil {
				s.logger.Error("failed to encode device response",
					zap.Error(err),
					zap.String("device_id", deviceID),
					zap.String("tenant_id", tenantID),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

		case http.MethodPut:
			// TODO: Parse device update request
			// TODO: Validate updated device data
			// TODO: Update device in store
			http.Error(w, "not implemented", http.StatusNotImplemented)

		case http.MethodDelete:
			// TODO: Implement device deletion with proper cleanup
			http.Error(w, "not implemented", http.StatusNotImplemented)

		default:
			s.logger.Warn("invalid method for device-specific endpoint",
				zap.String("method", r.Method),
				zap.String("device_id", deviceID),
				zap.String("tenant_id", tenantID),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
