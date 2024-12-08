package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/group"
)

func TestNew(t *testing.T) {
	ts := setupTest(t)
	assert.NotNil(t, ts.store.groups, "groups map should be initialized")
	assert.NotNil(t, ts.store.memberships, "memberships map should be initialized")
	assert.NotNil(t, ts.store.deviceStore, "device store should be set")
}

func TestStore_Create(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		group   *group.Group
		wantErr bool
	}{
		{
			name:    "valid static group",
			group:   ts.createTestGroup("test-1", "tenant-1", "Test Group", group.TypeStatic),
			wantErr: false,
		},
		{
			name: "valid dynamic group",
			group: ts.createTestDynamicGroup("test-2", "tenant-1", "Test Group", &group.MembershipQuery{
				Tags: map[string]string{"env": "prod"},
			}),
			wantErr: false,
		},
		{
			name:    "duplicate group",
			group:   ts.createTestGroup("test-1", "tenant-1", "Duplicate Group", group.TypeStatic),
			wantErr: true,
		},
		{
			name: "missing required fields",
			group: &group.Group{
				Name: "Invalid Group",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.store.Create(ctx, tt.group)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			stored, err := ts.store.Get(ctx, tt.group.TenantID, tt.group.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.group.Name, stored.Name)
			assert.Equal(t, tt.group.Type, stored.Type)
		})
	}
}

func TestStore_Get(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	group := ts.createTestGroup("test-1", "tenant-1", "Test Group", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, group))

	tests := []struct {
		name     string
		tenantID string
		groupID  string
		wantErr  bool
	}{
		{
			name:     "existing group",
			tenantID: "tenant-1",
			groupID:  "test-1",
			wantErr:  false,
		},
		{
			name:     "wrong tenant",
			tenantID: "wrong-tenant",
			groupID:  "test-1",
			wantErr:  true,
		},
		{
			name:     "non-existent group",
			tenantID: "tenant-1",
			groupID:  "missing",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ts.store.Get(ctx, tt.tenantID, tt.groupID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.groupID, got.ID)
			assert.Equal(t, tt.tenantID, got.TenantID)
		})
	}
}

func TestStore_Update(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	initial := ts.createTestGroup("test-1", "tenant-1", "Initial Name", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, initial))

	tests := []struct {
		name    string
		group   *group.Group
		wantErr bool
	}{
		{
			name:    "valid update",
			group:   ts.createTestGroup("test-1", "tenant-1", "Updated Name", group.TypeStatic),
			wantErr: false,
		},
		{
			name:    "non-existent group",
			group:   ts.createTestGroup("missing", "tenant-1", "Missing Group", group.TypeStatic),
			wantErr: true,
		},
		{
			name:    "wrong tenant",
			group:   ts.createTestGroup("test-1", "wrong-tenant", "Wrong Tenant", group.TypeStatic),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.store.Update(ctx, tt.group)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			updated, err := ts.store.Get(ctx, tt.group.TenantID, tt.group.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.group.Name, updated.Name)
		})
	}
}

func TestStore_Delete(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	group := ts.createTestGroup("test-1", "tenant-1", "Test Group", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, group))

	tests := []struct {
		name     string
		tenantID string
		groupID  string
		wantErr  bool
	}{
		{
			name:     "existing group",
			tenantID: "tenant-1",
			groupID:  "test-1",
			wantErr:  false,
		},
		{
			name:     "non-existent group",
			tenantID: "tenant-1",
			groupID:  "missing",
			wantErr:  true,
		},
		{
			name:     "wrong tenant",
			tenantID: "wrong-tenant",
			groupID:  "test-1",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.store.Delete(ctx, tt.tenantID, tt.groupID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			_, err = ts.store.Get(ctx, tt.tenantID, tt.groupID)
			require.Error(t, err)
		})
	}
}
