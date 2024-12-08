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

	// Validate the hierarchy change
	if err := h.ValidateHierarchyChange(ctx, currentGroup, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Clear existing parent relationship if any
	if currentGroup.ParentID != "" {
		oldParent, err := h.store.Get(ctx, currentGroup.TenantID, currentGroup.ParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}

		// Update the old parent's children list
		oldParent.RemoveChild(currentGroup.ID)
		if err := h.store.Update(ctx, oldParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update old parent group", err)
		}

		// Clear the current group's parent reference
		if err := currentGroup.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	// Establish new parent relationship if specified
	if newParentID != "" {
		// Get the new parent group
		newParent, err := h.store.Get(ctx, currentGroup.TenantID, newParentID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}

		// Update the parent-child relationship from both sides
		if err := currentGroup.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		// Update the new parent's children list
		newParent.AddChild(currentGroup.ID)
		if err := h.store.Update(ctx, newParent); err != nil {
			return E(op, ErrCodeStoreOperation, "failed to update new parent group", err)
		}
	}

	// Persist the updated group state
	if err := h.store.Update(ctx, currentGroup); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to update group", err)
	}

	// Update the input group to reflect the changes
	*group = *currentGroup

	// Verify the hierarchy integrity after the update
	if err := h.ValidateHierarchyIntegrity(ctx, group.TenantID); err != nil {
		return E(op, ErrCodeInvalidGroup, "hierarchy integrity validation failed after update", err)
	}

	return nil
}
