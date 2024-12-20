package memory

import (
	"context"
	"fmt"

	"github.com/wrale/wrale-fleet/internal/fleet/group"
)

// GetAncestors implements group.Store
func (s *Store) GetAncestors(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	ancestors := make([]*group.Group, 0, g.Ancestry.Depth)
	for _, ancestorID := range g.Ancestry.PathParts[:len(g.Ancestry.PathParts)-1] {
		ancestor, exists := tenantGroups[ancestorID]
		if !exists {
			return nil, group.E("Store.GetAncestors", group.ErrCodeStoreOperation,
				fmt.Sprintf("ancestor %s not found", ancestorID), nil)
		}
		ancestors = append(ancestors, ancestor.DeepCopy())
	}

	return ancestors, nil
}

// GetChildren implements group.Store
func (s *Store) GetChildren(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	children := make([]*group.Group, 0, len(g.Ancestry.Children))
	for _, childID := range g.Ancestry.Children {
		child, exists := tenantGroups[childID]
		if !exists {
			return nil, group.E("Store.GetChildren", group.ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}
		children = append(children, child.DeepCopy())
	}

	return children, nil
}

// GetDescendants implements group.Store
func (s *Store) GetDescendants(ctx context.Context, tenantID, groupID string) ([]*group.Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	g, exists := tenantGroups[groupID]
	if !exists {
		return nil, group.ErrGroupNotFound
	}

	descendants := make([]*group.Group, 0)
	for _, childID := range g.Ancestry.Children {
		child, exists := tenantGroups[childID]
		if !exists {
			return nil, group.E("Store.GetDescendants", group.ErrCodeStoreOperation,
				fmt.Sprintf("child %s not found", childID), nil)
		}

		descendants = append(descendants, child.DeepCopy())

		childDescendants, err := s.GetDescendants(ctx, tenantID, childID)
		if err != nil {
			return nil, fmt.Errorf("get child descendants: %w", err)
		}
		descendants = append(descendants, childDescendants...)
	}

	return descendants, nil
}

// ValidateHierarchy implements group.Store
func (s *Store) ValidateHierarchy(ctx context.Context, tenantID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenantGroups, exists := s.groups[tenantID]
	if !exists {
		return nil // No groups for tenant is valid
	}

	for _, g := range tenantGroups {
		if g.ParentID != "" {
			parent, exists := tenantGroups[g.ParentID]
			if !exists {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent parent %s", g.ID, g.ParentID), nil)
			}

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

		for _, childID := range g.Ancestry.Children {
			child, exists := tenantGroups[childID]
			if !exists {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("group %s references non-existent child %s", g.ID, childID), nil)
			}
			if child.ParentID != g.ID {
				return group.E("Store.ValidateHierarchy", group.ErrCodeInvalidGroup,
					fmt.Sprintf("child group %s does not reference parent %s", child.ID, g.ID), nil)
			}
		}
	}

	return nil
}
