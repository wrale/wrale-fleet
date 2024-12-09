package server

import (
	"context"
	"fmt"

	"github.com/wrale/wrale-fleet/internal/fleet/device"
	"github.com/wrale/wrale-fleet/internal/fleet/device/store/memory"
	"github.com/wrale/wrale-fleet/internal/fleet/health"
	healthmem "github.com/wrale/wrale-fleet/internal/fleet/health/store/memory"
	"go.uber.org/zap"
)

// initialize sets up all server components in the proper sequence.
// The initialization order is critical for proper dependency management:
// 1. Core services (device, etc.)
// 2. Health monitoring system
// 3. Stage-specific capabilities
func (s *Server) initialize() error {
	s.logger.Info("initializing central control plane server",
		zap.String("port", s.cfg.Port),
		zap.String("data_dir", s.cfg.DataDir),
		zap.Uint8("stage", uint8(s.stage)),
	)

	// First initialize core services
	if err := s.initCoreServices(); err != nil {
		return fmt.Errorf("core services initialization failed: %w", err)
	}

	// Next initialize health monitoring
	if err := s.initHealthSystem(); err != nil {
		return fmt.Errorf("health system initialization failed: %w", err)
	}

	// Initialize stage-specific capabilities
	if err := s.initStage1(); err != nil {
		return fmt.Errorf("stage 1 initialization failed: %w", err)
	}

	return nil
}

// initCoreServices initializes the fundamental services required by the system.
func (s *Server) initCoreServices() error {
	// Initialize device service
	s.logger.Info("initializing core services")
	store := memory.New()
	s.device = device.NewService(store, s.logger)

	return nil
}

// initHealthSystem initializes the health monitoring system.
func (s *Server) initHealthSystem() error {
	s.logger.Info("initializing health monitoring system")

	// Create health service with memory store
	healthStore := healthmem.New()
	s.health = health.NewService(healthStore, s.logger)

	// Register base components for health monitoring
	if err := s.registerHealthChecks(); err != nil {
		return fmt.Errorf("health check registration failed: %w", err)
	}

	// Start periodic health check goroutine
	go s.runHealthChecks(s.baseCtx)

	return nil
}

// cleanupDeviceService performs cleanup of device service resources.
// This is called during server shutdown to ensure proper resource release.
func (s *Server) cleanupDeviceService(ctx context.Context) error {
	if s.device == nil {
		s.logger.Debug("no device service to clean up")
		return nil
	}

	s.logger.Info("cleaning up device service resources")

	// Cleanup device store if it implements cleanup
	if closer, ok := s.device.Store().(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			s.logger.Error("failed to close device store", zap.Error(err))
			return fmt.Errorf("closing device store: %w", err)
		}
	}

	s.logger.Info("device service cleanup completed")
	return nil
}
