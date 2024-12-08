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

	// Get the current state of the group to handle rollbacks
	originalGroup := group.DeepCopy()

	// First handle the new parent relationship if it exists
	var newParentInfo *AncestryInfo
	if newParentID != "" {
		newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
		newParentInfo = &newParent.Ancestry

		// Update and save the new parent first
		newParent.AddChild(group.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	}

	// Then update the group's own parent reference and ancestry
	if err := group.SetParent(newParentID, newParentInfo); err != nil {
		// Rollback new parent change if we failed to update the group
		if newParentID != "" {
			newParent, _ := h.store.Get(ctx, group.TenantID, newParentID)
			if newParent != nil {
				newParent.RemoveChild(group.ID)
				_ = h.store.Update(ctx, newParent)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Now update the group
	if err := h.store.Update(ctx, group); err != nil {
		// Try to restore original state if update fails
		if newParentID != "" {
			newParent, _ := h.store.Get(ctx, group.TenantID, newParentID)
			if newParent != nil {
				newParent.RemoveChild(group.ID)
				_ = h.store.Update(ctx, newParent)
			}
		}
		_ = h.store.Update(ctx, originalGroup)
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	// Finally handle the old parent relationship if it exists
	// This ensures we don't have a period where the child points to a new parent
	// but is still in the old parent's children list
	if originalGroup.ParentID != "" && originalGroup.ParentID != newParentID {
		oldParent, err := h.store.Get(ctx, group.TenantID, originalGroup.ParentID)
		if err == nil { // Proceed even if old parent not found
			oldParent.RemoveChild(group.ID)
			if err := h.store.Update(ctx, oldParent); err != nil {
				// If we can't update old parent, revert all changes
				if newParentID != "" {
					newParent, _ := h.store.Get(ctx, group.TenantID, newParentID)
					if newParent != nil {
						newParent.RemoveChild(group.ID)
						_ = h.store.Update(ctx, newParent)
					}
				}
				_ = h.store.Update(ctx, originalGroup)
				return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
			}
		}
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

	// Map to track processed groups and prevent cycles
	processed := make(map[string]bool)
	descendants := make([]*Group, 0)

	// Helper function for recursive descent
	var collect func(*Group) error
	collect = func(current *Group) error {
		for _, childID := range current.Ancestry.Children {
			// Skip already processed groups
			if processed[childID] {
				continue
			}

			// Mark as processed to prevent cycles
			processed[childID] = true

			child, err := h.store.Get(ctx, current.TenantID, childID)
			if err != nil {
				return E(op, ErrCodeStoreOperation,
					fmt.Sprintf("failed to get child group %s", childID), err)
			}

			descendants = append(descendants, child.DeepCopy())

			// Recursively process child's descendants
			if err := collect(child); err != nil {
				return err
			}
		}
		return nil
	}

	// Start collection from the root group
	if err := collect(group); err != nil {
		return nil, err
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

		// Validate children relationships are bidirectional
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

		// Validate ancestry path and depth
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

		// Verify depth matches path parts
		expectedDepth := len(group.Ancestry.PathParts) - 1
		if group.Ancestry.Depth != expectedDepth {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has incorrect depth %d (expected %d)", group.ID, group.Ancestry.Depth, expectedDepth),
				nil)
		}
	}

	return nil
}
