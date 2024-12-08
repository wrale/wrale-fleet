package tenant

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tenant := New("test-tenant")
	assert.NotEmpty(t, tenant.ID)
	assert.Equal(t, "test-tenant", tenant.Name)
	assert.Equal(t, StatusProvisioning, tenant.Status)
	assert.NotNil(t, tenant.ResourceUsage)
	assert.NotNil(t, tenant.Metadata)
	assert.False(t, tenant.CreatedAt.IsZero())
	assert.False(t, tenant.UpdatedAt.IsZero())
}

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *Tenant
		wantErr string
	}{
		{
			name: "valid tenant",
			tenant: &Tenant{
				ID:     "test-id",
				Name:   "test-tenant",
				Status: StatusActive,
			},
			wantErr: "",
		},
		{
			name: "missing id",
			tenant: &Tenant{
				Name:   "test-tenant",
				Status: StatusActive,
			},
			wantErr: "tenant id cannot be empty",
		},
		{
			name: "missing name",
			tenant: &Tenant{
				ID:     "test-id",
				Status: StatusActive,
			},
			wantErr: "tenant name cannot be empty",
		},
		{
			name: "invalid resource quota - negative devices",
			tenant: &Tenant{
				ID:     "test-id",
				Name:   "test-tenant",
				Status: StatusActive,
				ResourceQuota: &ResourceQuota{
					MaxDevices: -1,
				},
			},
			wantErr: "max devices cannot be negative",
		},
		{
			name: "invalid airgap config - negative sync interval",
			tenant: &Tenant{
				ID:     "test-id",
				Name:   "test-tenant",
				Status: StatusActive,
				AirgapConfig: &AirgapConfig{
					Enabled:      true,
					SyncInterval: -1 * time.Hour,
				},
			},
			wantErr: "sync interval cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenant_ResourceQuotaManagement(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("set valid quota", func(t *testing.T) {
		quota := &ResourceQuota{
			MaxDevices:       1000,
			MaxGroups:        100,
			MaxUsers:         50,
			MaxStorageGB:     1000.0,
			MaxBandwidthMBps: 100.0,
		}
		err := tenant.SetResourceQuota(quota)
		require.NoError(t, err)
		assert.Equal(t, quota, tenant.ResourceQuota)
	})

	t.Run("nil quota", func(t *testing.T) {
		err := tenant.SetResourceQuota(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource quota cannot be nil")
	})

	t.Run("check quota enforcement", func(t *testing.T) {
		// Test within limits
		err := tenant.CheckQuota("devices", 500)
		require.NoError(t, err)

		// Test exceeding limits
		err = tenant.CheckQuota("devices", 1500)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded for devices")

		// Test unknown resource type
		err = tenant.CheckQuota("unknown", 100)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown resource type")
	})

	t.Run("update resource usage", func(t *testing.T) {
		err := tenant.UpdateResourceUsage("devices", 100)
		require.NoError(t, err)
		assert.Equal(t, int64(100), tenant.ResourceUsage["devices"])

		err = tenant.UpdateResourceUsage("", 100)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resource name cannot be empty")
	})
}

func TestTenant_ComplianceConfiguration(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("set valid compliance config", func(t *testing.T) {
		config := &ComplianceConfig{
			RequiredFrameworks: []string{"ISO27001", "SOC2"},
			AuditInterval:      24 * time.Hour,
			RetentionPeriod:    90 * 24 * time.Hour,
		}
		err := tenant.SetComplianceConfig(config)
		require.NoError(t, err)
		assert.Equal(t, config, tenant.ComplianceConfig)
	})

	t.Run("nil compliance config", func(t *testing.T) {
		err := tenant.SetComplianceConfig(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "compliance config cannot be nil")
	})

	t.Run("custom policies", func(t *testing.T) {
		policy := json.RawMessage(`{"key": "value"}`)
		config := &ComplianceConfig{
			RequiredFrameworks: []string{"ISO27001"},
			CustomPolicies:     []json.RawMessage{policy},
			AuditInterval:      24 * time.Hour,
			RetentionPeriod:    90 * 24 * time.Hour,
		}
		err := tenant.SetComplianceConfig(config)
		require.NoError(t, err)
		assert.Equal(t, config, tenant.ComplianceConfig)
	})
}

