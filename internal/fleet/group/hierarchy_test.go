package group_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group"
	grpmem "github.com/wrale/fleet/internal/fleet/group/store/memory"
)

type testEnv struct {
	ctx       context.Context
	hierarchy *group.HierarchyManager
	store     group.Store
	tenantID  string
}

// setupTestEnv creates an isolated test environment
func setupTestEnv(t *testing.T) *testEnv {
	ctx := context.Background()
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)
	hierarchy := group.NewHierarchyManager(store)

	err := store.Clear(ctx)
	require.NoError(t, err)

	return &testEnv{
		ctx:       ctx,
		hierarchy: hierarchy,
		store:     store,
		tenantID:  "test-tenant",
	}
}

// verifyHierarchyState checks complete hierarchy state
func verifyHierarchyState(t *testing.T, env *testEnv, group *group.Group) {
	updated, err := env.store.Get(env.ctx, env.tenantID, group.ID)
	require.NoError(t, err)

	// Verify path construction
	pathParts := strings.Split(strings.TrimPrefix(updated.Ancestry.Path, "/"), "/")
	assert.Equal(t, pathParts, updated.Ancestry.PathParts)
	assert.Equal(t, len(pathParts)-1, updated.Ancestry.Depth)

	// Verify parent-child consistency
	if updated.ParentID != "" {
		parent, err := env.store.Get(env.ctx, env.tenantID, updated.ParentID)
		require.NoError(t, err)
		assert.Contains(t, parent.Ancestry.Children, updated.ID)
	}

	// Verify child relationships
	for _, childID := range updated.Ancestry.Children {
		child, err := env.store.Get(env.ctx, env.tenantID, childID)
		require.NoError(t, err)
		assert.Equal(t, updated.ID, child.ParentID)
		assert.Equal(t, updated.Ancestry.Path+"/"+childID, child.Ancestry.Path)
		assert.Equal(t, updated.Ancestry.Depth+1, child.Ancestry.Depth)
	}
}

