package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/wrale-fleet/internal/fleet/device/store/memory"
	"github.com/wrale/wrale-fleet/internal/fleet/group"
)

func TestStore(t *testing.T) {
	// Create a new store with a device store
	deviceStore := devmem.New()
	store := New(deviceStore)

	// Test context with tenant ID
	ctx := context.Background()
	tenantID := "test-tenant"

	// Test group creation
	t.Run("Create", func(t *testing.T) {
		g := group.New(tenantID, "Test Group", group.TypeStatic)

		err := store.Create(ctx, g)
		require.NoError(t, err)

		// Verify group exists
		saved, err := store.Get(ctx, tenantID, g.ID)
		require.NoError(t, err)
		assert.Equal(t, g.Name, saved.Name)
	})

	// Test group retrieval
	t.Run("Get", func(t *testing.T) {
		g := group.New(tenantID, "Retrieval Test", group.TypeStatic)
		err := store.Create(ctx, g)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, tenantID, g.ID)
		require.NoError(t, err)
		assert.Equal(t, g.ID, retrieved.ID)
		assert.Equal(t, tenantID, retrieved.TenantID)
	})

	// Test group update
	t.Run("Update", func(t *testing.T) {
		g := group.New(tenantID, "Update Test", group.TypeStatic)
		err := store.Create(ctx, g)
		require.NoError(t, err)

		g.Name = "Updated Group"
		err = store.Update(ctx, g)
		require.NoError(t, err)

		updated, err := store.Get(ctx, tenantID, g.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Group", updated.Name)
	})

	// Test group deletion
	t.Run("Delete", func(t *testing.T) {
		g := group.New(tenantID, "Delete Test", group.TypeStatic)
		err := store.Create(ctx, g)
		require.NoError(t, err)

		err = store.Delete(ctx, tenantID, g.ID)
		require.NoError(t, err)

		_, err = store.Get(ctx, tenantID, g.ID)
		assert.Equal(t, group.ErrGroupNotFound, err)
	})
}

func TestHierarchy(t *testing.T) {
	deviceStore := devmem.New()
	store := New(deviceStore)
	ctx := context.Background()
	tenantID := "test-tenant"

	t.Run("HierarchyOperations", func(t *testing.T) {
		// Create root group
		root := group.New(tenantID, "Root", group.TypeStatic)
		err := store.Create(ctx, root)
		require.NoError(t, err)

		// Create child1 and set its parent
		child1 := group.New(tenantID, "Child 1", group.TypeStatic)
		err = store.Create(ctx, child1)
		require.NoError(t, err)

		// Set parent for child1
		err = child1.SetParent(root.ID, &root.Ancestry)
		require.NoError(t, err)
		err = store.Update(ctx, child1)
		require.NoError(t, err)

		// Create child2 and set its parent
		child2 := group.New(tenantID, "Child 2", group.TypeStatic)
		err = store.Create(ctx, child2)
		require.NoError(t, err)

		// Set parent for child2
		err = child2.SetParent(root.ID, &root.Ancestry)
		require.NoError(t, err)
		err = store.Update(ctx, child2)
		require.NoError(t, err)

		// Update root to include children
		root.AddChild(child1.ID)
		root.AddChild(child2.ID)
		err = store.Update(ctx, root)
		require.NoError(t, err)

		// Test GetChildren
		children, err := store.GetChildren(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.Len(t, children, 2)

		// Test ValidateHierarchy
		err = store.ValidateHierarchy(ctx, tenantID)
		require.NoError(t, err)

		// Verify ancestry paths
		assert.Equal(t, "/"+root.ID, root.Ancestry.Path)
		assert.Equal(t, "/"+root.ID+"/"+child1.ID, child1.Ancestry.Path)
		assert.Equal(t, "/"+root.ID+"/"+child2.ID, child2.Ancestry.Path)

		// Verify depths
		assert.Equal(t, 0, root.Ancestry.Depth)
		assert.Equal(t, 1, child1.Ancestry.Depth)
		assert.Equal(t, 1, child2.Ancestry.Depth)
	})
}
