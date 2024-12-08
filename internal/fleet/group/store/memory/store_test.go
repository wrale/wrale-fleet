package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group"
)

func TestStore(t *testing.T) {
	// Create a new store with a device store
	deviceStore := devmem.New()
	store := New(deviceStore)

	// Test context with tenant ID
	ctx := context.Background()
	tenantID := "test-tenant"
	groupID := "test-group"

	// Test group creation
	t.Run("Create", func(t *testing.T) {
		g := &group.Group{
			ID:       groupID,
			TenantID: tenantID,
			Name:     "Test Group",
			Type:     group.TypeStatic,
		}

		err := store.Create(ctx, g)
		require.NoError(t, err)

		// Verify group exists
		saved, err := store.Get(ctx, tenantID, groupID)
		require.NoError(t, err)
		assert.Equal(t, g.Name, saved.Name)
	})

	// Test group retrieval
	t.Run("Get", func(t *testing.T) {
		g, err := store.Get(ctx, tenantID, groupID)
		require.NoError(t, err)
		assert.Equal(t, groupID, g.ID)
		assert.Equal(t, tenantID, g.TenantID)
	})

	// Test group update
	t.Run("Update", func(t *testing.T) {
		g, err := store.Get(ctx, tenantID, groupID)
		require.NoError(t, err)

		g.Name = "Updated Group"
		err = store.Update(ctx, g)
		require.NoError(t, err)

		updated, err := store.Get(ctx, tenantID, groupID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Group", updated.Name)
	})

	// Test group deletion
	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(ctx, tenantID, groupID)
		require.NoError(t, err)

		_, err = store.Get(ctx, tenantID, groupID)
		assert.Equal(t, group.ErrGroupNotFound, err)
	})
}

func TestHierarchy(t *testing.T) {
	deviceStore := devmem.New()
	store := New(deviceStore)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Create a hierarchy of groups
	root := &group.Group{
		ID:       "root",
		TenantID: tenantID,
		Name:     "Root",
		Type:     group.TypeStatic,
	}

	child1 := &group.Group{
		ID:       "child1",
		TenantID: tenantID,
		Name:     "Child 1",
		Type:     group.TypeStatic,
		ParentID: "root",
	}

	child2 := &group.Group{
		ID:       "child2",
		TenantID: tenantID,
		Name:     "Child 2",
		Type:     group.TypeStatic,
		ParentID: "root",
	}

	// Test hierarchy operations
	t.Run("HierarchyOperations", func(t *testing.T) {
		err := store.Create(ctx, root)
		require.NoError(t, err)

		err = store.Create(ctx, child1)
		require.NoError(t, err)

		err = store.Create(ctx, child2)
		require.NoError(t, err)

		// Test GetChildren
		children, err := store.GetChildren(ctx, tenantID, "root")
		require.NoError(t, err)
		assert.Len(t, children, 2)

		// Test ValidateHierarchy
		err = store.ValidateHierarchy(ctx, tenantID)
		require.NoError(t, err)
	})
}