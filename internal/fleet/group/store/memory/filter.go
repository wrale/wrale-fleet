package memory

import (
	"github.com/wrale/fleet/internal/fleet/group"
)

// matchesFilter checks if a group matches the filter criteria
func (s *Store) matchesFilter(g *group.Group, opts group.ListOptions) bool {
	// Tenant isolation
	if opts.TenantID != "" && g.TenantID != opts.TenantID {
		return false
	}

	// Parent filtering
	if opts.ParentID != "" && g.ParentID != opts.ParentID {
		return false
	}

	// Type filtering
	if opts.Type != "" && g.Type != opts.Type {
		return false
	}

	// Tag filtering
	for key, value := range opts.Tags {
		if g.Properties.Metadata[key] != value {
			return false
		}
	}

	// Depth filtering
	if opts.Depth >= 0 && g.Ancestry.Depth != opts.Depth {
		return false
	}

	return true
}
