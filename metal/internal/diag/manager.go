package diag

import (
	"fmt"
	"sync"
	"time"
)

// Manager handles hardware diagnostics and testing
type Manager struct {
	mux sync.RWMutex
	cfg Config

	// Test history
	results []TestResult
}

// New creates a new hardware diagnostics manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller required")
	}

	// Set defaults
	if cfg.RetryAttempts == 0 {
		cfg.RetryAttempts = 3
	}
	if cfg.LoadTestTime == 0 {
		cfg.LoadTestTime = 30 * time.Second
	}
	if cfg.MinVoltage == 0 {
		cfg.MinVoltage = 4.8 // 4.8V minimum for 5V system
	}
	if cfg.TempRange == [2]float64{} {
		cfg.TempRange = [2]float64{-10, 50} // -10°C to 50°C
	}

	return &Manager{
		cfg: cfg,
	}, nil
}
