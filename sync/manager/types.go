// Package manager provides synchronized state management across fleet components.
package manager

import (
	"fmt"
	"sync"
	"time"
)

// Config holds sync manager configuration options
type Config struct {
	StoragePath   string        `json:"storage_path"`
	MaxRetries    int           `json:"max_retries"`
	Timeout       time.Duration `json:"timeout"`
	RetryInterval time.Duration `json:"retry_interval"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.StoragePath == "" {
		return fmt.Errorf("storage path required")
	}
	return nil
}

// SyncManager handles state synchronization between components
type SyncManager struct {
	config Config
	mu     sync.RWMutex
}

// New creates a new sync manager instance with the given configuration
func New(config Config) (*SyncManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	// Set defaults
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = time.Second
	}

	return &SyncManager{
		config: config,
		mu:     sync.RWMutex{},
	}, nil
}