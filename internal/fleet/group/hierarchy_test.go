package group_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group"
	grpmem "github.com/wrale/fleet/internal/fleet/group/store/memory"
)

// setupTestHierarchy creates a fresh test environment
func setupTestHierarchy(t *testing.T) (context.Context, *group.HierarchyManager, group.Store) {
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)
	hierarchy := group.NewHierarchyManager(store)

	ctx := context.Background()
	err := store.Clear(ctx)
	require.NoError(t, err)

	return ctx, hierarchy, store
}

func TestHierarchy(t *testing.T) {
	ctx, hierarchy, store := setupTestHierarchy(t)
	tenantID := "test-tenant"

	// Create test groups using proper initialization
	root := group.New(tenantID, "Root Group", group.TypeStatic)
	child1 := group.New(tenantID, "Child Group 1", group.TypeStatic)
	child2 := group.New(tenantID, "Child Group 2", group.TypeStatic)
	grandchild := group.New(tenantID, "Grandchild Group", group.TypeStatic)

	// Test hierarchy creation
	t.Run("CreateHierarchy", func(t *testing.T) {
		// Create root first
		err := store.Create(ctx, root)
		require.NoError(t, err)

		// Create and link child1
		err = store.Create(ctx, child1)
		require.NoError(t, err)
		err = hierarchy.UpdateHierarchy(ctx, child1, root.ID)
		require.NoError(t, err)

		// Create and link child2
		err = store.Create(ctx, child2)
		require.NoError(t, err)
		err = hierarchy.UpdateHierarchy(ctx, child2, root.ID)
		require.NoError(t, err)

		// Create and link grandchild
		err = store.Create(ctx, grandchild)
		require.NoError(t, err)
		err = hierarchy.UpdateHierarchy(ctx, grandchild, child1.ID)
		require.NoError(t, err)
	})

	// Test getting children
	t.Run("GetChildren", func(t *testing.T) {
		// Verify root's children
		root, err := store.Get(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.Len(t, root.Ancestry.Children, 2)
		assert.Contains(t, root.Ancestry.Children, child1.ID)
		assert.Contains(t, root.Ancestry.Children, child2.ID)

		// Verify child1's children
		child1Updated, err := store.Get(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		assert.Len(t, child1Updated.Ancestry.Children, 1)
		assert.Contains(t, child1Updated.Ancestry.Children, grandchild.ID)
	})

	// Test hierarchy validation
	t.Run("ValidateHierarchy", func(t *testing.T) {
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
	})

	// Test cycle detection
	t.Run("DetectCycles", func(t *testing.T) {
		err := hierarchy.UpdateHierarchy(ctx, root, grandchild.ID)
		assert.Error(t, err) // Should detect cycle and fail
	})

	// Test ancestor checks
	t.Run("AncestorChecks", func(t *testing.T) {
		grandchildUpdated, err := store.Get(ctx, tenantID, grandchild.ID)
		require.NoError(t, err)

		assert.True(t, grandchildUpdated.IsAncestor(root.ID))
		assert.True(t, grandchildUpdated.IsAncestor(child1.ID))
		assert.False(t, grandchildUpdated.IsAncestor(child2.ID))
		assert.False(t, grandchildUpdated.IsAncestor(grandchild.ID))

		expectedPath := "/" + root.ID + "/" + child1.ID + "/" + grandchild.ID
		assert.Equal(t, expectedPath, grandchildUpdated.Ancestry.Path)
		assert.Equal(t, 2, grandchildUpdated.Ancestry.Depth)
	})

	// Test moving nodes
	t.Run("MoveNodes", func(t *testing.T) {
		err := hierarchy.UpdateHierarchy(ctx, grandchild, child2.ID)
		require.NoError(t, err)

		child1Updated, err := store.Get(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		assert.NotContains(t, child1Updated.Ancestry.Children, grandchild.ID)

		child2Updated, err := store.Get(ctx, tenantID, child2.ID)
		require.NoError(t, err)
		assert.Contains(t, child2Updated.Ancestry.Children, grandchild.ID)

		grandchildUpdated, err := store.Get(ctx, tenantID, grandchild.ID)
		require.NoError(t, err)
		expectedPath := "/" + root.ID + "/" + child2.ID + "/" + grandchild.ID
		assert.Equal(t, expectedPath, grandchildUpdated.Ancestry.Path)
		assert.Equal(t, child2.ID, grandchildUpdated.ParentID)
	})

	// Test making root node
	t.Run("MakeRoot", func(t *testing.T) {
		err := hierarchy.UpdateHierarchy(ctx, child1, "")
		require.NoError(t, err)

		child1Updated, err := store.Get(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		assert.Empty(t, child1Updated.ParentID)
		assert.Equal(t, "/"+child1.ID, child1Updated.Ancestry.Path)
		assert.Equal(t, 0, child1Updated.Ancestry.Depth)

		rootUpdated, err := store.Get(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.NotContains(t, rootUpdated.Ancestry.Children, child1.ID)
	})
}

func TestHierarchyEdgeCases(t *testing.T) {
	ctx, hierarchy, store := setupTestHierarchy(t)
	tenantID := "test-tenant"

	t.Run("EmptyHierarchy", func(t *testing.T) {
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
	})

	t.Run("SingleNode", func(t *testing.T) {
		single := group.New(tenantID, "Single Node", group.TypeStatic)
		err := store.Create(ctx, single)
		require.NoError(t, err)

		assert.Empty(t, single.ParentID)
		assert.Equal(t, "/"+single.ID, single.Ancestry.Path)
		assert.Equal(t, 0, single.Ancestry.Depth)
	})

	t.Run("CrossTenantHierarchy", func(t *testing.T) {
		otherTenantID := "other-tenant"

		group1 := group.New(tenantID, "Group 1", group.TypeStatic)
		group2 := group.New(otherTenantID, "Group 2", group.TypeStatic)

		err := store.Create(ctx, group1)
		require.NoError(t, err)
		err = store.Create(ctx, group2)
		require.NoError(t, err)

		err = hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
		err = hierarchy.ValidateHierarchyIntegrity(ctx, otherTenantID)
		require.NoError(t, err)
	})

	t.Run("DeepHierarchy", func(t *testing.T) {
		var lastID string
		var lastGroup *group.Group

		// Create a deep chain of groups and verify hierarchy at each step
		for i := 0; i < 10; i++ {
			newGroup := group.New(tenantID, "Deep Group", group.TypeStatic)
			err := store.Create(ctx, newGroup)
			require.NoError(t, err)

			if lastID != "" {
				err = hierarchy.UpdateHierarchy(ctx, newGroup, lastID)
				require.NoError(t, err)

				// Verify group's ancestry after linking
				updatedGroup, err := store.Get(ctx, tenantID, newGroup.ID)
				require.NoError(t, err)
				assert.Equal(t, i, updatedGroup.Ancestry.Depth)
				assert.Equal(t, i+1, len(updatedGroup.Ancestry.PathParts))
			}

			lastID = newGroup.ID
			lastGroup = newGroup
		}

		// Verify deepest node has correct ancestry
		finalGroup, err := store.Get(ctx, tenantID, lastGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, 9, finalGroup.Ancestry.Depth)
		assert.Equal(t, 10, len(finalGroup.Ancestry.PathParts))
	})
}
