package group

import (
	"context"
	"fmt"
)

type stagedUpdate struct {
	group      *Group
	needUpdate bool
}

func (h *HierarchyManager) ValidateHierarchyChange(ctx context.Context, group *Group, newParentID string) error {
	const op = "HierarchyManager.ValidateHierarchyChange"
	if newParentID == "" {
		return nil
	}

	newParent, err := h.store.Get(ctx, group.TenantID, newParentID)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
	}

	if group.ID == newParentID {
		return E(op, ErrCodeCyclicDependency, "group cannot be its own parent", nil)
	}

	if newParent.IsAncestor(group.ID) {
		return E(op, ErrCodeCyclicDependency, "cyclic dependency detected", nil)
	}

	return nil
}

func (h *HierarchyManager) loadConnectedGroups(ctx context.Context, group *Group, staged map[string]*stagedUpdate) error {
	const op = "loadConnectedGroups"

	// Load children of the group if not already staged
	for _, childID := range group.Ancestry.Children {
		if _, exists := staged[childID]; !exists {
			child, err := h.store.Get(ctx, group.TenantID, childID)
			if err != nil {
				return E(op, ErrCodeStoreOperation, fmt.Sprintf("failed to load child group %s", childID), err)
			}
			staged[childID] = &stagedUpdate{group: child}
		}
	}

	// Load group's ancestors if not already staged
	if group.ParentID != "" {
		ancestors, err := h.store.GetAncestors(ctx, group.TenantID, group.ID)
		if err != nil {
			return E(op, ErrCodeStoreOperation, "failed to load ancestor groups", err)
		}
		for _, ancestor := range ancestors {
			if _, exists := staged[ancestor.ID]; !exists {
				staged[ancestor.ID] = &stagedUpdate{group: ancestor}
			}
		}
	}

	return nil
}

func (h *HierarchyManager) stageUpdates(ctx context.Context, group *Group, newParentID string) (map[string]*stagedUpdate, error) {
	const op = "stageUpdates"
	staged := make(map[string]*stagedUpdate)

	// Stage the target group
	currentGroup := *group
	staged[group.ID] = &stagedUpdate{
		group:      &currentGroup,
		needUpdate: true,
	}

	// Load all connected groups first
	if err := h.loadConnectedGroups(ctx, &currentGroup, staged); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Handle old parent relationship
	if group.ParentID != "" {
		oldParent, exists := staged[group.ParentID]
		if !exists {
			var err error
			oldParent := &stagedUpdate{needUpdate: true}
			oldParent.group, err = h.store.Get(ctx, group.TenantID, group.ParentID)
			if err != nil {
				return nil, E(op, ErrCodeStoreOperation, "failed to get old parent group", err)
			}
			staged[group.ParentID] = oldParent
		} else {
			oldParent.needUpdate = true
		}
		oldParent.group.RemoveChild(group.ID)
	}

	// Handle new parent relationship
	if newParentID != "" {
		newParent, exists := staged[newParentID]
		if !exists {
			var err error
			newParent := &stagedUpdate{needUpdate: true}
			newParent.group, err = h.store.Get(ctx, group.TenantID, newParentID)
			if err != nil {
				return nil, E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
			}
			staged[newParentID] = newParent
			if err := h.loadConnectedGroups(ctx, newParent.group, staged); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		} else {
			newParent.needUpdate = true
		}

		// Update current group's ancestry
		if err := staged[group.ID].group.SetParent(newParentID, &newParent.group.Ancestry); err != nil {
			return nil, E(op, ErrCodeInvalidOperation, "failed to update group ancestry", err)
		}

		// Update new parent's children
		newParent.group.AddChild(group.ID)
	} else {
		if err := staged[group.ID].group.SetParent("", nil); err != nil {
			return nil, E(op, ErrCodeInvalidOperation, "failed to update group ancestry", err)
		}
	}

	return staged, nil
}

