package group

import (
	"context"
	"fmt"
)

// ValidateHierarchyIntegrity checks the integrity of the entire hierarchy
func (h *HierarchyManager) ValidateHierarchyIntegrity(ctx context.Context, tenantID string) error {
	const op = "HierarchyManager.ValidateHierarchyIntegrity"

	// Get all groups for tenant
	groups, err := h.store.List(ctx, tenantID, ListOptions{})
	if err != nil {
		return E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build map for quick lookups
	groupMap := make(map[string]*Group)
	for _, g := range groups {
		groupMap[g.ID] = g
	}

	// Check parent-child relationships
	for _, g := range groups {
		// Validate parent reference
		if g.ParentID != "" {
			parent, exists := groupMap[g.ParentID]
			if !exists {
				return E(op, ErrCodeInvalidHierarchy,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

			// Verify parent has this group as child
			if !contains(parent.Ancestry.Children, g.ID) {
				return E(op, ErrCodeInvalidHierarchy,
					fmt.Sprintf("parent group %s does not list %s as child", parent.ID, g.ID), nil)
			}
		}

		// Validate children references
		for _, childID := range g.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return E(op, ErrCodeInvalidHierarchy,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}

			// Verify child's parent reference
			if child.ParentID != g.ID {
				return E(op, ErrCodeInvalidHierarchy,
					fmt.Sprintf("child group %s does not reference %s as parent", child.ID, g.ID), nil)
			}
		}

		// Validate ancestry path
		if err := h.validateAncestryPath(g, groupMap); err != nil {
			return E(op, ErrCodeInvalidHierarchy, "invalid ancestry path", err)
		}
	}

	return nil
}

// validateAncestryPath checks if a group's ancestry path is valid
func (h *HierarchyManager) validateAncestryPath(group *Group, groupMap map[string]*Group) error {
	const op = "HierarchyManager.validateAncestryPath"

	// Validate basic path structure
	if group.Ancestry.Path == "" || group.Ancestry.Path[0] != '/' {
		return E(op, ErrCodeInvalidHierarchy,
			fmt.Sprintf("group %s has invalid path format: %s", group.ID, group.Ancestry.Path), nil)
	}

	// Validate path parts
	if len(group.Ancestry.PathParts) == 0 || group.Ancestry.PathParts[len(group.Ancestry.PathParts)-1] != group.ID {
		return E(op, ErrCodeInvalidHierarchy,
			fmt.Sprintf("group %s has invalid path parts", group.ID), nil)
	}

	// Verify path matches ancestry chain
	current := group
	for i := len(group.Ancestry.PathParts) - 2; i >= 0; i-- {
		parentID := group.Ancestry.PathParts[i]
		parent, exists := groupMap[parentID]
		if !exists {
			return E(op, ErrCodeInvalidHierarchy,
				fmt.Sprintf("group %s has invalid ancestor %s in path", group.ID, parentID), nil)
		}

		if current.ParentID != parent.ID {
			return E(op, ErrCodeInvalidHierarchy,
				fmt.Sprintf("group %s has mismatched ancestry path", group.ID), nil)
		}
		current = parent
	}

	// Verify depth matches path
	expectedDepth := len(group.Ancestry.PathParts) - 1
	if group.Ancestry.Depth != expectedDepth {
		return E(op, ErrCodeInvalidHierarchy,
			fmt.Sprintf("group %s has incorrect depth: got %d, expected %d",
				group.ID, group.Ancestry.Depth, expectedDepth), nil)
	}

	return nil
}
