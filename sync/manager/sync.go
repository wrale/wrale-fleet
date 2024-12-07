package manager

import (
	"fmt"
	"sync"
)

// Config defines configuration for the sync manager
type Config struct {
	StoragePath string
	MaxRetries  int
	Timeout     int
}

// SyncManager handles state synchronization
type SyncManager struct {
	config Config
	mu     sync.RWMutex
}

// New creates a new sync manager instance
func New(config Config) (*SyncManager, error) {
	if config.StoragePath == "" {
		return nil, fmt.Errorf("storage path required")
	}

	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	return &SyncManager{
		config: config,
	}, nil
}
