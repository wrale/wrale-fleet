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

		// Verify hierarchy structure is complete
		root, err = store.Get(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.Len(t, root.Ancestry.Children, 2, "root should have two children")
	})

	t.Run("InvalidOperations", func(t *testing.T) {
		// Test moving a group to a non-existent parent
		err := hierarchy.UpdateHierarchy(ctx, child1, "non-existent-id")
		assert.Error(t, err, "should error on non-existent parent")

		// Test moving a root group under its own descendant
		err = hierarchy.UpdateHierarchy(ctx, root, grandchild.ID)
		assert.Error(t, err, "should error on cyclic dependency")

		// Verify original hierarchy remains intact
		root, err = store.Get(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.Len(t, root.Ancestry.Children, 2, "root children should be unchanged")
	})

	t.Run("DepthValidation", func(t *testing.T) {
		// Create a chain of groups to test depth calculations
		current := root
		depthMap := make(map[string]int)
		depthMap[root.ID] = 0

		for i := 0; i < 5; i++ {
			child := group.New(tenantID, "Depth Test", group.TypeStatic)
			err := store.Create(ctx, child)
			require.NoError(t, err)

			err = hierarchy.UpdateHierarchy(ctx, child, current.ID)
			require.NoError(t, err)

			// Verify depth calculation
			updated, err := store.Get(ctx, tenantID, child.ID)
			require.NoError(t, err)
			assert.Equal(t, depthMap[current.ID]+1, updated.Ancestry.Depth,
				"depth should increment by 1 at each level")

			depthMap[child.ID] = updated.Ancestry.Depth
			current = child
		}
	})

	t.Run("PathConsistency", func(t *testing.T) {
		// Get updated group states
		root, err := store.Get(ctx, tenantID, root.ID)
		require.NoError(t, err)
		child1, err := store.Get(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		grandchild, err := store.Get(ctx, tenantID, grandchild.ID)
		require.NoError(t, err)

		// Verify path construction
		assert.Equal(t, "/"+root.ID, root.Ancestry.Path, "root path should be direct")
		assert.Equal(t, root.Ancestry.Path+"/"+child1.ID, child1.Ancestry.Path,
			"child path should include parent")
		assert.Equal(t, child1.Ancestry.Path+"/"+grandchild.ID, grandchild.Ancestry.Path,
			"grandchild path should include full ancestry")

		// Verify path parts match depth
		assert.Equal(t, len(root.Ancestry.PathParts), root.Ancestry.Depth+1)
		assert.Equal(t, len(child1.Ancestry.PathParts), child1.Ancestry.Depth+1)
		assert.Equal(t, len(grandchild.Ancestry.PathParts), grandchild.Ancestry.Depth+1)
	})

	t.Run("BulkMoves", func(t *testing.T) {
		// Create a new parent group
		newParent := group.New(tenantID, "New Parent", group.TypeStatic)
		err := store.Create(ctx, newParent)
		require.NoError(t, err)

		// Move child1 and all its descendants
		err = hierarchy.UpdateHierarchy(ctx, child1, newParent.ID)
		require.NoError(t, err)

		// Verify all paths and depths are updated
		child1, err = store.Get(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		assert.Equal(t, newParent.ID, child1.ParentID, "child1 should have new parent")
		assert.Equal(t, "/"+newParent.ID+"/"+child1.ID, child1.Ancestry.Path)

		grandchild, err = store.Get(ctx, tenantID, grandchild.ID)
		require.NoError(t, err)
		assert.Equal(t, "/"+newParent.ID+"/"+child1.ID+"/"+grandchild.ID,
			grandchild.Ancestry.Path, "grandchild path should reflect new ancestry")
	})
}

func TestHierarchyEdgeCases(t *testing.T) {
	ctx, hierarchy, store := setupTestHierarchy(t)
	tenantID := "test-tenant"

	t.Run("EmptyHierarchy", func(t *testing.T) {
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err, "empty hierarchy should be valid")
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

		// Verify each tenant's hierarchy is valid
		err = hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
		err = hierarchy.ValidateHierarchyIntegrity(ctx, otherTenantID)
		require.NoError(t, err)

		// Attempt cross-tenant operation
		err = hierarchy.UpdateHierarchy(ctx, group1, group2.ID)
		assert.Error(t, err, "cross-tenant hierarchy operations should fail")
	})

	t.Run("DeepHierarchy", func(t *testing.T) {
		var lastID string
		var lastGroup *group.Group
		pathParts := make([]string, 0)

		// Create a deep chain of groups and verify hierarchy at each step
		for i := 0; i < 10; i++ {
			newGroup := group.New(tenantID, "Deep Group", group.TypeStatic)
			err := store.Create(ctx, newGroup)
			require.NoError(t, err)

			if lastID != "" {
				err = hierarchy.UpdateHierarchy(ctx, newGroup, lastID)
				require.NoError(t, err)
			}

			pathParts = append(pathParts, newGroup.ID)
			expectedPath := "/" + join(pathParts, "/")

			// Verify group's ancestry after linking
			updatedGroup, err := store.Get(ctx, tenantID, newGroup.ID)
			require.NoError(t, err)
			assert.Equal(t, i, updatedGroup.Ancestry.Depth)
			assert.Equal(t, expectedPath, updatedGroup.Ancestry.Path)
			assert.Equal(t, i+1, len(updatedGroup.Ancestry.PathParts))

			lastID = newGroup.ID
			lastGroup = newGroup
		}

		// Verify integrity of entire hierarchy
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err, "deep hierarchy should maintain integrity")

		// Verify deepest node maintains correct ancestry
		finalGroup, err := store.Get(ctx, tenantID, lastGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, 9, finalGroup.Ancestry.Depth)
		assert.Equal(t, 10, len(finalGroup.Ancestry.PathParts))
	})
}

// join concatenates strings with a separator
func join(parts []string, sep string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += sep
		}
		result += part
	}
	return result
}
