package memory

import "github.com/wrale/fleet/internal/fleet/config"

// validateInput is a helper function to check for required string fields
func (s *Store) validateInput(op string, fields map[string]string) error {
	for name, value := range fields {
		if value == "" {
			return config.NewError(op, config.ErrValidationFailed, name+" is required")
		}
	}
	return nil
}

// templateKey generates a unique key for template storage
func (s *Store) templateKey(tenantID, templateID string) string {
	return tenantID + "/" + templateID
}

// deploymentKey generates a unique key for deployment storage
func (s *Store) deploymentKey(tenantID, deploymentID string) string {
	return tenantID + "/" + deploymentID
}

// applyPagination applies offset and limit to a slice
func (s *Store) applyPagination(length int, opts config.ListOptions) (start, end int) {
	if opts.Offset >= length {
		return 0, 0
	}

	start = opts.Offset
	end = opts.Offset + opts.Limit
	if end > length || opts.Limit == 0 {
		end = length
	}
	return start, end
}