func TestHierarchy(t *testing.T) {
	t.Run("BasicHierarchyOperations", func(t *testing.T) {
		env := setupTestEnv(t)

		// Create base hierarchy
		root := group.New(env.tenantID, "Root Group", group.TypeStatic)
		require.NoError(t, env.store.Create(env.ctx, root))
		verifyHierarchyState(t, env, root)

		// Add first level children
		child1 := group.New(env.tenantID, "Child Group 1", group.TypeStatic)
		child2 := group.New(env.tenantID, "Child Group 2", group.TypeStatic)

		require.NoError(t, env.store.Create(env.ctx, child1))
		require.NoError(t, env.store.Create(env.ctx, child2))

		require.NoError(t, env.hierarchy.UpdateHierarchy(env.ctx, child1, root.ID))
		require.NoError(t, env.hierarchy.UpdateHierarchy(env.ctx, child2, root.ID))

		verifyHierarchyState(t, env, root)
		verifyHierarchyState(t, env, child1)
		verifyHierarchyState(t, env, child2)

		// Add grandchild
		grandchild := group.New(env.tenantID, "Grandchild Group", group.TypeStatic)
		require.NoError(t, env.store.Create(env.ctx, grandchild))
		require.NoError(t, env.hierarchy.UpdateHierarchy(env.ctx, grandchild, child1.ID))

		verifyHierarchyState(t, env, grandchild)

		// Verify complete hierarchy
		require.NoError(t, env.hierarchy.ValidateHierarchyIntegrity(env.ctx, env.tenantID))
	})

	t.Run("HierarchyModificationScenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			setupFunc   func(*testEnv) (*group.Group, *group.Group, error)
			verifyFunc  func(*testing.T, *testEnv, *group.Group, *group.Group)
			expectError bool
		}{
			{
				name: "MoveToNewParent",
				setupFunc: func(env *testEnv) (*group.Group, *group.Group, error) {
					oldParent := group.New(env.tenantID, "Old Parent", group.TypeStatic)
					newParent := group.New(env.tenantID, "New Parent", group.TypeStatic)
					child := group.New(env.tenantID, "Child", group.TypeStatic)

					for _, g := range []*group.Group{oldParent, newParent, child} {
						if err := env.store.Create(env.ctx, g); err != nil {
							return nil, nil, err
						}
					}

					if err := env.hierarchy.UpdateHierarchy(env.ctx, child, oldParent.ID); err != nil {
						return nil, nil, err
					}

					return child, newParent, env.hierarchy.UpdateHierarchy(env.ctx, child, newParent.ID)
				},
				verifyFunc: func(t *testing.T, env *testEnv, child, newParent *group.Group) {
					updated, err := env.store.Get(env.ctx, env.tenantID, child.ID)
					require.NoError(t, err)
					assert.Equal(t, newParent.ID, updated.ParentID)
					assert.Equal(t, newParent.Ancestry.Path+"/"+child.ID, updated.Ancestry.Path)
				},
				expectError: false,
			},
			{
				name: "PreventCyclicDependency",
				setupFunc: func(env *testEnv) (*group.Group, *group.Group, error) {
					parent := group.New(env.tenantID, "Parent", group.TypeStatic)
					child := group.New(env.tenantID, "Child", group.TypeStatic)

					for _, g := range []*group.Group{parent, child} {
						if err := env.store.Create(env.ctx, g); err != nil {
							return nil, nil, err
						}
					}

					if err := env.hierarchy.UpdateHierarchy(env.ctx, child, parent.ID); err != nil {
						return nil, nil, err
					}

					return parent, child, env.hierarchy.UpdateHierarchy(env.ctx, parent, child.ID)
				},
				verifyFunc: func(t *testing.T, env *testEnv, parent, child *group.Group) {
					// Verify original hierarchy remains intact
					updated, err := env.store.Get(env.ctx, env.tenantID, parent.ID)
					require.NoError(t, err)
					assert.Empty(t, updated.ParentID)
					assert.Contains(t, updated.Ancestry.Children, child.ID)
				},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				env := setupTestEnv(t)
				subject, target, err := tc.setupFunc(env)
				if tc.expectError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				tc.verifyFunc(t, env, subject, target)
				require.NoError(t, env.hierarchy.ValidateHierarchyIntegrity(env.ctx, env.tenantID))
			})
		}
	})

	t.Run("DeepHierarchyOperations", func(t *testing.T) {
		env := setupTestEnv(t)

		// Create and validate a deep hierarchy
		groups := make([]*group.Group, 10)
		for i := 0; i < len(groups); i++ {
			groups[i] = group.New(env.tenantID, "Level", group.TypeStatic)
			require.NoError(t, env.store.Create(env.ctx, groups[i]))

			if i > 0 {
				require.NoError(t, env.hierarchy.UpdateHierarchy(env.ctx, groups[i], groups[i-1].ID))
			}

			verifyHierarchyState(t, env, groups[i])
		}

		// Attempt bulk move of a subtree
		newParent := group.New(env.tenantID, "New Parent", group.TypeStatic)
		require.NoError(t, env.store.Create(env.ctx, newParent))

		// Move middle of chain to new parent (should move entire subtree)
		midPoint := len(groups) / 2
		err := env.hierarchy.UpdateHierarchy(env.ctx, groups[midPoint], newParent.ID)
		require.NoError(t, err)

		// Verify entire subtree moved correctly
		for i := midPoint; i < len(groups); i++ {
			updated, err := env.store.Get(env.ctx, env.tenantID, groups[i].ID)
			require.NoError(t, err)

			if i == midPoint {
				assert.Equal(t, newParent.ID, updated.ParentID)
				assert.Equal(t, newParent.Ancestry.Path+"/"+updated.ID, updated.Ancestry.Path)
			} else {
				previousGroup, err := env.store.Get(env.ctx, env.tenantID, groups[i-1].ID)
				require.NoError(t, err)
				assert.Equal(t, groups[i-1].ID, updated.ParentID)
				assert.Equal(t, previousGroup.Ancestry.Path+"/"+updated.ID, updated.Ancestry.Path)
			}
		}

		require.NoError(t, env.hierarchy.ValidateHierarchyIntegrity(env.ctx, env.tenantID))
	})
}
