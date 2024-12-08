package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/config"
)

func TestDeployment_Operations(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Reset store before each test
	t.Cleanup(store.clearStore)

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name       string
			deployment *config.Deployment
			wantErr    bool
		}{
			{
				name:       "valid deployment",
				deployment: createTestDeployment("deploy-1", "tenant-1", "device-1"),
				wantErr:    false,
			},
			{
				name:       "duplicate deployment",
				deployment: createTestDeployment("deploy-1", "tenant-1", "device-1"),
				wantErr:    true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := store.CreateDeployment(ctx, tt.deployment)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				// Verify deployment creation
				stored, err := store.GetDeployment(ctx, tt.deployment.TenantID, tt.deployment.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.deployment.ID, stored.ID)
				assert.Equal(t, "pending", stored.Status)
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		store.clearStore() // Reset state before get operations
		deployment := createTestDeployment("deploy-2", "tenant-2", "device-2")
		require.NoError(t, store.CreateDeployment(ctx, deployment))

		tests := []struct {
			name     string
			tenantID string
			id       string
			wantErr  bool
		}{
			{
				name:     "existing deployment",
				tenantID: "tenant-2",
				id:       "deploy-2",
				wantErr:  false,
			},
			{
				name:     "wrong tenant",
				tenantID: "wrong-tenant",
				id:       "deploy-2",
				wantErr:  true,
			},
			{
				name:     "non-existent deployment",
				tenantID: "tenant-2",
				id:       "missing",
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := store.GetDeployment(ctx, tt.tenantID, tt.id)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.id, got.ID)
				assert.Equal(t, tt.tenantID, got.TenantID)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		store.clearStore() // Reset state before update operations
		deployment := createTestDeployment("deploy-3", "tenant-3", "device-3")
		require.NoError(t, store.CreateDeployment(ctx, deployment))

		// Update deployment status
		deployment.Status = "completed"
		err := store.UpdateDeployment(ctx, deployment)
		require.NoError(t, err)

		// Verify update
		updated, err := store.GetDeployment(ctx, deployment.TenantID, deployment.ID)
		require.NoError(t, err)
		assert.Equal(t, "completed", updated.Status)

		// Test update of non-existent deployment
		deployment.ID = "missing"
		err = store.UpdateDeployment(ctx, deployment)
		require.Error(t, err)
	})

	t.Run("List", func(t *testing.T) {
		store.clearStore() // Reset state before list operations
		
		// Create test deployments with careful isolation
		deployments := []*config.Deployment{
			createTestDeployment("list-1", "tenant-list", "device-1"),
			createTestDeployment("list-2", "tenant-list", "device-1"),
			createTestDeployment("list-3", "tenant-other", "device-2"),
		}

		for _, d := range deployments {
			require.NoError(t, store.CreateDeployment(ctx, d))
		}

		tests := []struct {
			name    string
			opts    config.ListOptions
			want    int
			wantIDs []string
		}{
			{
				name:    "list all deployments",
				opts:    config.ListOptions{},
				want:    3,
				wantIDs: []string{"list-1", "list-2", "list-3"},
			},
			{
				name: "filter by tenant",
				opts: config.ListOptions{
					TenantID: "tenant-list",
				},
				want:    2,
				wantIDs: []string{"list-1", "list-2"},
			},
			{
				name: "filter by device",
				opts: config.ListOptions{
					DeviceID: "device-1",
				},
				want:    2,
				wantIDs: []string{"list-1", "list-2"},
			},
			{
				name: "pagination",
				opts: config.ListOptions{
					Offset: 1,
					Limit:  1,
				},
				want:    1,
				wantIDs: []string{"list-2"},
			},
			{
				name: "pagination - out of range",
				opts: config.ListOptions{
					Offset: 10,
					Limit:  1,
				},
				want:    0,
				wantIDs: []string{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := store.ListDeployments(ctx, tt.opts)
				require.NoError(t, err)
				assert.Len(t, got, tt.want)

				if tt.wantIDs != nil {
					var gotIDs []string
					for _, d := range got {
						gotIDs = append(gotIDs, d.ID)
					}
					assert.ElementsMatch(t, tt.wantIDs, gotIDs)
				}
			})
		}
	})
}

// createTestDeployment is a helper function that creates a deployment for testing
func createTestDeployment(id, tenantID, deviceID string) *config.Deployment {
	return &config.Deployment{
		ID:       id,
		TenantID: tenantID,
		DeviceID: deviceID,
		Status:   "pending",
	}
}
