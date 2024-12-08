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
	newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
	}

	// Prevent self-referential cycles
	if group.ID == newParentID {
		return E(op, ErrCodeCyclicDependency, "group cannot be its own parent", nil)
	}

	// Check for hierarchical cycles
	if group.IsAncestor(newParentID) {
		return E(op, ErrCodeCyclicDependency,
			fmt.Sprintf("cannot set parent to %s as it is already an ancestor", newParentID), nil)
	}
	if newParent.IsAncestor(group.ID) {
		return E(op, ErrCodeCyclicDependency,
			fmt.Sprintf("cannot set parent to %s as it is a descendant", newParentID), nil)
	}

	return nil
}

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

	// Store a copy of the original group to handle rollbacks if needed
	originalGroup := group.DeepCopy()

	// Validate the hierarchy change
	if err := h.ValidateHierarchyChange(ctx, group, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// If there's an existing parent, get it first so we can update it
	var oldParent *Group
	if group.ParentID != "" {
		var err error
		oldParent, err = h.store.Get(ctx, group.TenantID, group.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}
	}

	// If there's a new parent, get it and add the child
	var newParent *Group
	if newParentID != "" {
		var err error
		newParent, err = h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}

		// Add the child to the new parent
		newParent.AddChild(group.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	}

	// Update the group's parent and ancestry information
	var parentInfo *AncestryInfo
	if newParent != nil {
		parentInfo = &newParent.Ancestry
	}
	if err := group.SetParent(newParentID, parentInfo); err != nil {
		// Rollback new parent update if it failed
		if newParent != nil {
			newParent.RemoveChild(group.ID)
			_ = h.store.Update(ctx, newParent)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// If there was an old parent, remove the child from it
	if oldParent != nil {
		oldParent.RemoveChild(group.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			// Rollback all changes if old parent update failed
			if err := h.store.Update(ctx, originalGroup); err != nil {
				return E(op, ErrCodeStoreOperation, "failed to rollback group changes", err)
			}
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Finally update the group itself
	if err := h.store.Update(ctx, group); err != nil {
		// Attempt to rollback all changes if final update failed
		if oldParent != nil {
			oldParent.AddChild(group.ID)
			_ = h.store.Update(ctx, oldParent)
		}
		if newParent != nil {
			newParent.RemoveChild(group.ID)
			_ = h.store.Update(ctx, newParent)
		}
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	return nil
}

// GetAncestors retrieves all ancestor groups of the given group
func (h *HierarchyManager) GetAncestors(ctx context.Context, group *Group) ([]*Group, error) {
	const op = "HierarchyManager.GetAncestors"

	ancestors := make([]*Group, 0, group.Ancestry.Depth)
	for _, ancestorID := range group.Ancestry.PathParts[:len(group.Ancestry.PathParts)-1] {
		ancestor, err := h.store.Get(ctx, group.TenantID, ancestorID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation, "failed to get ancestor group", err)
		}
		ancestors = append(ancestors, ancestor)
	}
	return ancestors, nil
}

// GetDescendants retrieves all descendant groups of the given group
func (h *HierarchyManager) GetDescendants(ctx context.Context, group *Group) ([]*Group, error) {
	const op = "HierarchyManager.GetDescendants"

	descendants := make([]*Group, 0)
	for _, childID := range group.Ancestry.Children {
		child, err := h.store.Get(ctx, group.TenantID, childID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation, "failed to get child group", err)
		}

		// Add this child to descendants
		descendants = append(descendants, child)

		// Recursively get descendants of this child
		childDescendants, err := h.GetDescendants(ctx, child)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if len(childDescendants) > 0 {
			descendants = append(descendants, childDescendants...)
		}
	}
	return descendants, nil
}

// ValidateHierarchyIntegrity checks the integrity of the entire group hierarchy
func (h *HierarchyManager) ValidateHierarchyIntegrity(ctx context.Context, tenantID string) error {
	const op = "HierarchyManager.ValidateHierarchyIntegrity"

	// Get all groups for the tenant
	groups, err := h.store.List(ctx, ListOptions{TenantID: tenantID})
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build a map of group IDs to groups for efficient lookup
	groupMap := make(map[string]*Group)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	// Validate each group's hierarchy information
	for _, group := range groups {
		// Validate parent relationship
		if group.ParentID != "" {
			parent, exists := groupMap[group.ParentID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", group.ID, group.ParentID),
					nil)
			}
			// Validate that this group is in the parent's children list
			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == group.ID {
					found = true
					break
				}
			}
			if !found {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s not found in parent %s children list", group.ID, parent.ID),
					nil)
			}
		}

		// Validate children relationships
		for _, childID := range group.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", group.ID, childID),
					nil)
			}
			if child.ParentID != group.ID {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference parent %s", child.ID, group.ID),
					nil)
			}
		}

		// Validate ancestry path
		if len(group.Ancestry.PathParts) == 0 {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has empty ancestry path", group.ID),
				nil)
		}
		if group.Ancestry.PathParts[len(group.Ancestry.PathParts)-1] != group.ID {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid ancestry path - last part should be own ID", group.ID),
				nil)
		}

		// Validate depth matches path parts
		expectedDepth := len(group.Ancestry.PathParts) - 1
		if group.Ancestry.Depth != expectedDepth {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has incorrect depth %d (expected %d)", group.ID, group.Ancestry.Depth, expectedDepth),
				nil)
		}
	}

	return nil
}
