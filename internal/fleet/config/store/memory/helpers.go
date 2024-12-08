package memory

import (
	"fmt"

	"github.com/wrale/fleet/internal/fleet/config"
)

// validateInput checks that required string fields are not empty.
// It returns an error if any required field is empty.
func (s *Store) validateInput(op string, fields map[string]string) error {
	for field, value := range fields {
		if value == "" {
			return config.NewError(op, config.ErrInvalidInput, fmt.Sprintf("%s is required", field))
		}
	}
	return nil
}

// templateKey generates a composite key for template storage.
// It combines tenantID and templateID in a consistent format.
func (s *Store) templateKey(tenantID, templateID string) string {
	return fmt.Sprintf("%s/%s", tenantID, templateID)
}

// deploymentKey generates a composite key for deployment storage.
// It combines tenantID and deploymentID in a consistent format.
func (s *Store) deploymentKey(tenantID, deploymentID string) string {
	return fmt.Sprintf("%s/%s", tenantID, deploymentID)
}

// applyPagination calculates the correct slice bounds based on offset and limit.
// If no limit is specified (limit <= 0), it returns the full range.
// If offset is beyond the available range, it returns empty range (0, 0).
func (s *Store) applyPagination(total int, opts config.ListOptions) (start, end int) {
	// If no limit specified, return all items
	if opts.Limit <= 0 {
		return 0, total
	}

	// Calculate start index
	start = opts.Offset
	if start >= total {
		return 0, 0
	}

	// Calculate end index
	end = start + opts.Limit
	if end > total {
		end = total
	}

	return start, end
}
