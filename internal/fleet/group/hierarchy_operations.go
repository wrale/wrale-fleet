package group

import (
	"context"
	"fmt"
)

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

	// Check that the new parent isn't a descendant of the group by checking
	// if the group ID appears in its ancestry path
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

	// Store the original parent ID for potential rollback
	originalParentID := currentGroup.ParentID

	// Pre-validate the hierarchy change
	if err := h.ValidateHierarchyChange(ctx, currentGroup, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Prepare new parent information if needed
	var newParentInfo *AncestryInfo
	if newParentID != "" {
		newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
		newParentInfo = &newParent.Ancestry
	}

	// Save current children before any modifications
	currentChildren := make([]string, len(currentGroup.Ancestry.Children))
	copy(currentChildren, currentGroup.Ancestry.Children)

	// Update the current group's relationships
	if err := currentGroup.SetParent(newParentID, newParentInfo); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Restore children that were cleared by SetParent
	currentGroup.Ancestry.Children = currentChildren

	// Update old parent if it exists
	if originalParentID != "" {
		oldParent, err := h.store.Get(ctx, currentGroup.TenantID, originalParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}

		oldParent.RemoveChild(currentGroup.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}
	}

	// Update new parent if it exists
	if newParentID != "" {
		newParent, err := h.store.Get(ctx, currentGroup.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}

		newParent.AddChild(currentGroup.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	}

	// Update the current group with all changes
	if err := h.store.Update(ctx, currentGroup); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group state", err)
	}

	// Verify the hierarchy integrity after all updates
	if err := h.ValidateHierarchyIntegrity(ctx, currentGroup.TenantID); err != nil {
		// Attempt rollback to original state
		if rbErr := h.rollbackHierarchyChange(ctx, currentGroup, originalParentID); rbErr != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("hierarchy validation failed and rollback failed: %v (rollback error: %v)",
					err, rbErr), err)
		}
		return E(op, ErrCodeInvalidGroup, "hierarchy integrity validation failed", err)
	}

	// Update the input group to reflect the changes
	*group = *currentGroup

	return nil
}

// rollbackHierarchyChange attempts to restore a group's previous hierarchy state
func (h *HierarchyManager) rollbackHierarchyChange(ctx context.Context, group *Group, originalParentID string) error {
	const op = "rollbackHierarchyChange"

	// Get fresh copies of all affected groups
	rollbackGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to get group for rollback: %w", op, err)
	}

	// If rolling back to original parent
	if originalParentID != "" {
		originalParent, err := h.store.Get(ctx, group.TenantID, originalParentID)
		if err != nil {
			return fmt.Errorf("%s: failed to get original parent for rollback: %w", op, err)
		}

		if err := rollbackGroup.SetParent(originalParentID, &originalParent.Ancestry); err != nil {
			return fmt.Errorf("%s: failed to set original parent: %w", op, err)
		}

		originalParent.AddChild(rollbackGroup.ID)
		if err := h.store.Update(ctx, originalParent); err != nil {
			return fmt.Errorf("%s: failed to update original parent: %w", op, err)
		}
	} else {
		// Rolling back to root
		if err := rollbackGroup.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: failed to set as root: %w", op, err)
		}
	}

	// Save the rolled back state
	if err := h.store.Update(ctx, rollbackGroup); err != nil {
		return fmt.Errorf("%s: failed to update rolled back group: %w", op, err)
	}

	*group = *rollbackGroup
	return nil
}
