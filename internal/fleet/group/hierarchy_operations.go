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

// stagedUpdate represents a pending hierarchy update
type stagedUpdate struct {
	group      *Group
	needUpdate bool
}

// stageUpdates prepares all necessary changes for a hierarchy update
func (h *HierarchyManager) stageUpdates(ctx context.Context, group *Group, newParentID string) (map[string]*stagedUpdate, error) {
	const op = "stageUpdates"
	staged := make(map[string]*stagedUpdate)

	// Stage the current group first
	currentGroup := *group // Make a copy
	staged[group.ID] = &stagedUpdate{
		group:      &currentGroup,
		needUpdate: true,
	}

	// First phase: Load all affected groups
	// If there's a current parent, load it
	var oldParent *Group
	if group.ParentID != "" {
		var err error
		oldParent, err = h.store.Get(ctx, group.TenantID, group.ParentID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
		}
		staged[oldParent.ID] = &stagedUpdate{
			group:      oldParent,
			needUpdate: true,
		}
	}

	// If there's a new parent, load it
	var newParent *Group
	if newParentID != "" {
		var err error
		newParent, err = h.store.Get(ctx, group.TenantID, newParentID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
		}
		staged[newParentID] = &stagedUpdate{
			group:      newParent,
			needUpdate: true,
		}
	}

	// Second phase: Update parent references
	// Remove from old parent's children list if it exists
	if oldParent != nil {
		staged[oldParent.ID].group.RemoveChild(group.ID)
	}

	// Update current group's parent reference and ancestry
	if newParentID != "" {
		if err := staged[group.ID].group.SetParent(newParentID, &newParent.Ancestry); err != nil {
			return nil, E(op, ErrCodeInvalidOperation, "failed to update group ancestry", err)
		}
		// Add to new parent's children list
		staged[newParentID].group.AddChild(group.ID)
	} else {
		if err := staged[group.ID].group.SetParent("", nil); err != nil {
			return nil, E(op, ErrCodeInvalidOperation, "failed to update group ancestry", err)
		}
	}

	return staged, nil
}

// validateStagedUpdates verifies the consistency of staged changes
func (h *HierarchyManager) validateStagedUpdates(staged map[string]*stagedUpdate) error {
	const op = "validateStagedUpdates"

	// Check each staged group for bidirectional relationship consistency
	for _, update := range staged {
		group := update.group

		// If the group has a parent reference, verify the parent's children list
		if group.ParentID != "" {
			parentUpdate, exists := staged[group.ParentID]
			if exists {
				// Parent is included in staged changes
				found := false
				for _, childID := range parentUpdate.group.Ancestry.Children {
					if childID == group.ID {
						found = true
						break
					}
				}
				if !found {
					return E(op, ErrCodeInvalidGroup,
						fmt.Sprintf("parent %s does not list %s as child",
							group.ParentID, group.ID), nil)
				}
			}
		}

		// For each child in the group's children list
		for _, childID := range group.Ancestry.Children {
			childUpdate, exists := staged[childID]
			if exists {
				// Child is included in staged changes
				if childUpdate.group.ParentID != group.ID {
					return E(op, ErrCodeInvalidGroup,
						fmt.Sprintf("group %s lists %s as child, but child references parent %s",
							group.ID, childID, childUpdate.group.ParentID), nil)
				}
			}
		}
	}

	return nil
}

// applyUpdates persists staged changes to the store
func (h *HierarchyManager) applyUpdates(ctx context.Context, staged map[string]*stagedUpdate) error {
	const op = "applyUpdates"

	// Apply updates in a specific order to maintain consistency:
	// 1. Update old parent (remove child reference)
	// 2. Update group (update ancestry info)
	// 3. Update new parent (add child reference)
	for _, update := range staged {
		if !update.needUpdate {
			continue
		}

		if err := h.store.Update(ctx, update.group); err != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("failed to update group %s", update.group.ID), err)
		}
	}

	return nil
}

// UpdateHierarchy updates a group's position in the hierarchy
func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.UpdateHierarchy"

	// Validate the proposed change
	if err := h.ValidateHierarchyChange(ctx, group, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Store original parent ID for potential rollback
	originalParentID := group.ParentID

	// Stage all updates
	staged, err := h.stageUpdates(ctx, group, newParentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Validate staged changes
	if err := h.validateStagedUpdates(staged); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Apply the changes
	if err := h.applyUpdates(ctx, staged); err != nil {
		// Attempt rollback
		if rbErr := h.rollbackHierarchyChange(ctx, group, originalParentID); rbErr != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("update failed and rollback failed: %v (rollback error: %v)",
					err, rbErr), err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	// Verify final hierarchy integrity
	if err := h.ValidateHierarchyIntegrity(ctx, group.TenantID); err != nil {
		// Attempt rollback
		if rbErr := h.rollbackHierarchyChange(ctx, group, originalParentID); rbErr != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("integrity validation failed and rollback failed: %v (rollback error: %v)",
					err, rbErr), err)
		}
		return fmt.Errorf("%s: hierarchy integrity validation failed: %w", op, err)
	}

	// Update the input group to reflect changes
	*group = *staged[group.ID].group

	return nil
}

// rollbackHierarchyChange attempts to restore a group's previous hierarchy state
func (h *HierarchyManager) rollbackHierarchyChange(ctx context.Context, group *Group, originalParentID string) error {
	const op = "rollbackHierarchyChange"

	// Stage rollback updates
	staged := make(map[string]*stagedUpdate)

	// Get fresh copies of affected groups
	currentGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to get group for rollback: %w", op, err)
	}
	staged[currentGroup.ID] = &stagedUpdate{
		group:      currentGroup,
		needUpdate: true,
	}

	// Handle current parent if different from original
	if currentGroup.ParentID != "" && currentGroup.ParentID != originalParentID {
		currentParent, err := h.store.Get(ctx, group.TenantID, currentGroup.ParentID)
		if err != nil {
			return fmt.Errorf("%s: failed to get current parent for rollback: %w", op, err)
		}
		staged[currentParent.ID] = &stagedUpdate{
			group:      currentParent,
			needUpdate: true,
		}
		staged[currentParent.ID].group.RemoveChild(currentGroup.ID)
	}

	// Handle original parent if it exists
	if originalParentID != "" {
		originalParent, err := h.store.Get(ctx, group.TenantID, originalParentID)
		if err != nil {
			return fmt.Errorf("%s: failed to get original parent for rollback: %w", op, err)
		}
		staged[originalParent.ID] = &stagedUpdate{
			group:      originalParent,
			needUpdate: true,
		}

		// First update the current group's ancestry
		if err := staged[currentGroup.ID].group.SetParent(originalParentID, &originalParent.Ancestry); err != nil {
			return fmt.Errorf("%s: failed to restore original parent: %w", op, err)
		}
		// Then update the original parent's children list
		staged[originalParent.ID].group.AddChild(currentGroup.ID)
	} else {
		// Restore to root
		if err := staged[currentGroup.ID].group.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: failed to restore as root: %w", op, err)
		}
	}

	// Apply rollback changes in correct order
	if err := h.applyUpdates(ctx, staged); err != nil {
		return fmt.Errorf("%s: failed to apply rollback updates: %w", op, err)
	}

	return nil
}
