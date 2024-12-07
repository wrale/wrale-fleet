package server

import (
    "net/http"
    "time"
    
    "github.com/gorilla/websocket"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// Device handlers

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
    devices, err := s.deviceSvc.ListDevices()
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "list_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, devices)
}

func (s *Server) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
    var req apitypes.DeviceCreateRequest
    if err := s.parseJSON(r, &req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
        return
    }

    device, err := s.deviceSvc.CreateDevice(&req)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "create_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusCreated, device)
}

func (s *Server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
    deviceID := s.getDeviceID(r)
    
    device, err := s.deviceSvc.GetDevice(deviceID)
    if err != nil {
        s.sendError(w, http.StatusNotFound, "not_found", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, device)
}

func (s *Server) handleUpdateDevice(w http.ResponseWriter, r *http.Request) {
    deviceID := s.getDeviceID(r)

    var req apitypes.DeviceUpdateRequest
    if err := s.parseJSON(r, &req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
        return
    }

    device, err := s.deviceSvc.UpdateDevice(deviceID, &req)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "update_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, device)
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
    deviceID := s.getDeviceID(r)

    if err := s.deviceSvc.DeleteDevice(deviceID); err != nil {
        s.sendError(w, http.StatusInternalServerError, "delete_failed", err.Error())
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeviceCommand(w http.ResponseWriter, r *http.Request) {
    deviceID := s.getDeviceID(r)

    var req apitypes.DeviceCommandRequest
    if err := s.parseJSON(r, &req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
        return
    }

    resp, err := s.deviceSvc.ExecuteCommand(deviceID, &req)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "command_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, resp)
}

// Fleet handlers

func (s *Server) handleFleetCommand(w http.ResponseWriter, r *http.Request) {
    var req apitypes.FleetCommandRequest
    if err := s.parseJSON(r, &req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
        return
    }

    if err := s.fleetSvc.ExecuteFleetCommand(&req); err != nil {
        s.sendError(w, http.StatusInternalServerError, "command_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, map[string]string{
        "status": "accepted",
    })
}

func (s *Server) handleFleetMetrics(w http.ResponseWriter, r *http.Request) {
    metrics, err := s.fleetSvc.GetFleetMetrics()
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "metrics_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, metrics)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
    var req apitypes.ConfigRequest
    if err := s.parseJSON(r, &req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid_request", err.Error())
        return
    }

    if err := s.fleetSvc.UpdateFleetConfig(req.Config); err != nil {
        s.sendError(w, http.StatusInternalServerError, "config_failed", err.Error())
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
    config, err := s.fleetSvc.GetFleetConfig()
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "config_failed", err.Error())
        return
    }

    s.sendJSON(w, http.StatusOK, config)
}

// WebSocket handler

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    // Upgrade connection to WebSocket
    conn, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "ws_upgrade_failed", err.Error())
        return
    }
    defer conn.Close()

    // Add client to WebSocket service
    s.wsSvc.AddClient(conn)
    defer s.wsSvc.RemoveClient(conn)

    // Parse device ID from query param
    if deviceID := r.URL.Query().Get("device"); deviceID != "" {
        updates, err := s.wsSvc.GetDeviceUpdates(types.DeviceID(deviceID))
        if err != nil {
            conn.WriteJSON(map[string]interface{}{
                "type":  "error",
                "error": err.Error(),
            })
            return
        }

        // Forward device updates to WebSocket
        for update := range updates {
            if err := conn.WriteJSON(update); err != nil {
                return
            }
        }
    }

    // Keep connection alive
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            return
        }
    }
}

// Health check handler

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
    s.sendJSON(w, http.StatusOK, map[string]string{
        "status": "healthy",
        "time":   time.Now().Format(time.RFC3339),
    })
}