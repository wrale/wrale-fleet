package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplate(t *testing.T) {
	schema := json.RawMessage(`{"type": "object"}`)
	template := NewTemplate("tenant-1", "test-template", schema)

	assert.NotEmpty(t, template.ID)
	assert.Equal(t, "tenant-1", template.TenantID)
	assert.Equal(t, "test-template", template.Name)
	assert.Equal(t, schema, template.Schema)
	assert.Empty(t, template.Variables)
	assert.False(t, template.CreatedAt.IsZero())
	assert.False(t, template.UpdatedAt.IsZero())
}

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name        string
		template    *Template
		expectError bool
	}{
		{
			name: "valid template",
			template: &Template{
				ID:       "template-1",
				TenantID: "tenant-1",
				Name:     "test-template",
				Schema:   json.RawMessage(`{"type": "object"}`),
			},
			expectError: false,
		},
		{
			name: "missing ID",
			template: &Template{
				TenantID: "tenant-1",
				Name:     "test-template",
				Schema:   json.RawMessage(`{"type": "object"}`),
			},
			expectError: true,
		},
		{
			name: "missing tenant ID",
			template: &Template{
				ID:     "template-1",
				Name:   "test-template",
				Schema: json.RawMessage(`{"type": "object"}`),
			},
			expectError: true,
		},
		{
			name: "missing name",
			template: &Template{
				ID:       "template-1",
				TenantID: "tenant-1",
				Schema:   json.RawMessage(`{"type": "object"}`),
			},
			expectError: true,
		},
		{
			name: "missing schema",
			template: &Template{
				ID:       "template-1",
				TenantID: "tenant-1",
				Name:     "test-template",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplate_AddVariable(t *testing.T) {
	template := NewTemplate("tenant-1", "test-template", json.RawMessage(`{"type": "object"}`))
	originalUpdatedAt := template.UpdatedAt

	time.Sleep(time.Millisecond) // Ensure time difference

	err := template.AddVariable(Variable{
		Name:        "test-var",
		Type:        "string",
		Description: "Test variable",
		Default:     "default",
		Required:    true,
	})

	require.NoError(t, err)
	assert.Len(t, template.Variables, 1)
	assert.True(t, template.UpdatedAt.After(originalUpdatedAt))

	// Test validation
	err = template.AddVariable(Variable{})
	assert.Error(t, err)
}

func TestTemplate_SetDefault(t *testing.T) {
	template := NewTemplate("tenant-1", "test-template", json.RawMessage(`{"type": "object"}`))
	originalUpdatedAt := template.UpdatedAt

	time.Sleep(time.Millisecond) // Ensure time difference

	defaultConfig := json.RawMessage(`{"key": "value"}`)
	err := template.SetDefault(defaultConfig)

	require.NoError(t, err)
	assert.Equal(t, defaultConfig, template.Default)
	assert.True(t, template.UpdatedAt.After(originalUpdatedAt))

	// Test validation
	err = template.SetDefault(nil)
	assert.Error(t, err)
}

func TestNewVersion(t *testing.T) {
	config := json.RawMessage(`{"key": "value"}`)
	version := NewVersion(config, "template-1", "user-1")

	assert.Equal(t, config, version.Config)
	assert.NotEmpty(t, version.Hash)
	assert.Equal(t, "template-1", version.TemplateID)
	assert.Equal(t, "user-1", version.CreatedBy)
	assert.Equal(t, ValidationStatusPending, version.Status)
	assert.False(t, version.CreatedAt.IsZero())
}

func TestNewDeployment(t *testing.T) {
	version := &Version{
		Number:     1,
		Config:     json.RawMessage(`{"key": "value"}`),
		Hash:       "hash-1",
		CreatedBy:  "user-1",
		Status:     ValidationStatusValid,
		CreatedAt:  time.Now(),
		TemplateID: "template-1",
	}

	deployment := NewDeployment("tenant-1", "device-1", version)

	assert.NotEmpty(t, deployment.ID)
	assert.Equal(t, "tenant-1", deployment.TenantID)
	assert.Equal(t, "device-1", deployment.DeviceID)
	assert.Equal(t, version, deployment.ConfigVersion)
	assert.Equal(t, "pending", deployment.Status)
	assert.False(t, deployment.DeployedAt.IsZero())
	assert.Nil(t, deployment.CompletedAt)
	assert.Empty(t, deployment.Error)
}

func TestDeployment_Complete(t *testing.T) {
	deployment := NewDeployment("tenant-1", "device-1", &Version{})
	deployment.Complete()

	assert.Equal(t, "completed", deployment.Status)
	assert.NotNil(t, deployment.CompletedAt)
}

func TestDeployment_Fail(t *testing.T) {
	deployment := NewDeployment("tenant-1", "device-1", &Version{})
	deployment.Fail("test error")

	assert.Equal(t, "failed", deployment.Status)
	assert.NotNil(t, deployment.CompletedAt)
	assert.Equal(t, "test error", deployment.Error)
}

func Test_calculateHash(t *testing.T) {
	config := json.RawMessage(`{"key": "value"}`)
	hash1 := calculateHash(config)
	hash2 := calculateHash(config)

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2)

	differentConfig := json.RawMessage(`{"key": "different"}`)
	hash3 := calculateHash(differentConfig)
	assert.NotEqual(t, hash1, hash3)
}
