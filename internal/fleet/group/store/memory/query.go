package memory

import (
	"github.com/wrale/fleet/internal/fleet/group"
)

// matchesFilter checks if a group matches the provided filter options
func (s *Store) matchesFilter(g *group.Group, opts group.ListOptions) bool {
	// Parent ID filter
	if opts.ParentID != "" && g.ParentID != opts.ParentID {
		return false
	}

	// Group type filter
	if opts.Type != "" && g.Type != opts.Type {
		return false
	}

	// Tags filter
	if len(opts.Tags) > 0 {
		for k, v := range opts.Tags {
			if gv, ok := g.Properties.Metadata[k]; !ok || gv != v {
				return false
			}
		}
	}

	return true
}
