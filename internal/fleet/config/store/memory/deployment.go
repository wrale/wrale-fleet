package memory

import (
	"context"
	"sort"
	"time"

	"github.com/wrale/fleet/internal/fleet/config"
)

// validateDeployment ensures the deployment is valid before storage operations.
// It checks for required fields and proper initialization.
func (s *Store) validateDeployment(deployment *config.Deployment) error {
	if deployment == nil {
		return config.NewError("validate deployment", config.ErrInvalidDeployment, "deployment is nil")
	}
	return s.validateInput("validate deployment", map[string]string{
		"deployment ID": deployment.ID,
		"tenant ID":     deployment.TenantID,
		"device ID":     deployment.DeviceID,
	})
}

// filterDeployments applies filtering criteria to deployments based on the provided options.
// It filters by tenant ID, device ID, and status if specified in the options.
func (s *Store) filterDeployments(deployments []*config.Deployment, opts config.ListOptions) []*config.Deployment {
	filtered := make([]*config.Deployment, 0, len(deployments))

	// Apply all filters in one pass
	for _, d := range deployments {
		matches := true

		// Apply tenant filter if specified
		if opts.TenantID != "" && d.TenantID != opts.TenantID {
			matches = false
		}

		// Apply device filter if specified
		if matches && opts.DeviceID != "" && d.DeviceID != opts.DeviceID {
			matches = false
		}

		// Apply status filter if specified
		if matches && opts.Status != "" && d.Status != opts.Status {
			matches = false
		}

		// Only append if all filters pass
		if matches {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// sortDeployments sorts deployments by deployment time and ID for consistent ordering.
// This ensures deterministic results when listing deployments.
func (s *Store) sortDeployments(deployments []*config.Deployment) {
	sort.Slice(deployments, func(i, j int) bool {
		if deployments[i].DeployedAt.Equal(deployments[j].DeployedAt) {
			return deployments[i].ID < deployments[j].ID
		}
		return deployments[i].DeployedAt.Before(deployments[j].DeployedAt)
	})
}

// CreateDeployment stores a new configuration deployment.
// It validates the deployment, ensures no duplicate exists, and initializes deployment time.
func (s *Store) CreateDeployment(ctx context.Context, deployment *config.Deployment) error {
	if err := s.validateDeployment(deployment); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deploymentKey(deployment.TenantID, deployment.ID)
	if _, exists := s.deployments[key]; exists {
		return config.NewError("create deployment", config.ErrInvalidDeployment, "deployment already exists")
	}

	// Initialize deployment time if not set
	if deployment.DeployedAt.IsZero() {
		deployment.DeployedAt = time.Now()
	}

	s.deployments[key] = deployment
	return nil
}

// GetDeployment retrieves a specific deployment by tenant and deployment ID.
// Returns an error if the deployment doesn't exist or belongs to a different tenant.
func (s *Store) GetDeployment(ctx context.Context, tenantID, deploymentID string) (*config.Deployment, error) {
	if err := s.validateInput("get deployment", map[string]string{
		"tenant ID":     tenantID,
		"deployment ID": deploymentID,
	}); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.deploymentKey(tenantID, deploymentID)
	deployment, exists := s.deployments[key]
	if !exists {
		return nil, config.NewError("get deployment", config.ErrDeploymentNotFound, "deployment not found")
	}

	return deployment, nil
}

// UpdateDeployment updates an existing deployment with new information.
// It ensures the deployment exists and belongs to the correct tenant before updating.
func (s *Store) UpdateDeployment(ctx context.Context, deployment *config.Deployment) error {
	if err := s.validateDeployment(deployment); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deploymentKey(deployment.TenantID, deployment.ID)
	if _, exists := s.deployments[key]; !exists {
		return config.NewError("update deployment", config.ErrDeploymentNotFound, "deployment not found")
	}

	s.deployments[key] = deployment
	return nil
}

// ListDeployments retrieves deployments matching the given criteria.
// It supports filtering by tenant ID, device ID, and status, with pagination.
func (s *Store) ListDeployments(ctx context.Context, opts config.ListOptions) ([]*config.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all deployments in a slice for processing
	deployments := make([]*config.Deployment, 0, len(s.deployments))
	for _, d := range s.deployments {
		deployments = append(deployments, d)
	}

	// Apply filters before sorting
	deployments = s.filterDeployments(deployments, opts)

	// Sort for consistent ordering
	s.sortDeployments(deployments)

	// Apply pagination last
	start, end := s.applyPagination(len(deployments), opts)
	if start >= len(deployments) {
		return []*config.Deployment{}, nil
	}

	return deployments[start:end], nil
}
