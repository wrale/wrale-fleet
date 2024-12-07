package store

import (
	"sync"
)

// Store represents a thread-safe key-value store
type Store struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewStore creates a new store instance
func NewStore() *Store {
	return &Store{
		data: make(map[string]interface{}),
	}
}

// Set stores a value for a key
func (s *Store) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Get retrieves a value by key
func (s *Store) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if value, ok := s.data[key]; ok {
		return value, nil
	}
	return nil, nil
}

// Delete removes a key from the store
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}
