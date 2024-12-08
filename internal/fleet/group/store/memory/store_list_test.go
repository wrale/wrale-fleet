package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/group"
)

func TestStore_List(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	// Create test groups
	groups := []*group.Group{
		ts.createTestGroup("group-1", "tenant-1", "Group 1", group.TypeStatic),
		ts.createTestGroup("group-2", "tenant-1", "Group 2", group.TypeDynamic),
		ts.createTestGroup("group-3", "tenant-2", "Group 3", group.TypeStatic),
	}

	// Add metadata and parent relationships
	groups[0].Properties.Metadata["env"] = "prod"
	groups[1].ParentID = "group-1"
	groups[1].Properties.Metadata["env"] = "staging"
	groups[1].Query = &group.MembershipQuery{
		Tags: map[string]string{"env": "prod"},
	}

	for _, g := range groups {
		require.NoError(t, ts.store.Create(ctx, g))
	}

	tests := []struct {
		name    string
		opts    group.ListOptions
		want    int
		wantIDs []string
	}{
		{
			name: "list all groups",
			opts: group.ListOptions{},
			want: 3,
		},
		{
			name: "filter by tenant",
			opts: group.ListOptions{
				TenantID: "tenant-1",
			},
			want:    2,
			wantIDs: []string{"group-1", "group-2"},
		},
		{
			name: "filter by type",
			opts: group.ListOptions{
				Type: group.TypeStatic,
			},
			want:    2,
			wantIDs: []string{"group-1", "group-3"},
		},
		{
			name: "filter by parent",
			opts: group.ListOptions{
				ParentID: "group-1",
			},
			want:    1,
			wantIDs: []string{"group-2"},
		},
		{
			name: "filter by metadata",
			opts: group.ListOptions{
				Tags: map[string]string{"env": "prod"},
			},
			want:    1,
			wantIDs: []string{"group-1"},
		},
		{
			name: "pagination",
			opts: group.ListOptions{
				Offset: 1,
				Limit:  1,
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ts.store.List(ctx, tt.opts)
			require.NoError(t, err)
			assert.Len(t, got, tt.want)

			if tt.wantIDs != nil {
				var gotIDs []string
				for _, g := range got {
					gotIDs = append(gotIDs, g.ID)
				}
				assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			}
		})
	}
}
