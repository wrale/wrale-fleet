package server

import (
	"encoding/json"
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
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}
}

// handleStatus handles status check requests
func (s *Server) handleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		status, err := s.Status(r.Context())
		s.mu.RUnlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
