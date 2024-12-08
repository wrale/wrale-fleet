package group

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test creating a new group
	tenantID := "test-tenant"
	name := "test-group"
	groupType := TypeStatic

	group := New(tenantID, name, groupType)

	require.NotEmpty(t, group.ID, "group ID should be generated")
	require.Equal(t, tenantID, group.TenantID, "tenant ID should match")
	require.Equal(t, name, group.Name, "group name should match")
	require.Equal(t, groupType, group.Type, "group type should match")
	assert.NotNil(t, group.Properties.Metadata, "metadata map should be initialized")
	assert.False(t, group.CreatedAt.IsZero(), "created timestamp should be set")
	assert.False(t, group.UpdatedAt.IsZero(), "updated timestamp should be set")

	// Verify ancestry initialization
	assert.Equal(t, "/"+group.ID, group.Ancestry.Path)
	assert.Equal(t, []string{group.ID}, group.Ancestry.PathParts)
	assert.Equal(t, 0, group.Ancestry.Depth)
	assert.Empty(t, group.Ancestry.Children)
}

func TestGroup_Validate(t *testing.T) {
	tests := []struct {
		name    string
		group   *Group
		wantErr bool
	}{
		{
			name: "valid static group",
			group: &Group{
				ID:       "test-id",
				TenantID: "test-tenant",
				Name:     "test-group",
				Type:     TypeStatic,
				Ancestry: AncestryInfo{
					Path:      "/test-id",
					PathParts: []string{"test-id"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid dynamic group with query",
			group: &Group{
				ID:       "test-id",
				TenantID: "test-tenant",
				Name:     "test-group",
				Type:     TypeDynamic,
				Query: &MembershipQuery{
					Tags: map[string]string{"env": "prod"},
				},
				Ancestry: AncestryInfo{
					Path:      "/test-id",
					PathParts: []string{"test-id"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			group: &Group{
				TenantID: "test-tenant",
				Name:     "test-group",
				Type:     TypeStatic,
			},
			wantErr: true,
		},
		{
			name: "missing tenant ID",
			group: &Group{
				ID:   "test-id",
				Name: "test-group",
				Type: TypeStatic,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			group: &Group{
				ID:       "test-id",
				TenantID: "test-tenant",
				Type:     TypeStatic,
			},
			wantErr: true,
		},
		{
			name: "missing type",
			group: &Group{
				ID:       "test-id",
				TenantID: "test-tenant",
				Name:     "test-group",
			},
			wantErr: true,
		},
		{
			name: "dynamic group without query",
			group: &Group{
				ID:       "test-id",
				TenantID: "test-tenant",
				Name:     "test-group",
				Type:     TypeDynamic,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, ErrCodeInvalidGroup, err.(*Error).Code)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGroup_SetQuery(t *testing.T) {
	tests := []struct {
		name    string
		group   *Group
		query   *MembershipQuery
		wantErr bool
	}{
		{
			name: "set query on dynamic group",
			group: &Group{
				Type: TypeDynamic,
			},
			query: &MembershipQuery{
				Tags: map[string]string{"env": "prod"},
			},
			wantErr: false,
		},
		{
			name: "set query with multiple criteria",
			group: &Group{
				Type: TypeDynamic,
			},
			query: &MembershipQuery{
				Tags:    map[string]string{"env": "prod", "region": "us-west"},
				Status:  "online",
				Regions: []string{"us-west-1", "us-west-2"},
			},
			wantErr: false,
		},
		{
			name: "set query on static group",
			group: &Group{
				Type: TypeStatic,
			},
			query: &MembershipQuery{
				Tags: map[string]string{"env": "prod"},
			},
			wantErr: true,
		},
		{
			name: "set nil query",
			group: &Group{
				Type: TypeDynamic,
			},
			query:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalUpdate := tt.group.UpdatedAt
			time.Sleep(time.Millisecond) // Ensure timestamp changes

			err := tt.group.SetQuery(tt.query)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, ErrCodeInvalidOperation, err.(*Error).Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.query, tt.group.Query)
			assert.True(t, tt.group.UpdatedAt.After(originalUpdate))
		})
	}
}

func TestGroup_SetParent(t *testing.T) {
	tests := []struct {
		name       string
		group      *Group
		parentID   string
		parentInfo *AncestryInfo
		wantErr    bool
		wantPath   string
		wantDepth  int
		wantParts  []string
	}{
		{
			name:     "set valid parent",
			group:    &Group{ID: "child"},
			parentID: "parent",
			parentInfo: &AncestryInfo{
				Path:      "/parent",
				PathParts: []string{"parent"},
				Depth:     0,
			},
			wantErr:   false,
			wantPath:  "/parent/child",
			wantDepth: 1,
			wantParts: []string{"parent", "child"},
		},
		{
			name:       "clear parent",
			group:      &Group{ID: "child"},
			parentID:   "",
			parentInfo: nil,
			wantErr:    false,
			wantPath:   "/child",
			wantDepth:  0,
			wantParts:  []string{"child"},
		},
		{
			name:       "missing parent info",
			group:      &Group{ID: "child"},
			parentID:   "parent",
			parentInfo: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalUpdate := tt.group.UpdatedAt
			time.Sleep(time.Millisecond) // Ensure timestamp changes

			err := tt.group.SetParent(tt.parentID, tt.parentInfo)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, ErrCodeInvalidOperation, err.(*Error).Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.parentID, tt.group.ParentID)
			assert.Equal(t, tt.wantPath, tt.group.Ancestry.Path)
			assert.Equal(t, tt.wantDepth, tt.group.Ancestry.Depth)
			assert.Equal(t, tt.wantParts, tt.group.Ancestry.PathParts)
			assert.True(t, tt.group.UpdatedAt.After(originalUpdate))
		})
	}
}

func TestGroup_UpdateProperties(t *testing.T) {
	group := New("test-tenant", "test-group", TypeStatic)
	originalUpdate := group.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure timestamp changes

	newProps := Properties{
		ConfigTemplate: json.RawMessage(`{"setting": "value"}`),
		PolicyOverrides: map[string]json.RawMessage{
			"policy1": json.RawMessage(`{"enabled": true}`),
		},
		Metadata: map[string]string{
			"environment": "production",
			"owner":       "platform-team",
		},
	}

	err := group.UpdateProperties(newProps)
	require.NoError(t, err)

	assert.Equal(t, newProps.ConfigTemplate, group.Properties.ConfigTemplate)
	assert.Equal(t, newProps.PolicyOverrides, group.Properties.PolicyOverrides)
	assert.Equal(t, newProps.Metadata, group.Properties.Metadata)
	assert.True(t, group.UpdatedAt.After(originalUpdate))
}

func TestGroup_IsAncestor(t *testing.T) {
	tests := []struct {
		name       string
		group      *Group
		ancestorID string
		want       bool
	}{
		{
			name: "direct parent",
			group: &Group{
				ID:       "child",
				ParentID: "parent",
				Ancestry: AncestryInfo{
					Path:      "/parent/child",
					PathParts: []string{"parent", "child"},
					Depth:     1,
				},
			},
			ancestorID: "parent",
			want:       true,
		},
		{
			name: "not parent",
			group: &Group{
				ID:       "child",
				ParentID: "parent",
				Ancestry: AncestryInfo{
					Path:      "/parent/child",
					PathParts: []string{"parent", "child"},
					Depth:     1,
				},
			},
			ancestorID: "other",
			want:       false,
		},
		{
			name: "no parent",
			group: &Group{
				ID: "child",
				Ancestry: AncestryInfo{
					Path:      "/child",
					PathParts: []string{"child"},
					Depth:     0,
				},
			},
			ancestorID: "parent",
			want:       false,
		},
		{
			name: "self reference",
			group: &Group{
				ID:       "group1",
				ParentID: "parent",
				Ancestry: AncestryInfo{
					Path:      "/parent/group1",
					PathParts: []string{"parent", "group1"},
					Depth:     1,
				},
			},
			ancestorID: "group1",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.IsAncestor(tt.ancestorID)
			assert.Equal(t, tt.want, got)
		})
	}
}
