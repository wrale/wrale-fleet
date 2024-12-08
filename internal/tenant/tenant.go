package tenant

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ResourceQuota defines resource limits for a tenant
type ResourceQuota struct {
	MaxDevices        int     `json:"max_devices"`
	MaxGroups         int     `json:"max_groups"`
	MaxUsers          int     `json:"max_users"`
	MaxStorageGB      float64 `json:"max_storage_gb"`
	MaxBandwidthMBps  float64 `json:"max_bandwidth_mbps"`
	MaxConfigVersions int     `json:"max_config_versions"`
}

// ComplianceConfig defines tenant compliance requirements
type ComplianceConfig struct {
	RequiredFrameworks []string          `json:"required_frameworks"`
	CustomPolicies     []json.RawMessage `json:"custom_policies,omitempty"`
	AuditInterval      time.Duration     `json:"audit_interval"`
	RetentionPeriod    time.Duration     `json:"retention_period"`
}

// AirgapConfig defines tenant airgapped operation settings
type AirgapConfig struct {
	Enabled           bool          `json:"enabled"`
	SyncInterval      time.Duration `json:"sync_interval,omitempty"`
	MaxOfflinePeriod  time.Duration `json:"max_offline_period,omitempty"`
	AllowedOperations []string      `json:"allowed_operations,omitempty"`
	DataBufferSize    int64         `json:"data_buffer_size,omitempty"`
}

// Status represents the tenant's current state
type Status string

const (
	StatusActive          Status = "active"
	StatusSuspended       Status = "suspended"
	StatusProvisioning    Status = "provisioning"
	StatusDecommissioning Status = "decommissioning"
)

// Tenant represents an enterprise customer in the system
type Tenant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status Status `json:"status"`

	// Resource management
	ResourceQuota *ResourceQuota   `json:"resource_quota,omitempty"`
	ResourceUsage map[string]int64 `json:"resource_usage,omitempty"`

	// Security and compliance
	ComplianceConfig *ComplianceConfig `json:"compliance_config,omitempty"`
	SecurityLevel    string            `json:"security_level,omitempty"`

	// Airgapped operation support
	AirgapConfig *AirgapConfig `json:"airgap_config,omitempty"`

	// Tenant metadata
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Settings    json.RawMessage   `json:"settings,omitempty"`

	// Audit fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

