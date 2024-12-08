package group_test

import (
	"context"
	"sync"
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

func TestConcurrentHierarchyOperations(t *testing.T) {
	t.Run("ConcurrentParentUpdates", func(t *testing.T) {
		env := setupTestEnv(t)

		// Create root node
		root := group.New(env.tenantID, "Root", group.TypeStatic)
		require.NoError(t, env.store.Create(env.ctx, root))

		// Create child nodes
		children := make([]*group.Group, 5)
		for i := 0; i < len(children); i++ {
			children[i] = group.New(env.tenantID, "Child", group.TypeStatic)
			require.NoError(t, env.store.Create(env.ctx, children[i]))
			require.NoError(t, env.hierarchy.UpdateHierarchy(env.ctx, children[i], root.ID))
		}

		var wg sync.WaitGroup
		var startWg sync.WaitGroup
		errChan := make(chan error, len(children)*5) // Size based on total operations

		// Setup synchronization barrier
		startWg.Add(1)

		// Launch concurrent updates
		for i := 0; i < len(children); i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				startWg.Wait() // Wait for signal to start

				for j := 0; j < 5; j++ {
					parentID := ""
					if j%2 == 0 {
						parentID = root.ID
					}
					if err := env.hierarchy.UpdateHierarchy(env.ctx, children[idx], parentID); err != nil {
						errChan <- err
						return
					}
				}
			}(i)
		}

		// Start all goroutines simultaneously
		startWg.Done()
		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			assert.NoError(t, err)
		}

		// Verify hierarchy integrity
		require.NoError(t, env.hierarchy.ValidateHierarchyIntegrity(env.ctx, env.tenantID))

		// Verify final state of each child
		for _, child := range children {
			updated, err := env.store.Get(env.ctx, env.tenantID, child.ID)
			require.NoError(t, err)
			assert.NotNil(t, updated)

			// Verify path consistency based on final parent state
			if updated.ParentID != "" {
				assert.Equal(t, root.ID, updated.ParentID)
				assert.Equal(t, "/"+root.ID+"/"+updated.ID, updated.Ancestry.Path)
				assert.Equal(t, 1, updated.Ancestry.Depth)
				assert.Equal(t, []string{root.ID, updated.ID}, updated.Ancestry.PathParts)
			} else {
				assert.Equal(t, "/"+updated.ID, updated.Ancestry.Path)
				assert.Equal(t, 0, updated.Ancestry.Depth)
				assert.Equal(t, []string{updated.ID}, updated.Ancestry.PathParts)
			}
		}
	})

	t.Run("ConcurrentChildModification", func(t *testing.T) {
		env := setupTestEnv(t)

		// Create test structure
		parent := group.New(env.tenantID, "Parent", group.TypeStatic)
		require.NoError(t, env.store.Create(env.ctx, parent))

		children := make([]*group.Group, 5)
		for i := 0; i < len(children); i++ {
			children[i] = group.New(env.tenantID, "Child", group.TypeStatic)
			require.NoError(t, env.store.Create(env.ctx, children[i]))
		}

		var wg sync.WaitGroup
		var startWg sync.WaitGroup
		errChan := make(chan error, len(children)*2) // Add/remove operations

		// Setup synchronization barrier
		startWg.Add(1)

		// Launch concurrent operations
		for i := 0; i < len(children); i++ {
			wg.Add(2)
			child := children[i]

			// Goroutine to add child
			go func() {
				defer wg.Done()
				startWg.Wait()
				if err := env.hierarchy.UpdateHierarchy(env.ctx, child, parent.ID); err != nil {
					errChan <- err
				}
			}()

			// Goroutine to remove child
			go func() {
				defer wg.Done()
				startWg.Wait()
				if err := env.hierarchy.UpdateHierarchy(env.ctx, child, ""); err != nil {
					errChan <- err
				}
			}()
		}

		// Start all goroutines simultaneously
		startWg.Done()
		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			assert.NoError(t, err)
		}

		// Verify hierarchy integrity
		require.NoError(t, env.hierarchy.ValidateHierarchyIntegrity(env.ctx, env.tenantID))

		// Verify final state of parent
		updatedParent, err := env.store.Get(env.ctx, env.tenantID, parent.ID)
		require.NoError(t, err)

		// Verify final states of children
		childrenSeen := make(map[string]bool)
		for _, child := range children {
			updated, err := env.store.Get(env.ctx, env.tenantID, child.ID)
			require.NoError(t, err)

			// Verify path and ancestry consistency
			if updated.ParentID == "" {
				assert.Equal(t, "/"+updated.ID, updated.Ancestry.Path)
				assert.Equal(t, 0, updated.Ancestry.Depth)
			} else {
				assert.Equal(t, parent.ID, updated.ParentID)
				assert.Equal(t, "/"+parent.ID+"/"+updated.ID, updated.Ancestry.Path)
				assert.Equal(t, 1, updated.Ancestry.Depth)
				childrenSeen[updated.ID] = true
			}
		}

		// Verify parent's children list matches actual child relationships
		for _, childID := range updatedParent.Ancestry.Children {
			assert.True(t, childrenSeen[childID], "Parent refers to child that doesn't belong to it")
			delete(childrenSeen, childID)
		}
		assert.Empty(t, childrenSeen, "Found children belonging to parent not in parent's child list")
	})
}
