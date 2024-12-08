package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ValidationStatus represents the current state of configuration validation
type ValidationStatus string

const (
	ValidationStatusPending  ValidationStatus = "pending"
	ValidationStatusValid    ValidationStatus = "valid"
	ValidationStatusInvalid  ValidationStatus = "invalid"
	ValidationStatusRollback ValidationStatus = "rollback"
)

// Template represents a reusable configuration template
type Template struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema"`    // JSON Schema for validation
	Default     json.RawMessage `json:"default"`   // Default configuration values
	Variables   []Variable      `json:"variables"` // Configurable variables
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Variable represents a configurable parameter in a template
type Variable struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required"`
	Validation  string      `json:"validation,omitempty"` // JSON Schema validation rules
}

// Version represents a specific configuration version
type Version struct {
	Number      int              `json:"version"`
	Config      json.RawMessage  `json:"config"`
	Hash        string           `json:"hash"`
	TemplateID  string           `json:"template_id,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	CreatedBy   string           `json:"created_by"`
	ValidatedAt *time.Time       `json:"validated_at,omitempty"`
	Status      ValidationStatus `json:"status"`
}

// Deployment tracks configuration deployment to devices
type Deployment struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	ConfigVersion *Version   `json:"config_version"`
	DeviceID      string     `json:"device_id"`
	Status        string     `json:"status"`
	DeployedAt    time.Time  `json:"deployed_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Error         string     `json:"error,omitempty"`
}

// NewTemplate creates a new configuration template
func NewTemplate(tenantID, name string, schema json.RawMessage) *Template {
	now := time.Now().UTC()
	return &Template{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Schema:    schema,
		Variables: make([]Variable, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewVersion creates a new configuration version
func NewVersion(config json.RawMessage, templateID, createdBy string) *Version {
	return &Version{
		Config:     config,
		Hash:       calculateHash(config),
		TemplateID: templateID,
		CreatedAt:  time.Now().UTC(),
		CreatedBy:  createdBy,
		Status:     ValidationStatusPending,
	}
}

// NewDeployment creates a new configuration deployment
func NewDeployment(tenantID, deviceID string, version *Version) *Deployment {
	return &Deployment{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		ConfigVersion: version,
		DeviceID:      deviceID,
		Status:        "pending",
		DeployedAt:    time.Now().UTC(),
	}
}

// Validate checks if a configuration template is valid
func (t *Template) Validate() error {
	if t.ID == "" {
		return NewError("validate template", ErrInvalidTemplate, "template id cannot be empty")
	}
	if t.TenantID == "" {
		return NewError("validate template", ErrInvalidTemplate, "tenant id cannot be empty")
	}
	if t.Name == "" {
		return NewError("validate template", ErrInvalidTemplate, "template name cannot be empty")
	}
	if len(t.Schema) == 0 {
		return NewError("validate template", ErrInvalidTemplate, "template schema cannot be empty")
	}
	return nil
}

// AddVariable adds a new variable to the template
func (t *Template) AddVariable(v Variable) error {
	if v.Name == "" {
		return NewError("add variable", ErrInvalidTemplate, "variable name cannot be empty")
	}
	if v.Type == "" {
		return NewError("add variable", ErrInvalidTemplate, "variable type cannot be empty")
	}

	t.Variables = append(t.Variables, v)
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// SetDefault sets the default configuration for the template
func (t *Template) SetDefault(config json.RawMessage) error {
	if len(config) == 0 {
		return NewError("set default", ErrInvalidTemplate, "default config cannot be empty")
	}

	t.Default = config
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Complete marks a deployment as completed
func (d *Deployment) Complete() {
	now := time.Now().UTC()
	d.CompletedAt = &now
	d.Status = "completed"
}

// Fail marks a deployment as failed with an error
func (d *Deployment) Fail(err string) {
	now := time.Now().UTC()
	d.CompletedAt = &now
	d.Status = "failed"
	d.Error = err
}

// calculateHash generates a SHA-256 hash of the configuration
func calculateHash(config json.RawMessage) string {
	hash := sha256.Sum256(config)
	return hex.EncodeToString(hash[:])
}
