package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/config"
)

func TestTemplate_CRUD(t *testing.T) {
	store := New()
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			template *config.Template
			wantErr  bool
		}{
			{
				name:     "valid template",
				template: createTestTemplate("test-1", "tenant-1"),
				wantErr:  false,
			},
			{
				name:     "duplicate template",
				template: createTestTemplate("test-1", "tenant-1"),
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := store.CreateTemplate(ctx, tt.template)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				stored, err := store.GetTemplate(ctx, tt.template.TenantID, tt.template.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.template.Name, stored.Name)
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		template := createTestTemplate("test-2", "tenant-2")
		require.NoError(t, store.CreateTemplate(ctx, template))

		tests := []struct {
			name     string
			tenantID string
			id       string
			wantErr  bool
		}{
			{
				name:     "existing template",
				tenantID: "tenant-2",
				id:       "test-2",
				wantErr:  false,
			},
			{
				name:     "wrong tenant",
				tenantID: "wrong-tenant",
				id:       "test-2",
				wantErr:  true,
			},
			{
				name:     "non-existent template",
				tenantID: "tenant-2",
				id:       "missing",
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := store.GetTemplate(ctx, tt.tenantID, tt.id)
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
		template := createTestTemplate("test-3", "tenant-3")
		require.NoError(t, store.CreateTemplate(ctx, template))

		tests := []struct {
			name     string
			template *config.Template
			wantErr  bool
		}{
			{
				name: "valid update",
				template: func() *config.Template {
					t := createTestTemplate("test-3", "tenant-3")
					t.Name = "Updated Name"
					return t
				}(),
				wantErr: false,
			},
			{
				name:     "non-existent template",
				template: createTestTemplate("missing", "tenant-3"),
				wantErr:  true,
			},
			{
				name:     "wrong tenant",
				template: createTestTemplate("test-3", "wrong-tenant"),
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := store.UpdateTemplate(ctx, tt.template)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				updated, err := store.GetTemplate(ctx, tt.template.TenantID, tt.template.ID)
				require.NoError(t, err)
				assert.Equal(t, tt.template.Name, updated.Name)
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		template := createTestTemplate("test-4", "tenant-4")
		require.NoError(t, store.CreateTemplate(ctx, template))

		tests := []struct {
			name     string
			tenantID string
			id       string
			wantErr  bool
		}{
			{
				name:     "existing template",
				tenantID: "tenant-4",
				id:       "test-4",
				wantErr:  false,
			},
			{
				name:     "non-existent template",
				tenantID: "tenant-4",
				id:       "missing",
				wantErr:  true,
			},
			{
				name:     "wrong tenant",
				tenantID: "wrong-tenant",
				id:       "test-4",
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := store.DeleteTemplate(ctx, tt.tenantID, tt.id)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				_, err = store.GetTemplate(ctx, tt.tenantID, tt.id)
				require.Error(t, err)
			})
		}
	})

	t.Run("List", func(t *testing.T) {
		// Create fresh templates for listing tests
		templates := []*config.Template{
			createTestTemplate("list-1", "tenant-list"),
			createTestTemplate("list-2", "tenant-list"),
			createTestTemplate("list-3", "tenant-other"),
		}

		for _, tmpl := range templates {
			require.NoError(t, store.CreateTemplate(ctx, tmpl))
		}

		tests := []struct {
			name    string
			opts    config.ListOptions
			want    int
			wantIDs []string
		}{
			{
				name:    "list all templates",
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
				got, err := store.ListTemplates(ctx, tt.opts)
				require.NoError(t, err)
				assert.Len(t, got, tt.want)

				if tt.wantIDs != nil {
					var gotIDs []string
					for _, tmpl := range got {
						gotIDs = append(gotIDs, tmpl.ID)
					}
					assert.ElementsMatch(t, tt.wantIDs, gotIDs)
				}
			})
		}
	})
}
