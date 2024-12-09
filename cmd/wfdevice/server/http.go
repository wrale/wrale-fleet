package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// routes sets up the HTTP routes for the main API server
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Stage 1 routes for device operations
	mux.HandleFunc("/api/v1/status", s.handleStatus())
	mux.HandleFunc("/api/v1/config", s.handleConfig())
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics())

	return mux
}

// handleStatus handles device status requests
func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		status, err := s.Status(r.Context())
		s.mu.RUnlock()

		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting status: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			s.logger.Error("failed to encode status response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleConfig handles device configuration operations
func (s *Server) handleConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetConfig(w, r)
		case http.MethodPut:
			s.handleUpdateConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleGetConfig retrieves current device configuration
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.device.Config); err != nil {
		s.logger.Error("failed to encode config response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleUpdateConfig applies new device configuration
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newConfig json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "Invalid configuration format", http.StatusBadRequest)
		return
	}

	// Store new configuration
	s.device.Config = newConfig

	s.logger.Info("applied new device configuration", 
		zap.String("name", s.device.Name),
		zap.Int("config_size", len(newConfig)))

	w.WriteHeader(http.StatusOK)
}

// handleMetrics handles device metric collection
func (s *Server) handleMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		// Basic metric collection in Stage 1
		metrics := map[string]interface{}{
			"status":     s.device.Status,
			"registered": s.registered,
			"uptime":     time.Since(s.startTime).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			s.logger.Error("failed to encode metrics response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
