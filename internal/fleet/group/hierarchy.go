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

	// Check that the new parent isn't a descendant of the group
	if newParent.IsAncestor(group.ID) {
		return E(op, ErrCodeCyclicDependency, "cyclic dependency detected", nil)
	}

	return nil
}

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

	// Validate the hierarchy change
	if err := h.ValidateHierarchyChange(ctx, group, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// If there's an existing parent, remove this group from its children
	if group.ParentID != "" {
		oldParent, err := h.store.Get(ctx, group.TenantID, group.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}
		oldParent.RemoveChild(group.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Update the group's parent relationship
	if newParentID != "" {
		newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
		if err := group.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		newParent.AddChild(group.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	} else {
		if err := group.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	// Update the group itself
	if err := h.store.Update(ctx, group); err != nil {
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
		descendants = append(descendants, child)

		// Recursively get descendants of this child
		childDescendants, err := h.GetDescendants(ctx, child)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		descendants = append(descendants, childDescendants...)
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
		if len(group.Ancestry.PathParts) == 0 || group.Ancestry.PathParts[len(group.Ancestry.PathParts)-1] != group.ID {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid ancestry path", group.ID),
				nil)
		}
	}

	return nil
}
