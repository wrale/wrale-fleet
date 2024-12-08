package config

import "context"

// ListOptions provides filtering and pagination for list operations
type ListOptions struct {
	TenantID string
	DeviceID string
	Status   string
	Offset   int
	Limit    int
}

// Store defines the interface for configuration storage
type Store interface {
	// Template operations
	CreateTemplate(ctx context.Context, template *Template) error
	GetTemplate(ctx context.Context, tenantID, templateID string) (*Template, error)
	UpdateTemplate(ctx context.Context, template *Template) error
	DeleteTemplate(ctx context.Context, tenantID, templateID string) error
	ListTemplates(ctx context.Context, opts ListOptions) ([]*Template, error)

	// Version operations
	CreateVersion(ctx context.Context, tenantID, templateID string, version *Version) error
	GetVersion(ctx context.Context, tenantID, templateID string, versionNumber int) (*Version, error)
	ListVersions(ctx context.Context, tenantID, templateID string) ([]*Version, error)
	UpdateVersion(ctx context.Context, tenantID, templateID string, version *Version) error

	// Deployment operations
	CreateDeployment(ctx context.Context, deployment *Deployment) error
	GetDeployment(ctx context.Context, tenantID, deploymentID string) (*Deployment, error)
	UpdateDeployment(ctx context.Context, deployment *Deployment) error
	ListDeployments(ctx context.Context, opts ListOptions) ([]*Deployment, error)
}
