package group

import "context"

// HierarchyManager provides operations for managing group hierarchies
type HierarchyManager struct {
	store Store
}

// NewHierarchyManager creates a new hierarchy manager
func NewHierarchyManager(store Store) *HierarchyManager {
	return &HierarchyManager{
		store: store,
	}
}
