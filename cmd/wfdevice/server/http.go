package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// routes sets up the HTTP routes
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Stage 1 routes
	mux.HandleFunc("/healthz", s.handleHealth())
	mux.HandleFunc("/api/v1/status", s.handleStatus())

	return mux
}

// handleHealth handles basic health check requests
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		}); err != nil {
			s.logger.Error("failed to encode health response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleStatus handles status check requests
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
