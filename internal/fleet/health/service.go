package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Service provides health monitoring operations with multi-tenant isolation.
type Service struct {
	store    Store
	logger   *zap.Logger
	mu       sync.RWMutex
	checkers map[string]HealthChecker
}

// NewService creates a new health monitoring service with the provided
// storage backend and logger.
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:    store,
		logger:   logger,
		checkers: make(map[string]HealthChecker),
	}
}

// RegisterComponent adds a new component to the health monitoring system
func (s *Service) RegisterComponent(ctx context.Context, name string, checker HealthChecker, info ComponentInfo, opts ...Option) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply options
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return fmt.Errorf("invalid option: %w", err)
		}
	}

	// Register the component
	s.checkers[name] = checker
	if err := s.store.RegisterComponent(ctx, name, info); err != nil {
		delete(s.checkers, name)
		return fmt.Errorf("failed to register component: %w", err)
	}

	s.logger.Info("registered component for health monitoring",
		zap.String("component", name),
		zap.String("category", info.Category),
		zap.Bool("critical", info.Critical),
	)

	return nil
}

// UnregisterComponent removes a component from health monitoring
func (s *Service) UnregisterComponent(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.checkers, name)
	if err := s.store.UnregisterComponent(ctx, name); err != nil {
		return fmt.Errorf("failed to unregister component: %w", err)
	}

	s.logger.Info("unregistered component from health monitoring",
		zap.String("component", name),
	)

	return nil
}

// CheckHealth performs health checks on registered components and updates their status
func (s *Service) CheckHealth(ctx context.Context, opts ...Option) (*HealthResponse, error) {
	// Apply options
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}

	// Create context with timeout if specified
	if options.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.timeout)
		defer cancel()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check each component
	for name, checker := range s.checkers {
		err := checker.CheckHealth(ctx)
		status := &HealthStatus{
			TenantID:    options.tenantID,
			LastChecked: time.Now().UTC(),
		}

		if err != nil {
			status.Status = StatusUnhealthy
			status.Message = "Health check failed"
			status.LastError = err.Error()
			s.logger.Warn("component health check failed",
				zap.String("component", name),
				zap.Error(err),
			)
		} else {
			status.Status = StatusHealthy
			status.Message = "Component operational"
		}

		if err := s.store.UpdateComponentStatus(ctx, name, status); err != nil {
			return nil, fmt.Errorf("failed to update component status: %w", err)
		}
	}

	// Get current status from store
	components, err := s.store.ListComponentStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list component statuses: %w", err)
	}

	ready, err := s.store.GetReadyStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ready status: %w", err)
	}

	// Calculate overall status
	status := StatusHealthy
	for _, health := range components {
		if health.Status == StatusUnhealthy {
			status = StatusUnhealthy
			break
		} else if health.Status == StatusDegraded {
			status = StatusDegraded
		}
	}

	return &HealthResponse{
		TenantID:    options.tenantID,
		Status:      status,
		Ready:       ready,
		Components:  components,
		LastChecked: time.Now().UTC(),
	}, nil
}

// SetReady marks the system as ready to serve requests
func (s *Service) SetReady(ctx context.Context, ready bool, opts ...Option) error {
	// Apply options
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return fmt.Errorf("invalid option: %w", err)
		}
	}

	if err := s.store.SetReadyStatus(ctx, ready); err != nil {
		return fmt.Errorf("failed to set ready status: %w", err)
	}

	s.logger.Info("updated system ready status",
		zap.Bool("ready", ready),
		zap.String("tenant_id", options.tenantID),
	)

	return nil
}

// IsReady returns whether the system is ready to serve requests
func (s *Service) IsReady(ctx context.Context, opts ...Option) (bool, error) {
	// Apply options
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return false, fmt.Errorf("invalid option: %w", err)
		}
	}

	return s.store.GetReadyStatus(ctx)
}

// Store returns the health store instance.
// This provides controlled access to the underlying storage implementation
// while maintaining proper encapsulation.
func (s *Service) Store() Store {
	return s.store
}
