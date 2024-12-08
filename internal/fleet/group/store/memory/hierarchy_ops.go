package memory

import (
	"context"
	"fmt"

	"github.com/wrale/fleet/internal/fleet/group"
)

// GetAncestors retrieves all ancestor groups of the specified group
func (s *Store) GetAncestors(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	ancestors := make([]*group.Group, 0, g.Ancestry.Depth)
	for _, ancestorID := range g.Ancestry.PathParts[:len(g.Ancestry.PathParts)-1] {
		ancestorKey := s.key(tenantID, ancestorID)
		ancestor, exists := s.groups[ancestorKey]
		if !exists {
			return nil, group.E("Store.GetAncestors", group.ErrCodeStoreOperation,
				fmt.Sprintf("ancestor %s not found", ancestorID), nil)
		}
		// Add copy to prevent external modifications
		copy := *ancestor
		ancestors = append(ancestors, &copy)
	}

	return ancestors, nil
}

// GetDescendants retrieves all descendant groups of the specified group
func (s *Store) GetDescendants(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	descendants := make([]*group.Group, 0)
	for _, childID := range g.Ancestry.Children {
		childKey := s.key(tenantID, childID)
		child, exists := s.groups[childKey]
		if !exists {
			return nil, group.E("Store.GetDescendants", group.ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}

		// Add copy to prevent external modifications
		copy := *child
		descendants = append(descendants, &copy)

		// Recursively get descendants
		childDescendants, err := s.GetDescendants(ctx, tenantID, childID)
		if err != nil {
			return nil, fmt.Errorf("get child descendants: %w", err)
		}
		descendants = append(descendants, childDescendants...)
	}

	return descendants, nil
}

// GetChildren retrieves immediate child groups of the specified group
func (s *Store) GetChildren(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, groupID)
	g, exists := s.groups[key]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	children := make([]*group.Group, 0, len(g.Ancestry.Children))
	for _, childID := range g.Ancestry.Children {
		childKey := s.key(tenantID, childID)
		child, exists := s.groups[childKey]
		if !exists {
			return nil, group.E("Store.GetChildren", group.ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}
		// Add copy to prevent external modifications
		copy := *child
		children = append(children, &copy)
	}

	return children, nil
}

// ValidateHierarchy validates the integrity of the group hierarchy for a tenant
func (s *Store) ValidateHierarchy(ctx context.Context, tenantID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all groups for the tenant
	var tenantGroups []*group.Group
	for _, g := range s.groups {
		if g.TenantID == tenantID {
			copy := *g
			tenantGroups = append(tenantGroups, &copy)
		}
	}

	// Build a map for efficient lookup
	groupMap := make(map[string]*group.Group)
	for _, g := range tenantGroups {
		groupMap[g.ID] = g
	}

	// Validate each group's hierarchy information
	for _, g := range tenantGroups {
		// Validate parent relationship
		if g.ParentID != "" {
			parent, exists := groupMap[g.ParentID]
			if !exists {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

			// Verify this group is in parent's children list
			found := false
			for _, childID := range parent.Ancestry.Children {
				if childID == g.ID {
					found = true
					break
				}
			}
			if !found {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("group %s not found in parent %s children list", g.ID, parent.ID), nil)
			}
		}

		// Validate children relationships
		for _, childID := range g.Ancestry.Children {
			child, exists := groupMap[childID]
			if !exists {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}
			if child.ParentID != g.ID {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference parent %s", child.ID, g.ID), nil)
			}
		}

		// Validate ancestry path
		if len(g.Ancestry.PathParts) == 0 || g.Ancestry.PathParts[len(g.Ancestry.PathParts)-1] != g.ID {
			return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid ancestry path", g.ID), nil)
		}

		// Validate depth matches path parts
		if g.Ancestry.Depth != len(g.Ancestry.PathParts)-1 {
			return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
				fmt.Sprintf("group %s has invalid depth", g.ID), nil)
		}
	}

	return nil
}
