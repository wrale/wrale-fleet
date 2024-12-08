package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wrale/fleet/internal/fleet/config"
)

// Store implements an in-memory configuration store with thread-safe operations.
// It ensures proper validation, consistent ordering, and safe concurrent access.
type Store struct {
	mu          sync.RWMutex
	templates   map[string]*config.Template   // key: tenantID/templateID
	versions    map[string][]*config.Version  // key: tenantID/templateID
	deployments map[string]*config.Deployment // key: tenantID/deploymentID
}

// New creates a new in-memory configuration store with initialized maps
func New() *Store {
	return &Store{
		templates:   make(map[string]*config.Template),
		versions:    make(map[string][]*config.Version),
		deployments: make(map[string]*config.Deployment),
	}
}

// validateInput is a helper function to check for required string fields
func (s *Store) validateInput(op string, fields map[string]string) error {
	for name, value := range fields {
		if value == "" {
			return config.NewError(op, config.ErrValidationFailed, name+" is required")
		}
	}
	return nil
}

// templateKey generates a unique key for template storage
func (s *Store) templateKey(tenantID, templateID string) string {
	return tenantID + "/" + templateID
}

// deploymentKey generates a unique key for deployment storage
func (s *Store) deploymentKey(tenantID, deploymentID string) string {
	return tenantID + "/" + deploymentID
}

// validateTemplate ensures the template is valid before storage operations
func (s *Store) validateTemplate(template *config.Template) error {
	if template == nil {
		return config.NewError("validate template", config.ErrInvalidTemplate, "template is nil")
	}
	return s.validateInput("validate template", map[string]string{
		"template ID": template.ID,
		"tenant ID":   template.TenantID,
		"name":        template.Name,
	})
}

// validateDeployment ensures the deployment is valid before storage operations
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

// CreateTemplate stores a new configuration template
func (s *Store) CreateTemplate(ctx context.Context, template *config.Template) error {
	if err := s.validateTemplate(template); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.templateKey(template.TenantID, template.ID)
	if _, exists := s.templates[key]; exists {
		return config.NewError("create template", config.ErrInvalidTemplate, "template already exists")
	}

	s.templates[key] = template
	return nil
}

// GetTemplate retrieves a configuration template
func (s *Store) GetTemplate(ctx context.Context, tenantID, templateID string) (*config.Template, error) {
	if err := s.validateInput("get template", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.templateKey(tenantID, templateID)
	template, exists := s.templates[key]
	if !exists {
		return nil, config.NewError("get template", config.ErrTemplateNotFound, "template not found")
	}

	return template, nil
}

// UpdateTemplate updates an existing configuration template
func (s *Store) UpdateTemplate(ctx context.Context, template *config.Template) error {
	if err := s.validateTemplate(template); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.templateKey(template.TenantID, template.ID)
	if _, exists := s.templates[key]; !exists {
		return config.NewError("update template", config.ErrTemplateNotFound, "template not found")
	}

	s.templates[key] = template
	return nil
}

// DeleteTemplate removes a configuration template and its versions
func (s *Store) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	if err := s.validateInput("delete template", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.templateKey(tenantID, templateID)
	if _, exists := s.templates[key]; !exists {
		return config.NewError("delete template", config.ErrTemplateNotFound, "template not found")
	}

	delete(s.templates, key)
	delete(s.versions, key)
	return nil
}

// ListTemplates retrieves templates matching the given criteria
func (s *Store) ListTemplates(ctx context.Context, opts config.ListOptions) ([]*config.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var templates []*config.Template
	for _, template := range s.templates {
		if opts.TenantID != "" && template.TenantID != opts.TenantID {
			continue
		}
		templates = append(templates, template)
	}

	// Sort templates by creation time then ID for consistent ordering
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].CreatedAt.Equal(templates[j].CreatedAt) {
			return templates[i].ID < templates[j].ID
		}
		return templates[i].CreatedAt.Before(templates[j].CreatedAt)
	})

	// Apply pagination after sorting
	if opts.Offset >= len(templates) {
		return []*config.Template{}, nil
	}

	end := opts.Offset + opts.Limit
	if end > len(templates) || opts.Limit == 0 {
		end = len(templates)
	}

	return templates[opts.Offset:end], nil
}

