package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/config"
)

func TestVersion_Operations(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create test template for version tests
	template := createTestTemplate("version-test", "tenant-version")
	require.NoError(t, store.CreateTemplate(ctx, template))

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name      string
			tenantID  string
			template  string
			version   *config.Version
			wantErr   bool
			wantCount int
		}{
			{
				name:      "first version",
				tenantID:  template.TenantID,
				template:  template.ID,
				version:   createTestVersion(template.ID, 0),
				wantErr:   false,
				wantCount: 1,
			},
			{
				name:      "second version",
				tenantID:  template.TenantID,
				template:  template.ID,
				version:   createTestVersion(template.ID, 0),
				wantErr:   false,
				wantCount: 2,
			},
			{
				name:     "wrong tenant",
				tenantID: "wrong-tenant",
				template: template.ID,
				version:  createTestVersion(template.ID, 0),
				wantErr:  true,
			},
			{
				name:     "non-existent template",
				tenantID: template.TenantID,
				template: "missing",
				version:  createTestVersion("missing", 0),
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := store.CreateVersion(ctx, tt.tenantID, tt.template, tt.version)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				// Verify version creation
				versions, err := store.ListVersions(ctx, tt.tenantID, tt.template)
				require.NoError(t, err)
				if tt.wantCount > 0 {
					assert.Len(t, versions, tt.wantCount)
					assert.Equal(t, tt.wantCount, versions[len(versions)-1].Number)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		tests := []struct {
			name          string
			tenantID      string
			templateID    string
			versionNumber int
			wantErr       bool
		}{
			{
				name:          "existing version",
				tenantID:      template.TenantID,
				templateID:    template.ID,
				versionNumber: 1,
				wantErr:       false,
			},
			{
				name:          "wrong tenant",
				tenantID:      "wrong-tenant",
				templateID:    template.ID,
				versionNumber: 1,
				wantErr:       true,
			},
			{
				name:          "non-existent version",
				tenantID:      template.TenantID,
				templateID:    template.ID,
				versionNumber: 99,
				wantErr:       true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				version, err := store.GetVersion(ctx, tt.tenantID, tt.templateID, tt.versionNumber)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.versionNumber, version.Number)
			})
		}
	})

	t.Run("List", func(t *testing.T) {
		// List all versions
		versions, err := store.ListVersions(ctx, template.TenantID, template.ID)
		require.NoError(t, err)
		assert.Len(t, versions, 2)
		assert.Equal(t, 1, versions[0].Number)
		assert.Equal(t, 2, versions[1].Number)

		// List versions for non-existent template
		_, err = store.ListVersions(ctx, template.TenantID, "missing")
		require.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		// Get existing version
		version, err := store.GetVersion(ctx, template.TenantID, template.ID, 1)
		require.NoError(t, err)

		// Update version status
		version.Status = config.ValidationStatusValid
		err = store.UpdateVersion(ctx, template.TenantID, template.ID, version)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetVersion(ctx, template.TenantID, template.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, config.ValidationStatusValid, updated.Status)

		// Test update of non-existent version
		version.Number = 99
		err = store.UpdateVersion(ctx, template.TenantID, template.ID, version)
		require.Error(t, err)
	})
}
