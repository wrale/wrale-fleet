package testing

import (
	"context"
	"sort"

	"github.com/wrale/wrale-fleet/internal/fleet/logging"
)

// Query performs a structured query
func (s *Store) Query(ctx context.Context, query logging.QueryOptions) ([]*logging.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*logging.Event

	tenant, exists := s.events[query.TenantID]
	if !exists {
		return results, nil
	}

	// Collect matching events
	for _, event := range tenant {
		if matchesQueryOptions(event, query) {
			results = append(results, event)
		}
	}

	// Sort results
	sortEvents(results, query.OrderBy, query.OrderDirection)

	// Apply pagination
	if query.Offset >= len(results) {
		return []*logging.Event{}, nil
	}

	end := query.Offset + query.Limit
	if end > len(results) || query.Limit == 0 {
		end = len(results)
	}

	return results[query.Offset:end], nil
}

// matchesQueryOptions checks if an event matches query criteria
func matchesQueryOptions(event *logging.Event, query logging.QueryOptions) bool {
	// Check event type
	if len(query.Types) > 0 {
		typeMatch := false
		for _, t := range query.Types {
			if event.Type == t {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			return false
		}
	}

	// Check level
	if len(query.Levels) > 0 {
		levelMatch := false
		for _, l := range query.Levels {
			if event.Level == l {
				levelMatch = true
				break
			}
		}
		if !levelMatch {
			return false
		}
	}

	// Check time range
	if query.TimeRange != nil {
		if event.Timestamp.Before(query.TimeRange.Start) ||
			event.Timestamp.After(query.TimeRange.End) {
			return false
		}
	}

	// Check sources
	if len(query.Sources) > 0 {
		sourceMatch := false
		for _, s := range query.Sources {
			if event.Source == s {
				sourceMatch = true
				break
			}
		}
		if !sourceMatch {
			return false
		}
	}

	// Check tags
	if query.TagQuery != nil {
		// Must match all required tags
		for k, v := range query.TagQuery.Must {
			if event.Tags[k] != v {
				return false
			}
		}

		// Must not match any excluded tags
		for k, v := range query.TagQuery.MustNot {
			if event.Tags[k] == v {
				return false
			}
		}

		// Should match at least one optional tag if specified
		if len(query.TagQuery.Should) > 0 {
			shouldMatch := false
			for k, v := range query.TagQuery.Should {
				if event.Tags[k] == v {
					shouldMatch = true
					break
				}
			}
			if !shouldMatch {
				return false
			}
		}
	}

	return true
}

// sortEvents sorts events based on query options
func sortEvents(events []*logging.Event, orderBy, direction string) {
	if orderBy == "" {
		orderBy = "timestamp"
	}
	if direction == "" {
		direction = "desc"
	}

	sort.Slice(events, func(i, j int) bool {
		var less bool
		switch orderBy {
		case "timestamp":
			less = events[i].Timestamp.Before(events[j].Timestamp)
		case "level":
			less = string(events[i].Level) < string(events[j].Level)
		case "type":
			less = string(events[i].Type) < string(events[j].Type)
		default:
			less = events[i].Timestamp.Before(events[j].Timestamp)
		}

		if direction == "asc" {
			return less
		}
		return !less
	})
}
