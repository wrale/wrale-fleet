// Package testing provides testing utilities for the logging package
package testing

import (
	"context"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/internal/fleet/logging"
)

// Store implements logging.Store for testing
type Store struct {
	mu     sync.RWMutex
	events map[string]map[string]*logging.Event // tenant -> id -> event
}

// NewStore creates a new test store
func NewStore() *Store {
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

// BatchStore stores multiple events
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

// Sync ensures events are persisted
func (s *Store) Sync(ctx context.Context) error {
	return nil // No-op for test store
}