func TestTenant_AirgapConfiguration(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("set valid airgap config", func(t *testing.T) {
		config := &AirgapConfig{
			Enabled:           true,
			SyncInterval:      1 * time.Hour,
			MaxOfflinePeriod:  24 * time.Hour,
			AllowedOperations: []string{"READ", "WRITE"},
			DataBufferSize:    1024 * 1024 * 1024, // 1GB
		}
		err := tenant.SetAirgapConfig(config)
		require.NoError(t, err)
		assert.Equal(t, config, tenant.AirgapConfig)
	})

	t.Run("invalid airgap config - missing operations", func(t *testing.T) {
		config := &AirgapConfig{
			Enabled:          true,
			SyncInterval:     1 * time.Hour,
			MaxOfflinePeriod: 24 * time.Hour,
		}
		err := tenant.SetAirgapConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one allowed operation must be specified")
	})

	t.Run("invalid airgap config - negative sync interval", func(t *testing.T) {
		config := &AirgapConfig{
			Enabled:           true,
			SyncInterval:      -1 * time.Hour,
			MaxOfflinePeriod:  24 * time.Hour,
			AllowedOperations: []string{"READ"},
		}
		err := tenant.SetAirgapConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sync interval must be positive")
	})

	t.Run("nil airgap config", func(t *testing.T) {
		err := tenant.SetAirgapConfig(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "airgap config cannot be nil")
	})

	t.Run("disabled airgap config", func(t *testing.T) {
		config := &AirgapConfig{
			Enabled: false,
		}
		err := tenant.SetAirgapConfig(config)
		require.NoError(t, err)
		assert.Equal(t, config, tenant.AirgapConfig)
	})
}

func TestTenant_MetadataManagement(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("add valid metadata", func(t *testing.T) {
		err := tenant.AddMetadata("environment", "production")
		require.NoError(t, err)
		assert.Equal(t, "production", tenant.Metadata["environment"])
	})

	t.Run("empty key", func(t *testing.T) {
		err := tenant.AddMetadata("", "value")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "metadata key cannot be empty")
	})

	t.Run("update existing metadata", func(t *testing.T) {
		err := tenant.AddMetadata("environment", "staging")
		require.NoError(t, err)
		assert.Equal(t, "staging", tenant.Metadata["environment"])
	})
}

func TestTenant_StatusManagement(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("status transitions", func(t *testing.T) {
		// Initial status should be provisioning
		assert.Equal(t, StatusProvisioning, tenant.Status)
		assert.False(t, tenant.IsActive())
		assert.False(t, tenant.IsSuspended())

		// Transition to active
		tenant.SetStatus(StatusActive)
		assert.Equal(t, StatusActive, tenant.Status)
		assert.True(t, tenant.IsActive())
		assert.False(t, tenant.IsSuspended())

		// Transition to suspended
		tenant.SetStatus(StatusSuspended)
		assert.Equal(t, StatusSuspended, tenant.Status)
		assert.False(t, tenant.IsActive())
		assert.True(t, tenant.IsSuspended())
	})
}

func TestTenant_Settings(t *testing.T) {
	tenant := New("test-tenant")

	t.Run("update valid settings", func(t *testing.T) {
		settings := json.RawMessage(`{"key": "value"}`)
		err := tenant.UpdateSettings(settings)
		require.NoError(t, err)
		assert.Equal(t, settings, tenant.Settings)
	})

	t.Run("empty settings", func(t *testing.T) {
		err := tenant.UpdateSettings(json.RawMessage{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "settings cannot be empty")
	})
}
