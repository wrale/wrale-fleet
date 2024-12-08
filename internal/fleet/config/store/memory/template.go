package memory

import (
	"context"
	"sort"

	"github.com/wrale/wrale-fleet/internal/fleet/config"
)

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

// filterTemplates applies filtering criteria to templates based on the provided options
func (s *Store) filterTemplates(templates []*config.Template, opts config.ListOptions) []*config.Template {
	filtered := make([]*config.Template, 0, len(templates))

	for _, t := range templates {
		matches := true

		// Apply tenant filter if specified
		if opts.TenantID != "" && t.TenantID != opts.TenantID {
			matches = false
		}

		if matches {
			filtered = append(filtered, t)
		}
	}

	return filtered
}

// sortTemplates sorts templates by creation time and ID for consistent ordering
func (s *Store) sortTemplates(templates []*config.Template) {
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].CreatedAt.Equal(templates[j].CreatedAt) {
			return templates[i].ID < templates[j].ID
		}
		return templates[i].CreatedAt.Before(templates[j].CreatedAt)
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

// GetTemplate retrieves a specific configuration template
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

	// Collect all templates in a slice for processing
	templates := make([]*config.Template, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t)
	}

	// Apply filters before sorting
	templates = s.filterTemplates(templates, opts)

	// Sort for consistent ordering
	s.sortTemplates(templates)

	// Apply pagination last
	start, end := s.applyPagination(len(templates), opts)
	if start >= len(templates) {
		return []*config.Template{}, nil
	}

	return templates[start:end], nil
}
