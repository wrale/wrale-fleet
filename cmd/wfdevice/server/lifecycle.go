package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"go.uber.org/zap"
)

// Run starts the server and blocks until the context is canceled
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting wfdevice agent",
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
		zap.Uint8("stage", uint8(s.stage)),
	)

	// Write PID file
	if err := s.writePIDFile(); err != nil {
		return fmt.Errorf("writing pid file: %w", err)
	}
	defer s.removePIDFile()

	// Initialize HTTP server with security timeouts
	s.httpSrv = &http.Server{
		Addr:              ":" + s.cfg.Port,
		Handler:           s.routes(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Start HTTP server
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server",
			zap.String("addr", s.httpSrv.Addr),
			zap.Duration("header_timeout", readHeaderTimeout),
		)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	}()

	// Register with control plane if name is provided
	if s.cfg.Name != "" {
		regCtx, cancel := context.WithTimeout(ctx, registrationTimeout)
		defer cancel()

		if err := s.register(regCtx); err != nil {
			return fmt.Errorf("device registration failed: %w", err)
		}

		// Start health reporting after successful registration
		s.startHealthReporting()
	} else {
		s.logger.Info("device name not provided, skipping registration",
			zap.String("status", string(s.device.Status)))
	}

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		s.logger.Info("shutting down agent")
		return s.shutdown()
	case err := <-errChan:
		return fmt.Errorf("agent error: %w", err)
	}
}

// Stop initiates a graceful shutdown of the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping server")
	return s.shutdown()
}

// Status returns the current device status
func (s *Server) Status(ctx context.Context) (*DeviceStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &DeviceStatus{
		Name:         s.cfg.Name,
		Status:       s.device.Status,
		Tags:         s.device.Tags,
		ControlPlane: s.cfg.ControlPlane,
		Registered:   s.registered,
	}, nil
}

// NotifyShutdown informs the control plane of a planned shutdown
func (s *Server) NotifyShutdown(ctx context.Context) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.registered {
		return fmt.Errorf("device not registered with control plane")
	}

	s.notifyShutdown()
	return nil
}

// shutdown performs a graceful server shutdown
func (s *Server) shutdown() error {
	// Stop health reporting
	if s.stopHealth != nil {
		close(s.stopHealth)
	}

	// Notify control plane of shutdown if registered
	s.mu.RLock()
	if s.registered {
		s.notifyShutdown()
	}
	s.mu.RUnlock()

	// Shutdown HTTP server
	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
	}

	// Remove PID file
	s.removePIDFile()

	return nil
}

// register handles device registration with the control plane
func (s *Server) register(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate registration requirements
	if s.cfg.Name == "" {
		return fmt.Errorf("device name is required for registration")
	}

	s.logger.Info("registering device with control plane",
		zap.String("name", s.cfg.Name),
		zap.String("control_plane", s.cfg.ControlPlane),
	)

	// Update device identity now that we have the name
	s.device.Name = s.cfg.Name

	// TODO: Implement actual registration logic with control plane
	time.Sleep(time.Second)

	s.registered = true
	s.device.Status = device.StatusOnline

	s.logger.Info("device registration successful")
	return nil
}

// notifyShutdown informs the control plane of planned shutdown
func (s *Server) notifyShutdown() {
	s.logger.Info("notifying control plane of shutdown")
	// TODO: Implement shutdown notification to control plane
}

// startHealthReporting begins periodic health check submissions
func (s *Server) startHealthReporting() {
	go func() {
		ticker := time.NewTicker(healthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.submitHealthReport(); err != nil {
					s.logger.Error("failed to submit health report", zap.Error(err))
				}
			case <-s.stopHealth:
				return
			}
		}
	}()
}

// submitHealthReport sends a health report to the control plane
func (s *Server) submitHealthReport() error {
	s.logger.Debug("submitting health report")
	// TODO: Implement health report submission to control plane
	return nil
}
