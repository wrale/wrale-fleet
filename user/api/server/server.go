// Package server implements the API server and handlers
package server

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    
    "github.com/wrale/wrale-fleet/fleet/types"
    "github.com/wrale/wrale-fleet/user/api/service"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// Config contains server configuration
type Config struct {
    HTTPAddr string // HTTP listen address
}

// Server represents the API server
type Server struct {
    config     Config
    srv        *http.Server
    router     *mux.Router
    
    // Services
    deviceSvc  apitypes.DeviceService
    fleetSvc   apitypes.FleetService
    wsSvc      apitypes.WebSocketService
    authSvc    apitypes.AuthService
}

// New creates a new server instance
func New(cfg Config) (*Server, error) {
    s := &Server{
        config:    cfg,
        router:    mux.NewRouter(),
        deviceSvc: service.NewDeviceService(),
        fleetSvc:  service.NewFleetService(),
        wsSvc:     service.NewWebSocketService(),
        authSvc:   service.NewAuthService(),
    }

    s.setupRoutes()

    s.srv = &http.Server{
        Handler:      s.router,
        Addr:         cfg.HTTPAddr,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    return s, nil
}

// Run starts the server and blocks until context cancellation
func (s *Server) Run(ctx context.Context) error {
    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        log.Printf("Starting API server on %s", s.config.HTTPAddr)
        if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errCh <- fmt.Errorf("server error: %w", err)
        }
    }()

    // Wait for shutdown signal or error
    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        // Shutdown with 5 second timeout
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := s.srv.Shutdown(shutdownCtx); err != nil {
            return fmt.Errorf("error during shutdown: %w", err)
        }
        return nil
    }
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

// getDeviceID extracts device ID from request vars
func (s *Server) getDeviceID(r *http.Request) types.DeviceID {
    return types.DeviceID(mux.Vars(r)["id"])
}
