package group

import (
	"context"
	"fmt"
)

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

// buildAncestry constructs the ancestry information for a group
func (h *HierarchyManager) buildAncestry(ctx context.Context, group *Group, parent *Group) (*AncestryInfo, error) {
	ancestry := &AncestryInfo{
		Children: make([]string, 0),
	}

	if parent == nil {
		// Root node
		ancestry.Path = "/" + group.ID
		ancestry.PathParts = []string{group.ID}
		ancestry.Depth = 0
	} else {
		// Build path components by combining parent's path parts with current group
		ancestry.PathParts = make([]string, len(parent.Ancestry.PathParts)+1)
		copy(ancestry.PathParts, parent.Ancestry.PathParts)
		ancestry.PathParts[len(parent.Ancestry.PathParts)] = group.ID

		// Build the full path string
		ancestry.Path = parent.Ancestry.Path + "/" + group.ID
		ancestry.Depth = parent.Ancestry.Depth + 1

		// Validate the ancestry chain
		if ancestry.Depth+1 != len(ancestry.PathParts) {
			return nil, E("buildAncestry", ErrCodeInvalidHierarchy,
				"ancestry depth doesn't match path parts length", nil)
		}
	}

	return ancestry, nil
}

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

	// Lock the entire hierarchy operation
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get current state of the group and validate the change
	currentGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get current group", err)
	}

	if err := h.ValidateHierarchyChange(ctx, currentGroup, newParentID); err != nil {
		return err
	}

	// Get new parent (if any) and prepare its update
	var newParent *Group
	if newParentID != "" {
		newParent, err = h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
	}

	// Remove from old parent first if it exists
	if currentGroup.ParentID != "" {
		oldParent, err := h.store.Get(ctx, group.TenantID, currentGroup.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}

		oldParentCopy := oldParent.DeepCopy()
		oldParentCopy.RemoveChild(currentGroup.ID)
		if err := h.store.Update(ctx, oldParentCopy); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Update the group's ancestry
	newAncestry, err := h.buildAncestry(ctx, currentGroup, newParent)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to build ancestry", err)
	}

	// Prepare the new parent update if needed
	if newParent != nil {
		newParentCopy := newParent.DeepCopy()
		if !contains(newParentCopy.Ancestry.Children, currentGroup.ID) {
			newParentCopy.AddChild(currentGroup.ID)
			if err := h.store.Update(ctx, newParentCopy); err != nil {
				return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
			}
		}
	}

	// Update the group with new ancestry
	groupCopy := currentGroup.DeepCopy()
	groupCopy.ParentID = newParentID
	groupCopy.Ancestry = *newAncestry

	if err := h.store.Update(ctx, groupCopy); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	// Update all descendants with new paths
	descendants, err := h.GetDescendants(ctx, groupCopy)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get descendants for path update", err)
	}

	for _, desc := range descendants {
		parent, err := h.store.Get(ctx, desc.TenantID, desc.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, fmt.Sprintf("failed to get parent for descendant %s", desc.ID), err)
		}

		descCopy := desc.DeepCopy()
		descAncestry, err := h.buildAncestry(ctx, descCopy, parent)
		if err != nil {
			return E(op, ErrCodeStoreOperation, fmt.Sprintf("failed to build ancestry for descendant %s", desc.ID), err)
		}

		descCopy.Ancestry = *descAncestry
		if err := h.store.Update(ctx, descCopy); err != nil {
			return E(op, ErrCodeStoreOperation, fmt.Sprintf("failed to update descendant %s", desc.ID), err)
		}
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