// CreateVersion stores a new configuration version
func (s *Store) CreateVersion(ctx context.Context, tenantID, templateID string, version *config.Version) error {
	if err := s.validateInput("create version", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return err
	}
	if version == nil {
		return config.NewError("create version", config.ErrInvalidVersion, "version is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.templateKey(tenantID, templateID)
	if _, exists := s.templates[key]; !exists {
		return config.NewError("create version", config.ErrTemplateNotFound, "template not found")
	}

	versions := s.versions[key]
	version.Number = len(versions) + 1
	s.versions[key] = append(versions, version)
	return nil
}

// GetVersion retrieves a specific configuration version
func (s *Store) GetVersion(ctx context.Context, tenantID, templateID string, versionNumber int) (*config.Version, error) {
	if err := s.validateInput("get version", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.templateKey(tenantID, templateID)
	versions, exists := s.versions[key]
	if !exists || versionNumber < 1 || versionNumber > len(versions) {
		return nil, config.NewError("get version", config.ErrVersionNotFound, "version not found")
	}

	return versions[versionNumber-1], nil
}

// ListVersions retrieves all versions for a template
func (s *Store) ListVersions(ctx context.Context, tenantID, templateID string) ([]*config.Version, error) {
	if err := s.validateInput("list versions", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.templateKey(tenantID, templateID)
	if _, exists := s.templates[key]; !exists {
		return nil, config.NewError("list versions", config.ErrTemplateNotFound, "template not found")
	}

	versions := s.versions[key]
	if versions == nil {
		versions = make([]*config.Version, 0)
	}
	return versions, nil
}

// UpdateVersion updates an existing configuration version
func (s *Store) UpdateVersion(ctx context.Context, tenantID, templateID string, version *config.Version) error {
	if err := s.validateInput("update version", map[string]string{
		"tenant ID":   tenantID,
		"template ID": templateID,
	}); err != nil {
		return err
	}
	if version == nil {
		return config.NewError("update version", config.ErrInvalidVersion, "version is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.templateKey(tenantID, templateID)
	versions, exists := s.versions[key]
	if !exists || version.Number < 1 || version.Number > len(versions) {
		return config.NewError("update version", config.ErrVersionNotFound, "version not found")
	}

	versions[version.Number-1] = version
	return nil
}

// CreateDeployment stores a new configuration deployment
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

	if deployment.DeployedAt.IsZero() {
		deployment.DeployedAt = time.Now()
	}

	s.deployments[key] = deployment
	return nil
}

// GetDeployment retrieves a specific deployment
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

// UpdateDeployment updates an existing deployment
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

// ListDeployments retrieves deployments matching the given criteria
func (s *Store) ListDeployments(ctx context.Context, opts config.ListOptions) ([]*config.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First collect all matching deployments
	var deployments []*config.Deployment
	for _, deployment := range s.deployments {
		if opts.TenantID != "" && deployment.TenantID != opts.TenantID {
			continue
		}
		if opts.DeviceID != "" && deployment.DeviceID != opts.DeviceID {
			continue
		}
		if opts.Status != "" && deployment.Status != opts.Status {
			continue
		}
		deployments = append(deployments, deployment)
	}

	// Sort deployments by deployment time then ID for consistent ordering
	sort.Slice(deployments, func(i, j int) bool {
		if deployments[i].DeployedAt.Equal(deployments[j].DeployedAt) {
			return deployments[i].ID < deployments[j].ID
		}
		return deployments[i].DeployedAt.Before(deployments[j].DeployedAt)
	})

	// Apply pagination after filtering and sorting
	if opts.Offset >= len(deployments) {
		return []*config.Deployment{}, nil
	}

	end := opts.Offset + opts.Limit
	if end > len(deployments) || opts.Limit == 0 {
		end = len(deployments)
	}

	return deployments[opts.Offset:end], nil
}