// New creates a new Tenant with generated ID and timestamps
func New(name string) *Tenant {
	now := time.Now().UTC()
	return &Tenant{
		ID:            uuid.New().String(),
		Name:          name,
		Status:        StatusProvisioning,
		ResourceUsage: make(map[string]int64),
		Metadata:      make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate checks if the tenant data is valid
func (t *Tenant) Validate() error {
	const op = "Tenant.Validate"

	if t.ID == "" {
		return E(op, ErrCodeInvalidTenant, "tenant id cannot be empty", nil)
	}
	if t.Name == "" {
		return E(op, ErrCodeInvalidTenant, "tenant name cannot be empty", nil)
	}

	// Validate resource quota if present
	if t.ResourceQuota != nil {
		if t.ResourceQuota.MaxDevices < 0 {
			return E(op, ErrCodeInvalidTenant, "max devices cannot be negative", nil)
		}
		if t.ResourceQuota.MaxGroups < 0 {
			return E(op, ErrCodeInvalidTenant, "max groups cannot be negative", nil)
		}
		if t.ResourceQuota.MaxUsers < 0 {
			return E(op, ErrCodeInvalidTenant, "max users cannot be negative", nil)
		}
		if t.ResourceQuota.MaxStorageGB < 0 {
			return E(op, ErrCodeInvalidTenant, "max storage cannot be negative", nil)
		}
		if t.ResourceQuota.MaxBandwidthMBps < 0 {
			return E(op, ErrCodeInvalidTenant, "max bandwidth cannot be negative", nil)
		}
	}

	// Validate airgap config if present
	if t.AirgapConfig != nil {
		if t.AirgapConfig.SyncInterval < 0 {
			return E(op, ErrCodeInvalidTenant, "sync interval cannot be negative", nil)
		}
		if t.AirgapConfig.MaxOfflinePeriod < 0 {
			return E(op, ErrCodeInvalidTenant, "max offline period cannot be negative", nil)
		}
		if t.AirgapConfig.DataBufferSize < 0 {
			return E(op, ErrCodeInvalidTenant, "data buffer size cannot be negative", nil)
		}
	}

	return nil
}

// SetResourceQuota updates the tenant's resource quota
func (t *Tenant) SetResourceQuota(quota *ResourceQuota) error {
	const op = "Tenant.SetResourceQuota"

	if quota == nil {
		return E(op, ErrCodeInvalidOperation, "resource quota cannot be nil", nil)
	}

	t.ResourceQuota = quota
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateResourceUsage updates resource usage metrics
func (t *Tenant) UpdateResourceUsage(resource string, value int64) error {
	const op = "Tenant.UpdateResourceUsage"

	if resource == "" {
		return E(op, ErrCodeInvalidOperation, "resource name cannot be empty", nil)
	}

	if t.ResourceUsage == nil {
		t.ResourceUsage = make(map[string]int64)
	}
	t.ResourceUsage[resource] = value
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// CheckQuota verifies if a resource usage is within quota limits
func (t *Tenant) CheckQuota(resource string, requestedValue int64) error {
	const op = "Tenant.CheckQuota"

	if t.ResourceQuota == nil {
		return nil // No quotas defined
	}

	var limit int64
	switch resource {
	case "devices":
		limit = int64(t.ResourceQuota.MaxDevices)
	case "groups":
		limit = int64(t.ResourceQuota.MaxGroups)
	case "users":
		limit = int64(t.ResourceQuota.MaxUsers)
	default:
		return E(op, ErrCodeInvalidOperation, "unknown resource type", nil)
	}

	if limit > 0 && requestedValue > limit {
		return E(op, ErrCodeQuotaExceeded,
			fmt.Sprintf("quota exceeded for %s: requested %d, limit %d",
				resource, requestedValue, limit), nil)
	}

	return nil
}

// SetComplianceConfig updates the tenant's compliance configuration
func (t *Tenant) SetComplianceConfig(config *ComplianceConfig) error {
	const op = "Tenant.SetComplianceConfig"

	if config == nil {
		return E(op, ErrCodeInvalidOperation, "compliance config cannot be nil", nil)
	}

	t.ComplianceConfig = config
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// SetAirgapConfig updates the tenant's airgapped operation configuration
func (t *Tenant) SetAirgapConfig(config *AirgapConfig) error {
	const op = "Tenant.SetAirgapConfig"

	if config == nil {
		return E(op, ErrCodeInvalidOperation, "airgap config cannot be nil", nil)
	}

	if config.Enabled {
		if config.SyncInterval <= 0 {
			return E(op, ErrCodeInvalidOperation, "sync interval must be positive when airgap is enabled", nil)
		}
		if config.MaxOfflinePeriod <= 0 {
			return E(op, ErrCodeInvalidOperation, "max offline period must be positive when airgap is enabled", nil)
		}
		if len(config.AllowedOperations) == 0 {
			return E(op, ErrCodeInvalidOperation, "at least one allowed operation must be specified", nil)
		}
	}

	t.AirgapConfig = config
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// AddMetadata adds or updates metadata
func (t *Tenant) AddMetadata(key, value string) error {
	const op = "Tenant.AddMetadata"

	if key == "" {
		return E(op, ErrCodeInvalidOperation, "metadata key cannot be empty", nil)
	}

	if t.Metadata == nil {
		t.Metadata = make(map[string]string)
	}
	t.Metadata[key] = value
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// SetStatus updates the tenant's status
func (t *Tenant) SetStatus(status Status) {
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
}

// IsActive checks if the tenant is in active status
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}

// IsSuspended checks if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == StatusSuspended
}

// UpdateSettings updates tenant-specific settings
func (t *Tenant) UpdateSettings(settings json.RawMessage) error {
	const op = "Tenant.UpdateSettings"

	if len(settings) == 0 {
		return E(op, ErrCodeInvalidOperation, "settings cannot be empty", nil)
	}

	t.Settings = settings
	t.UpdatedAt = time.Now().UTC()
	return nil
}
