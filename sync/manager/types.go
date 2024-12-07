package manager

import "time"

// Config holds sync manager configuration
type Config struct {
	StoragePath   string
	MaxRetries    int
	Timeout       time.Duration
	RetryInterval time.Duration
}

// SyncManager handles state synchronization between components
type SyncManager struct {
	config Config
}

// New creates a new sync manager instance
func New(cfg Config) (*SyncManager, error) {
	return &SyncManager{
		config: cfg,
	}, nil
}