package server

import (
	"encoding/json"
	"net/http"
)

// Device information response
type deviceInfo struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// handleGetInfo returns basic device information
func (s *Server) handleGetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := deviceInfo{
		ID:      s.config.DeviceID,
		Version: s.version(),
	}

	json.NewEncoder(w).Encode(info)
}

// handleGetThermalStatus returns current thermal state
func (s *Server) handleGetThermalStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.thermalMgr.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleThermalPolicy handles getting/updating thermal policy
func (s *Server) handleThermalPolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(s.thermalMgr.GetPolicy())

	case http.MethodPut:
		var policy interface{}
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.thermalMgr.UpdatePolicy(policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSecurityStatus returns current security state
func (s *Server) handleGetSecurityStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.securityMgr.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// handleSecurityPolicy handles getting/updating security policy
func (s *Server) handleSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(s.securityMgr.GetPolicy())

	case http.MethodPut:
		var policy interface{}
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.securityMgr.UpdatePolicy(policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// version returns the server version
func (s *Server) version() string {
	// TODO: Get version from build info
	return "1.0.0"
}