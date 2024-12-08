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

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

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

	// Prepare the new parent update if needed
	if newParent != nil {
		newParentCopy := newParent.DeepCopy()
		newParentCopy.AddChild(currentGroup.ID)
		if err := h.store.Update(ctx, newParentCopy); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	}

	// Finally update the group itself
	groupCopy := currentGroup.DeepCopy()
	if newParent != nil {
		if err := groupCopy.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to set parent reference", err)
		}
	} else {
		if err := groupCopy.SetParent("", nil); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to set as root group", err)
		}
	}

	if err := h.store.Update(ctx, groupCopy); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	return nil
}
