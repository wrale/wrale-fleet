package group

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
)

func newTestStore() Store {
	deviceStore := devmem.New()
	return newMemoryStore(deviceStore)
}

func TestHierarchyManager_ValidateHierarchyChange(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(store Store) (*Group, string)
		expectError bool
		errorCode   string
	}{
		{
			name: "valid parent change",
			setupFunc: func(store Store) (*Group, string) {
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)
				require.NoError(t, store.Create(context.Background(), parent))
				require.NoError(t, store.Create(context.Background(), child))
				return child, parent.ID
			},
			expectError: false,
		},
		{
			name: "prevent self reference",
			setupFunc: func(store Store) (*Group, string) {
				group := New("tenant1", "group", TypeStatic)
				require.NoError(t, store.Create(context.Background(), group))
				return group, group.ID
			},
			expectError: true,
			errorCode:   ErrCodeCyclicDependency,
		},
		{
			name: "prevent cycle through descendants",
			setupFunc: func(store Store) (*Group, string) {
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)
				require.NoError(t, store.Create(context.Background(), parent))
				require.NoError(t, store.Create(context.Background(), child))

				manager := NewHierarchyManager(store)
				require.NoError(t, manager.UpdateHierarchy(context.Background(), child, parent.ID))

				return parent, child.ID
			},
			expectError: true,
			errorCode:   ErrCodeCyclicDependency,
		},
		{
			name: "moving to root is valid",
			setupFunc: func(store Store) (*Group, string) {
				group := New("tenant1", "group", TypeStatic)
				require.NoError(t, store.Create(context.Background(), group))
				return group, ""
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()
			manager := NewHierarchyManager(store)

			group, newParentID := tt.setupFunc(store)
			err := manager.ValidateHierarchyChange(context.Background(), group, newParentID)

			if tt.expectError {
				require.Error(t, err)
				var groupErr *Error
				require.ErrorAs(t, err, &groupErr)
				assert.Equal(t, tt.errorCode, groupErr.Code)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHierarchyManager_UpdateHierarchy(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(store Store) (*Group, string)
		validate  func(t *testing.T, store Store, group *Group, newParentID string)
	}{
		{
			name: "update to new parent",
			setupFunc: func(store Store) (*Group, string) {
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)
				require.NoError(t, store.Create(context.Background(), parent))
				require.NoError(t, store.Create(context.Background(), child))
				return child, parent.ID
			},
			validate: func(t *testing.T, store Store, group *Group, newParentID string) {
				// Verify group's ancestry
				assert.Equal(t, newParentID, group.ParentID)
				assert.Equal(t, 1, group.Ancestry.Depth)
				assert.Equal(t, 2, len(group.Ancestry.PathParts))
				assert.Equal(t, newParentID, group.Ancestry.PathParts[0])
				assert.Equal(t, group.ID, group.Ancestry.PathParts[1])

				// Verify parent's children
				parent, err := store.Get(context.Background(), group.TenantID, newParentID)
				require.NoError(t, err)
				assert.Contains(t, parent.Ancestry.Children, group.ID)
				assert.Equal(t, parent.ID, group.ParentID)
			},
		},
		{
			name: "move to root",
			setupFunc: func(store Store) (*Group, string) {
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)
				require.NoError(t, store.Create(context.Background(), parent))
				require.NoError(t, store.Create(context.Background(), child))

				manager := NewHierarchyManager(store)
				require.NoError(t, manager.UpdateHierarchy(context.Background(), child, parent.ID))

				return child, ""
			},
			validate: func(t *testing.T, store Store, group *Group, newParentID string) {
				// Verify group is now root
				assert.Empty(t, group.ParentID)
				assert.Equal(t, 0, group.Ancestry.Depth)
				assert.Equal(t, 1, len(group.Ancestry.PathParts))
				assert.Equal(t, group.ID, group.Ancestry.PathParts[0])

				// Verify old parent's children
				oldParent, err := store.List(context.Background(), ListOptions{TenantID: group.TenantID})
				require.NoError(t, err)
				for _, p := range oldParent {
					assert.NotContains(t, p.Ancestry.Children, group.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()
			manager := NewHierarchyManager(store)

			group, newParentID := tt.setupFunc(store)
			err := manager.UpdateHierarchy(context.Background(), group, newParentID)
			require.NoError(t, err)

			updatedGroup, err := store.Get(context.Background(), group.TenantID, group.ID)
			require.NoError(t, err)

			tt.validate(t, store, updatedGroup, newParentID)
		})
	}
}

func TestHierarchyManager_GetAncestors(t *testing.T) {
	ctx := context.Background()
	store := newTestStore()
	manager := NewHierarchyManager(store)

	// Create a hierarchy: root -> parent -> child
	root := New("tenant1", "root", TypeStatic)
	parent := New("tenant1", "parent", TypeStatic)
	child := New("tenant1", "child", TypeStatic)

	require.NoError(t, store.Create(ctx, root))
	require.NoError(t, store.Create(ctx, parent))
	require.NoError(t, store.Create(ctx, child))

	require.NoError(t, manager.UpdateHierarchy(ctx, parent, root.ID))
	require.NoError(t, manager.UpdateHierarchy(ctx, child, parent.ID))

	// Test getting ancestors
	ancestors, err := manager.GetAncestors(ctx, child)
	require.NoError(t, err)
	require.Len(t, ancestors, 2)
	assert.Equal(t, root.ID, ancestors[0].ID)
	assert.Equal(t, parent.ID, ancestors[1].ID)
}

func TestHierarchyManager_GetDescendants(t *testing.T) {
	ctx := context.Background()
	store := newTestStore()
	manager := NewHierarchyManager(store)

	// Create a hierarchy: root -> (child1, child2 -> grandchild)
	root := New("tenant1", "root", TypeStatic)
	child1 := New("tenant1", "child1", TypeStatic)
	child2 := New("tenant1", "child2", TypeStatic)
	grandchild := New("tenant1", "grandchild", TypeStatic)

	// First create all groups
	for _, g := range []*Group{root, child1, child2, grandchild} {
		require.NoError(t, store.Create(ctx, g))
	}

	// Setup hierarchy
	require.NoError(t, manager.UpdateHierarchy(ctx, child1, root.ID))
	require.NoError(t, manager.UpdateHierarchy(ctx, child2, root.ID))
	require.NoError(t, manager.UpdateHierarchy(ctx, grandchild, child2.ID))

	// Verify immediate children of root
	updatedRoot, err := store.Get(ctx, root.TenantID, root.ID)
	require.NoError(t, err)
	require.Len(t, updatedRoot.Ancestry.Children, 2, "root should have 2 children")
	assert.Contains(t, updatedRoot.Ancestry.Children, child1.ID)
	assert.Contains(t, updatedRoot.Ancestry.Children, child2.ID)

	// Verify child2's children
	updatedChild2, err := store.Get(ctx, child2.TenantID, child2.ID)
	require.NoError(t, err)
	require.Len(t, updatedChild2.Ancestry.Children, 1, "child2 should have 1 child")
	assert.Contains(t, updatedChild2.Ancestry.Children, grandchild.ID)

	// Test getting descendants
	descendants, err := manager.GetDescendants(ctx, root)
	require.NoError(t, err)
	require.Len(t, descendants, 3, "root should have 3 descendants")

	// Create map of descendants for easy checking
	descendantMap := make(map[string]*Group)
	for _, d := range descendants {
		descendantMap[d.ID] = d
	}

	assert.Contains(t, descendantMap, child1.ID)
	assert.Contains(t, descendantMap, child2.ID)
	assert.Contains(t, descendantMap, grandchild.ID)

	// Verify each descendant's parent-child relationships
	for _, d := range descendants {
		require.NotEmpty(t, d.ParentID, "descendant should have parent ID")
		parent, err := store.Get(ctx, d.TenantID, d.ParentID)
		require.NoError(t, err)
		assert.Contains(t, parent.Ancestry.Children, d.ID)
	}
}

func TestHierarchyManager_ValidateHierarchyIntegrity(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(store Store) error
		expectError bool
		errorCode   string
	}{
		{
			name: "valid hierarchy",
			setup: func(store Store) error {
				// Create test groups
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)

				// First create both groups
				if err := store.Create(context.Background(), parent); err != nil {
					return err
				}
				if err := store.Create(context.Background(), child); err != nil {
					return err
				}

				// Use HierarchyManager to set up the relationship
				manager := NewHierarchyManager(store)
				if err := manager.UpdateHierarchy(context.Background(), child, parent.ID); err != nil {
					return err
				}

				// Verify the setup worked
				updatedParent, err := store.Get(context.Background(), parent.TenantID, parent.ID)
				if err != nil {
					return err
				}
				if len(updatedParent.Ancestry.Children) != 1 {
					return fmt.Errorf("parent should have 1 child, got %d", len(updatedParent.Ancestry.Children))
				}

				return nil
			},
			expectError: false,
		},
		{
			name: "missing parent reference",
			setup: func(store Store) error {
				child := New("tenant1", "child", TypeStatic)
				child.ParentID = "non-existent"
				return store.Create(context.Background(), child)
			},
			expectError: true,
			errorCode:   ErrCodeInvalidGroup,
		},
		{
			name: "inconsistent parent-child relationship",
			setup: func(store Store) error {
				parent := New("tenant1", "parent", TypeStatic)
				child := New("tenant1", "child", TypeStatic)

				if err := store.Create(context.Background(), parent); err != nil {
					return err
				}
				if err := store.Create(context.Background(), child); err != nil {
					return err
				}

				parent.Ancestry.Children = append(parent.Ancestry.Children, child.ID)
				return store.Update(context.Background(), parent)
			},
			expectError: true,
			errorCode:   ErrCodeInvalidGroup,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newTestStore()
			manager := NewHierarchyManager(store)

			require.NoError(t, tt.setup(store))

			err := manager.ValidateHierarchyIntegrity(context.Background(), "tenant1")

			if tt.expectError {
				require.Error(t, err)
				var groupErr *Error
				require.ErrorAs(t, err, &groupErr)
				assert.Equal(t, tt.errorCode, groupErr.Code)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
