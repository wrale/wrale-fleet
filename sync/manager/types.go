package manager

import "time"
import "sync"

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
	mu     sync.RWMutex
}
