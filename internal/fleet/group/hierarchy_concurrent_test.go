package group_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group"
	grpmem "github.com/wrale/fleet/internal/fleet/group/store/memory"
)

func TestConcurrentHierarchyOperations(t *testing.T) {
	ctx := context.Background()
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)
	hierarchy := group.NewHierarchyManager(store)
	tenantID := "test-tenant"

	// Create initial structure
	root := group.New(tenantID, "Root", group.TypeStatic)
	require.NoError(t, store.Create(ctx, root))

	children := make([]*group.Group, 5)
	for i := 0; i < 5; i++ {
		children[i] = group.New(tenantID, "Child", group.TypeStatic)
		require.NoError(t, store.Create(ctx, children[i]))
		require.NoError(t, hierarchy.UpdateHierarchy(ctx, children[i], root.ID))
	}

	t.Run("ConcurrentParentUpdates", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Attempt concurrent parent updates
		for i := 0; i < len(children); i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				// Move child between root and no parent repeatedly
				for j := 0; j < 5; j++ {
					parentID := ""
					if j%2 == 0 {
						parentID = root.ID
					}
					if err := hierarchy.UpdateHierarchy(ctx, children[idx], parentID); err != nil {
						errors <- err
						return
					}
					time.Sleep(time.Millisecond) // Small delay to increase contention
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			require.NoError(t, err)
		}

		// Verify hierarchy integrity after concurrent operations
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
	})

	t.Run("ConcurrentChildModification", func(t *testing.T) {
		parent := group.New(tenantID, "Parent", group.TypeStatic)
		require.NoError(t, store.Create(ctx, parent))

		var wg sync.WaitGroup
		errors := make(chan error, 10)

		// Concurrent addition/removal of children
		for i := 0; i < 5; i++ {
			wg.Add(2) // One goroutine to add, one to remove
			child := children[i]

			// Goroutine to add child
			go func() {
				defer wg.Done()
				if err := hierarchy.UpdateHierarchy(ctx, child, parent.ID); err != nil {
					errors <- err
				}
			}()

			// Goroutine to remove child
			go func() {
				defer wg.Done()
				if err := hierarchy.UpdateHierarchy(ctx, child, ""); err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			require.NoError(t, err)
		}

		// Verify final state
		err := hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)
	})
}

func TestHierarchyRecovery(t *testing.T) {
	ctx := context.Background()
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)
	hierarchy := group.NewHierarchyManager(store)
	tenantID := "test-tenant"

	t.Run("PartialUpdateRecovery", func(t *testing.T) {
		// Create a deep hierarchy to test partial updates
		root := group.New(tenantID, "Root", group.TypeStatic)
		require.NoError(t, store.Create(ctx, root))

		current := root
		var groups []*group.Group
		for i := 0; i < 5; i++ {
			child := group.New(tenantID, "Child", group.TypeStatic)
			require.NoError(t, store.Create(ctx, child))
			require.NoError(t, hierarchy.UpdateHierarchy(ctx, child, current.ID))
			groups = append(groups, child)
			current = child
		}

		// Get initial state for verification
		initialPaths := make(map[string]string)
		for _, g := range groups {
			updated, err := store.Get(ctx, tenantID, g.ID)
			require.NoError(t, err)
			initialPaths[g.ID] = updated.Ancestry.Path
		}

		// Attempt to move entire chain to new parent
		newParent := group.New(tenantID, "NewParent", group.TypeStatic)
		require.NoError(t, store.Create(ctx, newParent))

		// Move the middle of the chain, which should fail due to validation
		err := hierarchy.UpdateHierarchy(ctx, groups[2], newParent.ID)
		assert.Error(t, err)

		// Verify hierarchy is still valid
		err = hierarchy.ValidateHierarchyIntegrity(ctx, tenantID)
		require.NoError(t, err)

		// Verify paths remained unchanged
		for _, g := range groups {
			updated, err := store.Get(ctx, tenantID, g.ID)
			require.NoError(t, err)
			assert.Equal(t, initialPaths[g.ID], updated.Ancestry.Path)
		}
	})
}

func TestHierarchyValidationEnhanced(t *testing.T) {
	ctx := context.Background()
	deviceStore := devmem.New()
	store := grpmem.New(deviceStore)
	hierarchy := group.NewHierarchyManager(store)

	t.Run("CrossTenantIsolation", func(t *testing.T) {
		tenant1 := "tenant-1"
		tenant2 := "tenant-2"

		// Create groups in different tenants
		root1 := group.New(tenant1, "Root", group.TypeStatic)
		root2 := group.New(tenant2, "Root", group.TypeStatic)
		require.NoError(t, store.Create(ctx, root1))
		require.NoError(t, store.Create(ctx, root2))

		child1 := group.New(tenant1, "Child", group.TypeStatic)
		require.NoError(t, store.Create(ctx, child1))
		require.NoError(t, hierarchy.UpdateHierarchy(ctx, child1, root1.ID))

		// Attempt cross-tenant hierarchy operations
		err := hierarchy.UpdateHierarchy(ctx, child1, root2.ID)
		assert.Error(t, err)

		// Verify both hierarchies remain valid
		require.NoError(t, hierarchy.ValidateHierarchyIntegrity(ctx, tenant1))
		require.NoError(t, hierarchy.ValidateHierarchyIntegrity(ctx, tenant2))

		// Verify child remains with original parent
		updated, err := store.Get(ctx, tenant1, child1.ID)
		require.NoError(t, err)
		assert.Equal(t, root1.ID, updated.ParentID)
	})

	t.Run("PathIntegrityCheck", func(t *testing.T) {
		tenantID := "test-tenant"
		root := group.New(tenantID, "Root", group.TypeStatic)
		require.NoError(t, store.Create(ctx, root))

		// Create a chain of 5 groups
		current := root
		for i := 0; i < 5; i++ {
			child := group.New(tenantID, "Child", group.TypeStatic)
			require.NoError(t, store.Create(ctx, child))
			require.NoError(t, hierarchy.UpdateHierarchy(ctx, child, current.ID))

			// Verify path integrity after each addition
			updated, err := store.Get(ctx, tenantID, child.ID)
			require.NoError(t, err)

			// Check path construction
			assert.Equal(t, current.Ancestry.Path+"/"+child.ID, updated.Ancestry.Path)
			assert.Equal(t, current.Ancestry.Depth+1, updated.Ancestry.Depth)
			assert.Equal(t, len(updated.Ancestry.PathParts), updated.Ancestry.Depth+1)

			current = child
		}
	})
}
