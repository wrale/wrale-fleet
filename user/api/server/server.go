// Package server implements the API server and handlers
package server

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
    
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// Server represents the API server
type Server struct {
    router     *mux.Router
    upgrader   websocket.Upgrader
    
    // Services
    deviceSvc  apitypes.DeviceService
    fleetSvc   apitypes.FleetService
    wsSvc      apitypes.WebSocketService
    authSvc    apitypes.AuthService
}

// NewServer creates a new API server instance
func NewServer(
    deviceSvc apitypes.DeviceService,
    fleetSvc apitypes.FleetService,
    wsSvc apitypes.WebSocketService,
    authSvc apitypes.AuthService,
) *Server {
    s := &Server{
        router:    mux.NewRouter(),
        deviceSvc: deviceSvc,
        fleetSvc:  fleetSvc,
        wsSvc:     wsSvc,
        authSvc:   authSvc,
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
            CheckOrigin: func(r *http.Request) bool {
                // TODO: Implement proper origin checking for v1.0
                return true
            },
        },
    }

    s.setupRoutes()
    return s
}

// Start starts the API server
func (s *Server) Start(addr string) error {
    srv := &http.Server{
        Handler:      s.router,
        Addr:         addr,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Printf("Starting API server on %s", addr)
    return srv.ListenAndServe()
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
    // API versioning
    api := s.router.PathPrefix("/api/v1").Subrouter()

    // Middleware
    api.Use(s.loggingMiddleware)
    api.Use(s.authMiddleware)

    // Device routes
    api.HandleFunc("/devices", s.handleListDevices).Methods("GET")
    api.HandleFunc("/devices", s.handleCreateDevice).Methods("POST")
    api.HandleFunc("/devices/{id}", s.handleGetDevice).Methods("GET")
    api.HandleFunc("/devices/{id}", s.handleUpdateDevice).Methods("PUT")
    api.HandleFunc("/devices/{id}", s.handleDeleteDevice).Methods("DELETE")
    api.HandleFunc("/devices/{id}/command", s.handleDeviceCommand).Methods("POST")

    // Fleet routes
    api.HandleFunc("/fleet/command", s.handleFleetCommand).Methods("POST")
    api.HandleFunc("/fleet/metrics", s.handleFleetMetrics).Methods("GET")
    api.HandleFunc("/fleet/config", s.handleUpdateConfig).Methods("PUT")
    api.HandleFunc("/fleet/config", s.handleGetConfig).Methods("GET")

    // WebSocket route
    api.HandleFunc("/ws", s.handleWebSocket)

    // Health check
    api.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
}

// Middleware functions

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Call the next handler
        next.ServeHTTP(w, r)

        // Log the request
        log.Printf(
            "%s %s %s %s",
            r.RemoteAddr,
            r.Method,
            r.RequestURI,
            time.Since(start),
        )
    })
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            s.sendError(w, http.StatusUnauthorized, "missing_auth", "No authorization token provided")
            return
        }

        // Authenticate request
        valid, err := s.authSvc.Authenticate(token)
        if err != nil {
            s.sendError(w, http.StatusInternalServerError, "auth_error", "Authentication failed")
            return
        }
        if !valid {
            s.sendError(w, http.StatusUnauthorized, "invalid_auth", "Invalid authorization token")
            return
        }

        // Authorize action
        authorized, err := s.authSvc.Authorize(token, r.URL.Path, r.Method)
        if err != nil {
            s.sendError(w, http.StatusInternalServerError, "auth_error", "Authorization failed")
            return
        }
        if !authorized {
            s.sendError(w, http.StatusForbidden, "unauthorized", "Not authorized for this operation")
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Helper methods

func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    if err := json.NewEncoder(w).Encode(apitypes.APIResponse{
        Success: true,
        Data:    data,
    }); err != nil {
        log.Printf("Error encoding response: %v", err)
    }
}

func (s *Server) sendError(w http.ResponseWriter, status int, code string, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    if err := json.NewEncoder(w).Encode(apitypes.APIResponse{
        Success: false,
        Error: &apitypes.APIError{
            Code:    code,
            Message: message,
        },
    }); err != nil {
        log.Printf("Error encoding error response: %v", err)
    }
}

// Request parsing helpers

func (s *Server) parseJSON(r *http.Request, v interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(v); err != nil {
        return fmt.Errorf("invalid request body: %w", err)
    }
    return nil
}

func (s *Server) getDeviceID(r *http.Request) types.DeviceID {
    return types.DeviceID(mux.Vars(r)["id"])
}