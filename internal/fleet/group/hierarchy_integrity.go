package group

import (
	"context"
	"fmt"
)

// ValidateHierarchyIntegrity validates the complete hierarchy structure for a tenant
func (h *HierarchyManager) ValidateHierarchyIntegrity(ctx context.Context, tenantID string) error {
	const op = "HierarchyManager.ValidateHierarchyIntegrity"

	// Get all groups for the tenant
	groups, err := h.store.List(ctx, ListOptions{TenantID: tenantID})
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build a map for efficient lookup
	groupMap := make(map[string]*Group)
	for _, g := range groups {
		groupMap[g.ID] = g
	}

	// Validate each group's relationships
	for _, g := range groups {
		// Validate parent reference
		if g.ParentID != "" {
			parent, exists := groupMap[g.ParentID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

			// Verify parent has this group as a child
			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == g.ID {
					found = true
					break
				}
			}
			if !found {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s not found in parent %s children list", g.ID, g.ParentID), nil)
			}
		}

		// Validate ancestry path and depth
		if g.ParentID == "" && g.Ancestry.Path != "/"+g.ID {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("root group %s has invalid path %s", g.ID, g.Ancestry.Path), nil)
		}

		expectedDepth := len(g.Ancestry.PathParts) - 1
		if g.Ancestry.Depth != expectedDepth {
			return E(op, ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid depth %d, expected %d", g.ID, g.Ancestry.Depth, expectedDepth), nil)
		}

		// Validate child references
		for _, childID := range g.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}
			if child.ParentID != g.ID {
				return E(op, ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference parent %s", child.ID, g.ID), nil)
			}
		}
	}

	return nil
}
