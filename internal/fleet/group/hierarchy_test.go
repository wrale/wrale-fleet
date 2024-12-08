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

// TestHierarchy verifies hierarchy management functionality
func TestHierarchy(t *testing.T) {
	// Initialize stores
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)

	ctx := context.Background()
	tenantID := "test-tenant"

	// Create test groups
	root := &group.Group{
		ID:       "root",
		TenantID: tenantID,
		Name:     "Root Group",
		Type:     group.TypeStatic,
	}

	child1 := &group.Group{
		ID:       "child1",
		TenantID: tenantID,
		Name:     "Child Group 1",
		Type:     group.TypeStatic,
		ParentID: root.ID,
	}

	child2 := &group.Group{
		ID:       "child2",
		TenantID: tenantID,
		Name:     "Child Group 2",
		Type:     group.TypeStatic,
		ParentID: root.ID,
	}

	grandchild := &group.Group{
		ID:       "grandchild",
		TenantID: tenantID,
		Name:     "Grandchild Group",
		Type:     group.TypeStatic,
		ParentID: child1.ID,
	}

	// Test hierarchy creation
	t.Run("CreateHierarchy", func(t *testing.T) {
		err := store.Create(ctx, root)
		require.NoError(t, err)

		err = store.Create(ctx, child1)
		require.NoError(t, err)

		err = store.Create(ctx, child2)
		require.NoError(t, err)

		err = store.Create(ctx, grandchild)
		require.NoError(t, err)
	})

	// Test getting children
	t.Run("GetChildren", func(t *testing.T) {
		children, err := store.GetChildren(ctx, tenantID, root.ID)
		require.NoError(t, err)
		assert.Len(t, children, 2)

		children, err = store.GetChildren(ctx, tenantID, child1.ID)
		require.NoError(t, err)
		assert.Len(t, children, 1)
		assert.Equal(t, grandchild.ID, children[0].ID)
	})

	// Test hierarchy validation
	t.Run("ValidateHierarchy", func(t *testing.T) {
		err := store.ValidateHierarchy(ctx, tenantID)
		require.NoError(t, err)

		// Test invalid parent reference
		invalid := &group.Group{
			ID:       "invalid",
			TenantID: tenantID,
			Name:     "Invalid Parent",
			Type:     group.TypeStatic,
			ParentID: "nonexistent",
		}
		err = store.Create(ctx, invalid)
		require.NoError(t, err)

		err = store.ValidateHierarchy(ctx, tenantID)
		assert.Error(t, err)
	})

	// Test cycle detection
	t.Run("DetectCycles", func(t *testing.T) {
		// Attempt to create a cycle
		root.ParentID = grandchild.ID
		err := store.Update(ctx, root)
		require.NoError(t, err)

		err = store.ValidateHierarchy(ctx, tenantID)
		assert.Error(t, err)
	})
}
