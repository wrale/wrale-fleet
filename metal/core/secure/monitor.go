package secure

import (
	"context"
	"fmt"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/secure"
)

// Monitor coordinates security monitoring and policy enforcement
type Monitor struct {
	sync.RWMutex

	// Core components
	hwManager     *hw.Manager
	policyManager *PolicyManager
	stateStore    StateStore

	// Configuration
	deviceID         string
	monitorInterval  time.Duration
	retryDelay      time.Duration
	maxRetries      int
	shutdownTimeout time.Duration

	// Runtime state
	running  bool
	metrics  MonitorMetrics
	lastSync time.Time
}

// MonitorConfig holds monitor configuration
type MonitorConfig struct {
	HWManager       *hw.Manager
	PolicyManager   *PolicyManager
	StateStore      StateStore
	DeviceID        string
	MonitorInterval time.Duration
	RetryDelay      time.Duration
	MaxRetries      int
	ShutdownTimeout time.Duration
}

// MonitorMetrics tracks monitoring statistics
type MonitorMetrics struct {
	CheckCount     uint64
	ErrorCount     uint64
	LastError      error
	LastErrorTime  time.Time
	LastSyncTime   time.Time
	UptimeSeconds  uint64
	PolicyVersions []string
}

// NewMonitor creates a new security monitor
func NewMonitor(cfg MonitorConfig) (*Monitor, error) {
	if cfg.HWManager == nil {
		return nil, fmt.Errorf("hardware manager is required")
	}
	if cfg.PolicyManager == nil {
		return nil, fmt.Errorf("policy manager is required")
	}
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	// Set defaults
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = 1 * time.Second
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 5 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}

	return &Monitor{
		hwManager:       cfg.HWManager,
		policyManager:   cfg.PolicyManager,
		stateStore:      cfg.StateStore,
		deviceID:        cfg.DeviceID,
		monitorInterval: cfg.MonitorInterval,
		retryDelay:      cfg.RetryDelay,
		maxRetries:      cfg.MaxRetries,
		shutdownTimeout: cfg.ShutdownTimeout,
	}, nil
}

// Start begins security monitoring
func (m *Monitor) Start(ctx context.Context) error {
	m.Lock()
	if m.running {
		m.Unlock()
		return fmt.Errorf("monitor already running")
	}
	m.running = true
	m.Unlock()

	// Start hardware monitoring
	hwCtx, hwCancel := context.WithCancel(ctx)
	defer hwCancel()

	hwErrCh := make(chan error, 1)
	go func() {
		hwErrCh <- m.hwManager.Monitor(hwCtx)
	}()

	// Main monitoring loop
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return m.shutdown(hwCancel, hwErrCh)

		case err := <-hwErrCh:
			if err != nil {
				return fmt.Errorf("hardware monitor failed: %w", err)
			}
			return nil

		case <-ticker.C:
			m.metrics.UptimeSeconds = uint64(time.Since(startTime).Seconds())
			
			if err := m.check(ctx); err != nil {
				m.Lock()
				m.metrics.ErrorCount++
				m.metrics.LastError = err
				m.metrics.LastErrorTime = time.Now()
				m.Unlock()

				// Log error but continue monitoring
				fmt.Printf("Security check failed: %v\n", err)
			}
		}
	}
}

// check performs a single monitoring cycle
func (m *Monitor) check(ctx context.Context) error {
	m.Lock()
	m.metrics.CheckCount++
	m.Unlock()

	// Get current hardware state
	state := m.hwManager.GetState()

	// Process through policy manager
	if err := m.policyManager.HandleStateUpdate(ctx, state); err != nil {
		return fmt.Errorf("policy enforcement failed: %w", err)
	}

	return nil
}

// shutdown handles graceful shutdown
func (m *Monitor) shutdown(hwCancel context.CancelFunc, hwErrCh chan error) error {
	// Cancel hardware monitoring
	hwCancel()

	// Wait for hardware monitor to stop
	select {
	case err := <-hwErrCh:
		if err != nil {
			return fmt.Errorf("hardware monitor failed during shutdown: %w", err)
		}
	case <-time.After(m.shutdownTimeout):
		return fmt.Errorf("hardware monitor shutdown timed out")
	}

	m.Lock()
	m.running = false
	m.Unlock()

	return nil
}

// GetMetrics returns current monitoring metrics
func (m *Monitor) GetMetrics() MonitorMetrics {
	m.RLock()
	defer m.RUnlock()
	return m.metrics
}

// IsRunning returns whether the monitor is currently running
func (m *Monitor) IsRunning() bool {
	m.RLock()
	defer m.RUnlock()
	return m.running
}