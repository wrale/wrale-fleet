package memory

import (
	"context"
	"sync"

	"github.com/wrale/fleet/internal/fleet/config"
)

// Store implements an in-memory configuration store with thread-safe operations
type Store struct {
	mu          sync.RWMutex
	templates   map[string]*config.Template   // key: tenantID/templateID
	versions    map[string][]*config.Version  // key: tenantID/templateID
	deployments map[string]*config.Deployment // key: tenantID/deploymentID
}

// New creates a new in-memory configuration store
func New() *Store {
	return &Store{
		templates:   make(map[string]*config.Template),
		versions:    make(map[string][]*config.Version),
		deployments: make(map[string]*config.Deployment),
	}
}

// templateKey generates a unique key for template storage
func (s *Store) templateKey(tenantID, templateID string) string {
	return tenantID + "/" + templateID
}

// deploymentKey generates a unique key for deployment storage
func (s *Store) deploymentKey(tenantID, deploymentID string) string {
	return tenantID + "/" + deploymentID
}

// CreateTemplate stores a new configuration template
func (s *Store) CreateTemplate(ctx context.Context, template *config.Template) error {
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

	// Apply pagination
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
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deploymentKey(deployment.TenantID, deployment.ID)
	if _, exists := s.deployments[key]; exists {
		return config.NewError("create deployment", config.ErrInvalidDeployment, "deployment already exists")
	}

	s.deployments[key] = deployment
	return nil
}

// GetDeployment retrieves a specific deployment
func (s *Store) GetDeployment(ctx context.Context, tenantID, deploymentID string) (*config.Deployment, error) {
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

	// Apply pagination
	if opts.Offset >= len(deployments) {
		return []*config.Deployment{}, nil
	}

	end := opts.Offset + opts.Limit
	if end > len(deployments) || opts.Limit == 0 {
		end = len(deployments)
	}

	return deployments[opts.Offset:end], nil
}
