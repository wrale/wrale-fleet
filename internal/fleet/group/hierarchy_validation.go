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
	if _, err := h.store.Get(ctx, group.TenantID, newParentID); err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get new parent group", err)
	}

	// Prevent self-referential cycles
	if group.ID == newParentID {
		return E(op, ErrCodeCyclicDependency, "group cannot be its own parent", nil)
	}

	// Check if new parent is a descendant of the current group
	descendants, err := h.GetDescendants(ctx, group)
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to get descendants", err)
	}

	// Check if the new parent is among the descendants
	for _, desc := range descendants {
		if desc.ID == newParentID {
			return E(op, ErrCodeCyclicDependency,
				fmt.Sprintf("cannot set parent to %s as it would create a cycle", newParentID), nil)
		}
	}

	return nil
}
