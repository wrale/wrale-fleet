package group

import (
	"context"
	"fmt"
)

// GetAncestors retrieves all ancestor groups of the given group
func (h *HierarchyManager) GetAncestors(ctx context.Context, group *Group) ([]*Group, error) {
	const op = "HierarchyManager.GetAncestors"

	ancestors := make([]*Group, 0, group.Ancestry.Depth)
	for _, ancestorID := range group.Ancestry.PathParts[:len(group.Ancestry.PathParts)-1] {
		ancestor, err := h.store.Get(ctx, group.TenantID, ancestorID)
		if err != nil {
			return nil, E(op, ErrCodeStoreOperation, "failed to get ancestor group", err)
		}
		ancestors = append(ancestors, ancestor)
	}
	return ancestors, nil
}

// GetDescendants retrieves all descendant groups of the given group
func (h *HierarchyManager) GetDescendants(ctx context.Context, group *Group) ([]*Group, error) {
	const op = "HierarchyManager.GetDescendants"

	// First get all groups for the tenant for efficient processing
	allGroups, err := h.store.List(ctx, ListOptions{TenantID: group.TenantID})
	if err != nil {
		return nil, E(op, ErrCodeStoreOperation, "failed to list groups", err)
	}

	// Build maps for efficient lookup and tracking
	groupMap := make(map[string]*Group)
	for _, g := range allGroups {
		groupMap[g.ID] = g
	}

	var descendants []*Group
	visited := make(map[string]bool)

	// Helper function to recursively collect descendants with cycle detection
	var collect func(groupID string) error
	collect = func(groupID string) error {
		currentGroup, exists := groupMap[groupID]
		if !exists || visited[groupID] {
			return nil
		}
		visited[groupID] = true

		// Process direct children first
		for _, childID := range currentGroup.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				continue
			}

			// Verify bi-directional relationship
			if child.ParentID == currentGroup.ID {
				descendants = append(descendants, child)
				if err := collect(childID); err != nil {
					return err
				}
			}
		}

		return nil
	}

	// Start collection from the root group
	if err := collect(group.ID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return descendants, nil
}
