package config

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

// Service provides configuration management operations
type Service struct {
	store  Store
	logger *zap.Logger
}

// NewService creates a new configuration management service
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// CreateTemplate creates a new configuration template
func (s *Service) CreateTemplate(ctx context.Context, tenantID, name string, schema json.RawMessage) (*Template, error) {
	template := NewTemplate(tenantID, name, schema)

	if err := template.Validate(); err != nil {
		return nil, NewError("create template", ErrInvalidTemplate, "invalid template data")
	}

	if err := s.store.CreateTemplate(ctx, template); err != nil {
		return nil, err
	}

	s.logger.Info("created configuration template",
		zap.String("template_id", template.ID),
		zap.String("tenant_id", template.TenantID),
		zap.String("name", template.Name),
	)

	return template, nil
}

// CreateVersion creates a new configuration version from a template
func (s *Service) CreateVersion(ctx context.Context, tenantID, templateID string, config json.RawMessage, createdBy string) (*Version, error) {
	template, err := s.store.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, err
	}

	version := NewVersion(config, template.ID, createdBy)

	if err := s.store.CreateVersion(ctx, tenantID, templateID, version); err != nil {
		return nil, err
	}

	s.logger.Info("created configuration version",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", version.Number),
		zap.String("created_by", createdBy),
	)

	return version, nil
}

// DeployConfiguration deploys a configuration version to a device
func (s *Service) DeployConfiguration(ctx context.Context, tenantID, templateID string, version *Version, deviceID string) (*Deployment, error) {
	deployment := NewDeployment(tenantID, deviceID, version)

	if err := s.store.CreateDeployment(ctx, deployment); err != nil {
		return nil, err
	}

	s.logger.Info("initiated configuration deployment",
		zap.String("deployment_id", deployment.ID),
		zap.String("device_id", deviceID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", version.Number),
	)

	return deployment, nil
}

// CompleteDeployment marks a deployment as successfully completed
func (s *Service) CompleteDeployment(ctx context.Context, tenantID, deploymentID string) error {
	deployment, err := s.store.GetDeployment(ctx, tenantID, deploymentID)
	if err != nil {
		return err
	}

	deployment.Complete()

	if err := s.store.UpdateDeployment(ctx, deployment); err != nil {
		return err
	}

	s.logger.Info("completed configuration deployment",
		zap.String("deployment_id", deploymentID),
		zap.String("device_id", deployment.DeviceID),
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// FailDeployment marks a deployment as failed
func (s *Service) FailDeployment(ctx context.Context, tenantID, deploymentID, errorMsg string) error {
	deployment, err := s.store.GetDeployment(ctx, tenantID, deploymentID)
	if err != nil {
		return err
	}

	deployment.Fail(errorMsg)

	if err := s.store.UpdateDeployment(ctx, deployment); err != nil {
		return err
	}

	s.logger.Error("configuration deployment failed",
		zap.String("deployment_id", deploymentID),
		zap.String("device_id", deployment.DeviceID),
		zap.String("tenant_id", tenantID),
		zap.String("error", errorMsg),
	)

	return nil
}

// ValidateVersion marks a configuration version as validated
func (s *Service) ValidateVersion(ctx context.Context, tenantID, templateID string, versionNumber int) error {
	version, err := s.store.GetVersion(ctx, tenantID, templateID, versionNumber)
	if err != nil {
		return err
	}

	version.Status = ValidationStatusValid

	if err := s.store.UpdateVersion(ctx, tenantID, templateID, version); err != nil {
		return err
	}

	s.logger.Info("validated configuration version",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", versionNumber),
	)

	return nil
}

// RollbackVersion marks a configuration version for rollback
func (s *Service) RollbackVersion(ctx context.Context, tenantID, templateID string, versionNumber int) error {
	version, err := s.store.GetVersion(ctx, tenantID, templateID, versionNumber)
	if err != nil {
		return err
	}

	version.Status = ValidationStatusRollback

	if err := s.store.UpdateVersion(ctx, tenantID, templateID, version); err != nil {
		return err
	}

	s.logger.Warn("marked configuration version for rollback",
		zap.String("template_id", templateID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", versionNumber),
	)

	return nil
}
