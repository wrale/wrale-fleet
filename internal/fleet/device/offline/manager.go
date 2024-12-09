package offline

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// offlineManager implements the Manager interface
type offlineManager struct {
	mu           sync.RWMutex
	capabilities *Capabilities
	logger       *zap.Logger
	lastSync     time.Time
}

// NewManager creates a new offline capabilities manager
func NewManager(logger *zap.Logger) Manager {
	return &offlineManager{
		logger: logger.With(zap.String("component", "offline_manager")),
	}
}

// GetCapabilities retrieves the current offline capabilities
func (m *offlineManager) GetCapabilities(ctx context.Context) (*Capabilities, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.capabilities == nil {
		return nil, newError(ErrNotSupported, "offline capabilities not configured")
	}

	return m.capabilities, nil
}

// UpdateCapabilities updates the offline capabilities configuration
func (m *offlineManager) UpdateCapabilities(ctx context.Context, caps *Capabilities) error {
	if caps == nil {
		return newError(ErrInvalidConfig, "capabilities cannot be nil")
	}

	// Validate all components
	if err := validateBufferSize(caps.LocalBufferSize); err != nil {
		return err
	}
	if err := validateSyncInterval(caps.SyncInterval); err != nil {
		return err
	}
	if err := validateOperations(caps.SupportedOperations); err != nil {
		return err
	}
	if err := validateSyncSchedule(caps.SyncSchedule); err != nil {
		return err
	}
	if caps.BufferStats != nil {
		if err := validateBufferStats(caps.BufferStats); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update capabilities
	now := time.Now().UTC()
	if caps.SyncStatus == nil {
		caps.SyncStatus = &SyncStatus{
			LastAttempt: now,
			NextSync:    now.Add(caps.SyncInterval),
			Interval:    caps.SyncInterval,
		}
	}

	m.capabilities = caps
	m.lastSync = now

	m.logger.Info("updated offline capabilities",
		zap.Bool("supports_airgap", caps.SupportsAirgap),
		zap.Duration("sync_interval", caps.SyncInterval),
		zap.Int64("buffer_size", caps.LocalBufferSize),
		zap.Int("operations", len(caps.SupportedOperations)),
	)

	return nil
}

// IsSyncDue checks if synchronization is due based on schedule
func (m *offlineManager) IsSyncDue(ctx context.Context) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.capabilities == nil {
		return false, newError(ErrNotSupported, "offline capabilities not configured")
	}

	now := time.Now().UTC()

	// Check if we're within a scheduled sync window
	if len(m.capabilities.SyncSchedule) > 0 {
		return m.isInSyncWindow(now), nil
	}

	// Fall back to interval-based check
	nextSync := m.lastSync.Add(m.capabilities.SyncInterval)
	return now.After(nextSync), nil
}

// isInSyncWindow checks if the current time falls within a scheduled sync window
func (m *offlineManager) isInSyncWindow(now time.Time) bool {
	if m.capabilities == nil {
		return false
	}

	day := strings.ToLower(now.Weekday().String())
	timeRange, exists := m.capabilities.SyncSchedule[day]
	if !exists {
		return false
	}

	// Parse time range
	times := strings.Split(timeRange, "-")
	if len(times) != 2 {
		m.logger.Error("invalid time range format",
			zap.String("time_range", timeRange))
		return false
	}

	start, err := parseTimeString(times[0])
	if err != nil {
		m.logger.Error("invalid start time",
			zap.String("time", times[0]),
			zap.Error(err))
		return false
	}

	end, err := parseTimeString(times[1])
	if err != nil {
		m.logger.Error("invalid end time",
			zap.String("time", times[1]),
			zap.Error(err))
		return false
	}

	// Check if current time is within window
	currentTime := time.Date(now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), 0, 0, time.UTC)
	return currentTime.After(start) && currentTime.Before(end)
}

// Sync performs a synchronization operation
func (m *offlineManager) Sync(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.capabilities == nil {
		return newError(ErrNotSupported, "offline capabilities not configured")
	}

	if !m.capabilities.SupportsAirgap {
		return newError(ErrNotSupported, "device does not support airgapped operation")
	}

	now := time.Now().UTC()
	nextSync := now.Add(m.capabilities.SyncInterval)

	// Update sync status
	m.capabilities.SyncStatus = &SyncStatus{
		LastSuccess: now,
		LastAttempt: now,
		NextSync:    nextSync,
		Interval:    m.capabilities.SyncInterval,
	}

	m.lastSync = now

	m.logger.Info("sync completed",
		zap.Time("next_sync", nextSync))

	return nil
}

// UpdateBufferStats updates the local buffer statistics
func (m *offlineManager) UpdateBufferStats(ctx context.Context, stats *BufferStats) error {
	if stats == nil {
		return newError(ErrInvalidConfig, "buffer stats cannot be nil")
	}

	if err := validateBufferStats(stats); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.capabilities == nil {
		return newError(ErrNotSupported, "offline capabilities not configured")
	}

	// Update stats
	m.capabilities.BufferStats = stats

	// Log if buffer is getting full
	usagePercent := float64(stats.UsedSize) / float64(stats.TotalSize) * 100
	if usagePercent > 80 {
		m.logger.Warn("buffer usage high",
			zap.Float64("usage_percent", usagePercent),
			zap.Int64("used_size", stats.UsedSize),
			zap.Int64("total_size", stats.TotalSize))
	}

	return nil
}

// IsOperationSupported checks if an operation can be performed offline
func (m *offlineManager) IsOperationSupported(ctx context.Context, op Operation) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.capabilities == nil {
		return false, newError(ErrNotSupported, "offline capabilities not configured")
	}

	// Check operation type validity
	switch op {
	case OpStatusUpdate, OpMetricCollection, OpLogCollection,
		OpConfigValidation, OpHealthCheck:
		// Valid operation type
	default:
		return false, newError(ErrInvalidOperation, "invalid operation type")
	}

	// Check if operation is supported
	for _, supported := range m.capabilities.SupportedOperations {
		if supported == op {
			return true, nil
		}
	}

	return false, nil
}