func (h *HierarchyManager) validateStagedUpdates(staged map[string]*stagedUpdate) error {
	const op = "validateStagedUpdates"

	for _, update := range staged {
		group := update.group
		if group.ParentID != "" {
			// Verify parent exists and lists this group as a child
			parentUpdate, exists := staged[group.ParentID]
			if exists {
				found := false
				for _, childID := range parentUpdate.group.Ancestry.Children {
					if childID == group.ID {
						found = true
						break
					}
				}
				if !found {
					return E(op, ErrCodeInvalidGroup,
						fmt.Sprintf("group %s's parent %s does not list it as a child",
							group.ID, group.ParentID), nil)
				}
			}
		}

		// Verify all children exist and reference this group as parent
		for _, childID := range group.Ancestry.Children {
			childUpdate, exists := staged[childID]
			if exists && childUpdate.group.ParentID != group.ID {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s lists %s as child, but child has different parent %s",
						group.ID, childID, childUpdate.group.ParentID), nil)
			}
		}
	}

	return nil
}

func (h *HierarchyManager) applyUpdates(ctx context.Context, staged map[string]*stagedUpdate) error {
	const op = "applyUpdates"

	// First update groups being removed from hierarchy
	// Then update the target group
	// Finally update groups being added to hierarchy
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

func (h *HierarchyManager) UpdateHierarchy(ctx context.Context, group *Group, newParentID string) error {
	const op = "UpdateHierarchy"

	if err := h.ValidateHierarchyChange(ctx, group, newParentID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	originalParentID := group.ParentID
	staged, err := h.stageUpdates(ctx, group, newParentID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.validateStagedUpdates(staged); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.applyUpdates(ctx, staged); err != nil {
		if rbErr := h.rollbackHierarchyChange(ctx, group, originalParentID); rbErr != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("update failed and rollback failed: %v (rollback error: %v)",
					err, rbErr), err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.ValidateHierarchyIntegrity(ctx, group.TenantID); err != nil {
		if rbErr := h.rollbackHierarchyChange(ctx, group, originalParentID); rbErr != nil {
			return E(op, ErrCodeStoreOperation,
				fmt.Sprintf("integrity check failed and rollback failed: %v (rollback error: %v)",
					err, rbErr), err)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	*group = *staged[group.ID].group
	return nil
}

func (h *HierarchyManager) rollbackHierarchyChange(ctx context.Context, group *Group, originalParentID string) error {
	const op = "rollbackHierarchyChange"

	staged := make(map[string]*stagedUpdate)

	rollbackGroup, err := h.store.Get(ctx, group.TenantID, group.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to get group for rollback: %w", op, err)
	}
	staged[rollbackGroup.ID] = &stagedUpdate{group: rollbackGroup, needUpdate: true}

	if rollbackGroup.ParentID != "" && rollbackGroup.ParentID != originalParentID {
		currentParent, err := h.store.Get(ctx, group.TenantID, rollbackGroup.ParentID)
		if err != nil {
			return fmt.Errorf("%s: failed to get current parent for rollback: %w", op, err)
		}
		staged[currentParent.ID] = &stagedUpdate{group: currentParent, needUpdate: true}
		staged[currentParent.ID].group.RemoveChild(rollbackGroup.ID)
	}

	if originalParentID != "" {
		originalParent, err := h.store.Get(ctx, group.TenantID, originalParentID)
		if err != nil {
			return fmt.Errorf("%s: failed to get original parent for rollback: %w", op, err)
		}
		staged[originalParent.ID] = &stagedUpdate{group: originalParent, needUpdate: true}

		if err := staged[rollbackGroup.ID].group.SetParent(originalParentID, &originalParent.Ancestry); err != nil {
			return fmt.Errorf("%s: failed to restore original parent: %w", op, err)
		}
		staged[originalParent.ID].group.AddChild(rollbackGroup.ID)
	} else {
		if err := staged[rollbackGroup.ID].group.SetParent("", nil); err != nil {
			return fmt.Errorf("%s: failed to restore as root: %w", op, err)
		}
	}

	if err := h.applyUpdates(ctx, staged); err != nil {
		return fmt.Errorf("%s: failed to apply rollback updates: %w", op, err)
	}

	return nil
}
