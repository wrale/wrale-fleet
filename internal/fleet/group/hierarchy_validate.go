package group

import (
	"context"
	"fmt"
)

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

			// Verify bi-directional parent-child relationship
			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == group.ID {
					found = true
					break
				}
			}
			if !found {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s's parent %s does not list it as a child", group.ID, parent.ID),
					nil)
			}
		}

		// Validate child relationships
		for _, childID := range group.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s lists non-existent child %s", group.ID, childID),
					nil)
			}

			// Verify bi-directional relationship from child's perspective
			if child.ParentID != group.ID {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s lists %s as child, but child has different parent %s",
						group.ID, childID, child.ParentID),
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
			if _, exists := groupMap[ancestorID]; !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent ancestor %s in path", group.ID, ancestorID),
					nil)
			}
		}

		// Verify ancestry depth matches path length
		expectedDepth := len(group.Ancestry.PathParts) - 1
		if group.Ancestry.Depth != expectedDepth {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has incorrect depth %d (expected %d)", group.ID, group.Ancestry.Depth, expectedDepth),
				nil)
		}
	}

	return nil
}
