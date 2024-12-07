package secure

import (
	"context"
	"fmt"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/secure"
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

	// Event correlation tracking
	eventHistory     []SecurityEvent
	tamperAttempts   []TamperAttempt
	stateTransitions []StateTransition
	lastStateChange  time.Time
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
	CheckCount        uint64
	ErrorCount        uint64
	LastError         error
	LastErrorTime     time.Time
	LastSyncTime      time.Time
	UptimeSeconds     uint64
	PolicyVersions    []string
	TamperAttempts    uint64
	SuccessfulChecks  uint64
	StateTransitions  uint64
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
		hwManager:        cfg.HWManager,
		policyManager:    cfg.PolicyManager,
		stateStore:       cfg.StateStore,
		deviceID:         cfg.DeviceID,
		monitorInterval:  cfg.MonitorInterval,
		retryDelay:       cfg.RetryDelay,
		maxRetries:       cfg.MaxRetries,
		shutdownTimeout:  cfg.ShutdownTimeout,
		eventHistory:     make([]SecurityEvent, 0, 1000),
		tamperAttempts:   make([]TamperAttempt, 0),
		stateTransitions: make([]StateTransition, 0),
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
	
	// Initialize recovery timer for transient errors
	recoveryTicker := time.NewTicker(m.retryDelay * 2)
	defer recoveryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return m.shutdown(hwCancel, hwErrCh)

		case err := <-hwErrCh:
			if err != nil {
				m.recordSecurityEvent("HARDWARE_FAILURE", "critical", nil)
				return fmt.Errorf("hardware monitor failed: %w", err)
			}
			return nil

		case <-recoveryTicker.C:
			m.attemptErrorRecovery()

		case <-ticker.C:
			m.metrics.UptimeSeconds = uint64(time.Since(startTime).Seconds())
			
			if err := m.check(ctx); err != nil {
				m.Lock()
				m.metrics.ErrorCount++
				m.metrics.LastError = err
				m.metrics.LastErrorTime = time.Now()
				m.Unlock()

				m.recordSecurityEvent("CHECK_FAILURE", "warning", map[string]interface{}{
					"error": err.Error(),
				})

				fmt.Printf("Security check failed: %v\n", err)
			} else {
				m.Lock()
				m.metrics.SuccessfulChecks++
				m.Unlock()
			}

			m.analyzeTamperPatterns()
		}
	}
}

func (m *Monitor) check(ctx context.Context) error {
	m.Lock()
	m.metrics.CheckCount++
	m.Unlock()

	state := m.hwManager.GetState()
	m.detectStateTransition(state)

	if err := m.policyManager.HandleStateUpdate(ctx, state); err != nil {
		return fmt.Errorf("policy enforcement failed: %w", err)
	}

	return nil
}

func (m *Monitor) shutdown(hwCancel context.CancelFunc, hwErrCh chan error) error {
	m.recordSecurityEvent("MONITOR_SHUTDOWN", "info", nil)
	hwCancel()

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

// Helper methods
func (m *Monitor) recordSecurityEvent(eventType string, severity string, context map[string]interface{}) {
	m.Lock()
	defer m.Unlock()

	event := SecurityEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Source:    m.deviceID,
		Severity:  severity,
		State:     m.hwManager.GetState(),
		Context:   context,
	}

	m.eventHistory = append(m.eventHistory, event)
	if len(m.eventHistory) > 1000 {
		m.eventHistory = m.eventHistory[1:]
	}
}

func (m *Monitor) detectStateTransition(newState hw.TamperState) {
	m.Lock()
	defer m.Unlock()

	if len(m.stateTransitions) > 0 && m.stateEqual(m.stateTransitions[len(m.stateTransitions)-1].ToState, newState) {
		return
	}

	transition := StateTransition{
		Timestamp: time.Now(),
		ToState:   newState,
		Context:   make(map[string]interface{}),
	}

	if len(m.stateTransitions) > 0 {
		transition.FromState = m.stateTransitions[len(m.stateTransitions)-1].ToState
	}

	transition.Trigger = m.determineTrigger(transition.FromState, newState)
	m.stateTransitions = append(m.stateTransitions, transition)
	m.metrics.StateTransitions++
	m.lastStateChange = time.Now()
}

func (m *Monitor) stateEqual(a, b hw.TamperState) bool {
	return a.CaseOpen == b.CaseOpen &&
		a.MotionDetected == b.MotionDetected &&
		a.VoltageNormal == b.VoltageNormal
}

func (m *Monitor) determineTrigger(from, to hw.TamperState) string {
	switch {
	case from.CaseOpen != to.CaseOpen:
		return "CASE_STATE_CHANGE"
	case from.MotionDetected != to.MotionDetected:
		return "MOTION_DETECTED"
	case from.VoltageNormal != to.VoltageNormal:
		return "VOLTAGE_CHANGE"
	default:
		return "UNKNOWN"
	}
}

func (m *Monitor) analyzeTamperPatterns() {
	m.Lock()
	defer m.Unlock()

	if len(m.eventHistory) < 2 {
		return
	}

	// Look for rapid transitions
	recentEvents := m.eventHistory[max(0, len(m.eventHistory)-10):]
	transitions := 0
	lastEventType := ""

	for _, event := range recentEvents {
		if event.Type != lastEventType {
			transitions++
			lastEventType = event.Type
		}
	}

	if transitions > 5 {
		attempt := TamperAttempt{
			StartTime:  recentEvents[0].Timestamp,
			EndTime:    recentEvents[len(recentEvents)-1].Timestamp,
			EventCount: transitions,
			Pattern:    "RAPID_TRANSITIONS",
			Severity:   "WARNING",
		}

		m.tamperAttempts = append(m.tamperAttempts, attempt)
		m.metrics.TamperAttempts++

		m.recordSecurityEvent("TAMPER_PATTERN_DETECTED", "warning", map[string]interface{}{
			"pattern": "RAPID_TRANSITIONS",
			"count":   transitions,
		})
	}
}

func (m *Monitor) attemptErrorRecovery() {
	m.Lock()
	defer m.Unlock()

	if m.metrics.ErrorCount > 0 && time.Since(m.metrics.LastErrorTime) > m.retryDelay {
		if time.Since(m.metrics.LastErrorTime) > m.retryDelay*5 {
			m.metrics.ErrorCount = 0
			m.recordSecurityEvent("ERROR_STATE_RECOVERED", "info", nil)
		}
	}
}

// Helper function
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}