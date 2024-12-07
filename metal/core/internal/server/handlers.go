package server

import (
	"encoding/json"
	"net/http"

	core_secure "github.com/wrale/wrale-fleet/metal/core/secure"
	core_thermal "github.com/wrale/wrale-fleet/metal/core/thermal"
	hw_secure "github.com/wrale/wrale-fleet/metal/hw/secure"
	hw_thermal "github.com/wrale/wrale-fleet/metal/hw/thermal"
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

	status, err := s.thermalMgr.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}

// handleThermalPolicy handles getting/updating thermal policy
func (s *Server) handleThermalPolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		policy, err := s.thermalMgr.GetPolicy()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(policy)

	case http.MethodPut:
		var policy hw_thermal.Policy
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.thermalMgr.UpdatePolicy(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	status, err := s.securityMgr.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}

// handleSecurityPolicy handles getting/updating security policy
func (s *Server) handleSecurityPolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		policy, err := s.securityMgr.GetPolicy()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(policy)

	case http.MethodPut:
		var policy hw_secure.Policy
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.securityMgr.UpdatePolicy(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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