package group

import "sync"

// HierarchyManager provides operations for managing group hierarchies
type HierarchyManager struct {
	store Store
	mu    sync.Mutex // Protects hierarchy modifications
}

// NewHierarchyManager creates a new hierarchy manager
func NewHierarchyManager(store Store) *HierarchyManager {
	return &HierarchyManager{
		store: store,
	}
}
