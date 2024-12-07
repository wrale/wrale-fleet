package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wrale/wrale-fleet/user/api/types"
)

// API handlers

func (s *Server) handleGetInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"version":   "v1.0.0",
		"status":    "healthy",
		"timestamp": time.Now(),
	}
	s.sendJSON(w, http.StatusOK, info)
}

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.deviceSvc.List(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "list_devices_failed", err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, devices)
}

func (s *Server) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
	var device types.Device
	if err := s.parseJSON(r, &device); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := s.deviceSvc.Create(r.Context(), &device); err != nil {
		s.sendError(w, http.StatusInternalServerError, "create_device_failed", err.Error())
		return
	}

	s.sendJSON(w, http.StatusCreated, device)
}

func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := s.getDeviceID(r)
	device, err := s.deviceSvc.Get(r.Context(), string(deviceID))
	if err != nil {
		s.sendError(w, http.StatusNotFound, "device_not_found", err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, device)
}

func (s *Server) handleUpdateDevice(w http.ResponseWriter, r *http.Request) {
	var device types.Device
	if err := s.parseJSON(r, &device); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	device.ID = string(s.getDeviceID(r))
	if err := s.deviceSvc.Update(r.Context(), &device); err != nil {
		s.sendError(w, http.StatusInternalServerError, "update_device_failed", err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, device)
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := s.getDeviceID(r)
	if err := s.deviceSvc.Delete(r.Context(), string(deviceID)); err != nil {
		s.sendError(w, http.StatusInternalServerError, "delete_device_failed", err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleDeviceCommand(w http.ResponseWriter, r *http.Request) {
	var cmd types.DeviceCommand
	if err := s.parseJSON(r, &cmd); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	deviceID := s.getDeviceID(r)
	if err := s.deviceSvc.SendCommand(r.Context(), string(deviceID), &cmd); err != nil {
		s.sendError(w, http.StatusInternalServerError, "command_failed", err.Error())
		return
	}

	s.sendJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (s *Server) handleFleetCommand(w http.ResponseWriter, r *http.Request) {
	var cmd types.FleetCommand
	if err := s.parseJSON(r, &cmd); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := s.fleetSvc.SendCommand(r.Context(), &cmd); err != nil {
		s.sendError(w, http.StatusInternalServerError, "command_failed", err.Error())
		return
	}

	s.sendJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (s *Server) handleFleetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.fleetSvc.GetMetrics(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "metrics_failed", err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, metrics)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.fleetSvc.GetConfig(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "get_config_failed", err.Error())
		return
	}
	s.sendJSON(w, http.StatusOK, config)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var config types.FleetConfig
	if err := s.parseJSON(r, &config); err != nil {
		s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := s.fleetSvc.UpdateConfig(r.Context(), &config); err != nil {
		s.sendError(w, http.StatusInternalServerError, "update_config_failed", err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, config)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	if err := s.wsSvc.HandleConnection(conn); err != nil {
		log.Printf("WebSocket handler error: %v", err)
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.sendJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
