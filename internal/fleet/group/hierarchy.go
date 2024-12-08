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

	// Get a fresh copy of the group to ensure we have the latest state
	currentGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get current group state", err)
	}

	// Validate the hierarchy change
	if err := h.ValidateHierarchyChange(ctx, currentGroup, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// If there's an existing parent, remove this group from its children
	if currentGroup.ParentID != "" {
		oldParent, err := h.store.Get(ctx, currentGroup.TenantID, currentGroup.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}
		oldParent.RemoveChild(currentGroup.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Update the group's parent relationship
	if newParentID != "" {
		newParent, err := h.store.Get(ctx, currentGroup.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
		if err := currentGroup.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		newParent.AddChild(currentGroup.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	} else {
		if err := currentGroup.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	// Update the group with its new state
	if err := h.store.Update(ctx, currentGroup); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	// Update the input group to reflect the changes
	*group = *currentGroup

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

	var descendants []*Group

	// First, get all groups for the tenant for efficient processing
	allGroups, err := h.store.List(ctx, ListOptions{TenantID: group.TenantID})
	if err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build a map for efficient lookup
	groupMap := make(map[string]*Group)
	for _, g := range allGroups {
		groupMap[g.ID] = g
	}

	// Helper function to recursively collect descendants
	var collectDescendants func(parentID string) error
	collectDescendants = func(parentID string) error {
		parent, exists := groupMap[parentID]
		if !exists {
			return nil // Skip if group doesn't exist
		}

		for _, childID := range parent.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				continue // Skip invalid children
			}

			// Ensure we haven't already added this descendant
			alreadyAdded := false
			for _, d := range descendants {
				if d.ID == child.ID {
					alreadyAdded = true
					break
				}
			}

			if !alreadyAdded {
				descendants = append(descendants, child)
				if err := collectDescendants(childID); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Start collection from the root group
	if err := collectDescendants(group.ID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
		// Validate parent relationship if it exists
		if group.ParentID != "" {
			parent, exists := groupMap[group.ParentID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", group.ID, group.ParentID),
					nil)
			}

			// Verify the parent-child relationship is bi-directional
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

		// Validate ancestry path structure
		if len(group.Ancestry.PathParts) == 0 || group.Ancestry.PathParts[len(group.Ancestry.PathParts)-1] != group.ID {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid ancestry path", group.ID),
				nil)
		}

		// Verify each ancestor in the path exists and maintains proper relationships
		for i := 0; i < len(group.Ancestry.PathParts)-1; i++ {
			ancestorID := group.Ancestry.PathParts[i]
			ancestor, exists := groupMap[ancestorID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent ancestor %s in path", group.ID, ancestorID),
					nil)
			}

			// If this is the immediate parent, verify the parent-child relationship
			if i == len(group.Ancestry.PathParts)-2 {
				found := false
				for _, childID := range ancestor.Ancestry.Children {
					if childID == group.ID {
						found = true
						break
					}
				}
				if !found {
					return E(op, ErrCodeInvalidGroup,
						fmt.Sprintf("group %s not found in ancestor %s children list", group.ID, ancestor.ID),
						nil)
				}
			}
		}
	}

	return nil
}
