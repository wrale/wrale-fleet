// Package memory provides an in-memory implementation of the logging store
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/logging"
)

// Store implements the logging.Store interface with in-memory storage
type Store struct {
	mu     sync.RWMutex
	events map[string]map[string]*logging.Event // tenant -> id -> event
}

// New creates a new in-memory event store
func New() *Store {
	return &Store{
		events: make(map[string]map[string]*logging.Event),
	}
}

// Store stores a new event
func (s *Store) Store(ctx context.Context, event *logging.Event) error {
	if err := event.Validate(); err != nil {
		return logging.E("Store.Store", logging.ErrCodeValidation, "invalid event", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.events[event.TenantID] == nil {
		s.events[event.TenantID] = make(map[string]*logging.Event)
	}
	s.events[event.TenantID][event.ID] = event

	return nil
}

// Get retrieves an event by ID
func (s *Store) Get(ctx context.Context, tenantID, eventID string) (*logging.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tenant, exists := s.events[tenantID]; exists {
		if event, exists := tenant[eventID]; exists {
			return event, nil
		}
	}

	return nil, logging.E("Store.Get", logging.ErrCodeNotFound, "event not found", logging.ErrEventNotFound)
}

// List retrieves events matching the given options
func (s *Store) List(ctx context.Context, opts logging.ListOptions) ([]*logging.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*logging.Event

	tenant, exists := s.events[opts.TenantID]
	if !exists {
		return results, nil
	}

	// Collect all matching events
	for _, event := range tenant {
		if matchesListOptions(event, opts) {
			results = append(results, event)
		}
	}

	// Sort by timestamp descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// Apply pagination
	if opts.Offset >= len(results) {
		return []*logging.Event{}, nil
	}

	end := opts.Offset + opts.Limit
	if end > len(results) || opts.Limit == 0 {
		end = len(results)
	}

	return results[opts.Offset:end], nil
}

// Delete removes an event
func (s *Store) Delete(ctx context.Context, tenantID, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tenant, exists := s.events[tenantID]; exists {
		if _, exists := tenant[eventID]; exists {
			delete(tenant, eventID)
			return nil
		}
	}

	return logging.E("Store.Delete", logging.ErrCodeNotFound, "event not found", logging.ErrEventNotFound)
}

// DeleteBefore removes events older than the given time
func (s *Store) DeleteBefore(ctx context.Context, tenantID string, before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tenant, exists := s.events[tenantID]
	if !exists {
		return nil
	}

	for id, event := range tenant {
		if event.Timestamp.Before(before) {
			delete(tenant, id)
		}
	}

	return nil
}

// Query performs a structured query on events
func (s *Store) Query(ctx context.Context, query logging.QueryOptions) ([]*logging.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*logging.Event

	tenant, exists := s.events[query.TenantID]
	if !exists {
		return results, nil
	}

	// Collect all matching events
	for _, event := range tenant {
		if matchesQueryOptions(event, query) {
			results = append(results, event)
		}
	}

	// Apply sorting
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

// BatchStore stores multiple events in a single operation
func (s *Store) BatchStore(ctx context.Context, events []*logging.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		if err := event.Validate(); err != nil {
			return logging.E("Store.BatchStore", logging.ErrCodeValidation, "invalid event", err)
		}

		if s.events[event.TenantID] == nil {
			s.events[event.TenantID] = make(map[string]*logging.Event)
		}
		s.events[event.TenantID][event.ID] = event
	}

	return nil
}

// Sync ensures all events are persisted
func (s *Store) Sync(ctx context.Context) error {
	// No-op for memory store as all operations are synchronous
	return nil
}

// matchesListOptions checks if an event matches the list options
func matchesListOptions(event *logging.Event, opts logging.ListOptions) bool {
	if opts.Type != "" && event.Type != opts.Type {
		return false
	}
	if opts.Level != "" && event.Level != opts.Level {
		return false
	}
	if opts.Source != "" && event.Source != opts.Source {
		return false
	}
	if opts.StartTime != nil && event.Timestamp.Before(*opts.StartTime) {
		return false
	}
	if opts.EndTime != nil && event.Timestamp.After(*opts.EndTime) {
		return false
	}
	if opts.ComponentID != "" && event.Context.ComponentID != opts.ComponentID {
		return false
	}
	if opts.DeviceID != "" && event.Context.DeviceID != opts.DeviceID {
		return false
	}
	if len(opts.Tags) > 0 {
		for k, v := range opts.Tags {
			if event.Tags[k] != v {
				return false
			}
		}
	}
	return true
}

// matchesQueryOptions checks if an event matches the query options
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

	// Check tag query
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

	// Check context query
	if query.ContextQuery != nil {
		if len(query.ContextQuery.ComponentIDs) > 0 {
			componentMatch := false
			for _, id := range query.ContextQuery.ComponentIDs {
				if event.Context.ComponentID == id {
					componentMatch = true
					break
				}
			}
			if !componentMatch {
				return false
			}
		}

		if len(query.ContextQuery.DeviceIDs) > 0 {
			deviceMatch := false
			for _, id := range query.ContextQuery.DeviceIDs {
				if event.Context.DeviceID == id {
					deviceMatch = true
					break
				}
			}
			if !deviceMatch {
				return false
			}
		}

		if query.ContextQuery.MinStage > 0 && event.Context.Stage < query.ContextQuery.MinStage {
			return false
		}
		if query.ContextQuery.MaxStage > 0 && event.Context.Stage > query.ContextQuery.MaxStage {
			return false
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
