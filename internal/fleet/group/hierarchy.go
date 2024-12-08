package group

import (
	"context"
	"fmt"
)

// HierarchyManager provides operations for managing group hierarchies
type HierarchyManager struct {
	store Store
}

// NewHierarchyManager creates a new hierarchy manager
func NewHierarchyManager(store Store) *HierarchyManager {
	return &HierarchyManager{
		store: store,
	}
}

// ValidateHierarchyChange checks if a proposed parent change would create cycles
func (h *HierarchyManager) ValidateHierarchyChange(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.ValidateHierarchyChange"

	if newParentID == "" {
		return nil // Moving to root is always valid
	}

	// Check that the new parent exists and is in the same tenant
	if _, err := h.store.Get(ctx, group.TenantID, newParentID); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
	}

	// Prevent self-referential cycles
	if group.ID == newParentID {
		return E(op, ErrCodeCyclicDependency, "group cannot be its own parent", nil)
	}

	// Check if new parent is a descendant of the current group
	descendants, err := h.GetDescendants(ctx, group)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get descendants", err)
	}

	// Check if the new parent is among the descendants
	for _, desc := range descendants {
		if desc.ID == newParentID {
			return E(op, ErrCodeCyclicDependency,
				fmt.Sprintf("cannot set parent to %s as it would create a cycle", newParentID), nil)
		}
	}

	return nil
}

// ValidateHierarchyIntegrity validates the complete hierarchy structure for a tenant
func (h *HierarchyManager) ValidateHierarchyIntegrity(ctx context.Context, tenantID string) error {
	const op = "HierarchyManager.ValidateHierarchyIntegrity"

	// Get all groups for the tenant
	groups, err := h.store.List(ctx, ListOptions{TenantID: tenantID})
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build a map for efficient lookup
	groupMap := make(map[string]*Group)
	for _, g := range groups {
		groupMap[g.ID] = g
	}

	// Validate each group's relationships
	for _, g := range groups {
		// Validate parent reference
		if g.ParentID != "" {
			parent, exists := groupMap[g.ParentID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

			// Verify parent has this group as a child
			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == g.ID {
					found = true
					break
				}
			}
			if !found {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s not found in parent %s children list", g.ID, g.ParentID), nil)
			}
		}

		// Validate child references
		for _, childID := range g.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}
			if child.ParentID != g.ID {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference group %s as parent", childID, g.ID), nil)
			}
		}

		// Validate ancestry path
		if g.ParentID == "" && g.Ancestry.Path != "/"+g.ID {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("root group %s has invalid path %s", g.ID, g.Ancestry.Path), nil)
		}

		// Validate depth
		expectedDepth := len(g.Ancestry.PathParts) - 1
		if g.Ancestry.Depth != expectedDepth {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid depth %d, expected %d", g.ID, g.Ancestry.Depth, expectedDepth), nil)
		}
	}

	return nil
}

// GetDescendants returns all descendant groups of the given group
func (h *HierarchyManager) GetDescendants(ctx context.Context, group *Group) ([]*Group, error) {
	const op = "HierarchyManager.GetDescendants"

	descendants := make([]*Group, 0)

	// Process each child
	for _, childID := range group.Ancestry.Children {
		// Get the child group
		child, err := h.store.Get(ctx, group.TenantID, childID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation,
				fmt.Sprintf("failed to get child group %s", childID), err)
		}

		// Add child to descendants
		descendants = append(descendants, child)

		// Recursively get descendants of the child
		childDescendants, err := h.GetDescendants(ctx, child)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation,
				fmt.Sprintf("failed to get descendants of child group %s", childID), err)
		}

		descendants = append(descendants, childDescendants...)
	}

	return descendants, nil
}

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

	// Validate hierarchy change won't create cycles
	if err := h.ValidateHierarchyChange(ctx, group, newParentID); err != nil {
		return err
	}

	// Get group from store to ensure we have latest version
	currentGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get current group", err)
	}

	// If current parent exists, remove this group from its children
	if currentGroup.ParentID != "" {
		oldParent, err := h.store.Get(ctx, group.TenantID, currentGroup.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}

		oldParent.RemoveChild(group.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Update group's parent reference and ancestry
	groupCopy := group.DeepCopy()

	if newParentID != "" {
		// Get new parent to update bidirectional relationship
		newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}

		// Update ancestry info
		if err := groupCopy.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to set parent reference", err)
		}

		// Add this group as child of new parent
		newParent.AddChild(group.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	} else {
		// Set as root group
		if err := groupCopy.SetParent("", nil); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to set as root group", err)
		}
	}

	// Update the group
	if err := h.store.Update(ctx, groupCopy); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	return nil
}
