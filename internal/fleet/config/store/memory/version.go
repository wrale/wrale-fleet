package memory

import (
	"context"

	"github.com/wrale/wrale-fleet/internal/fleet/config"
)

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
