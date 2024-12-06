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

	// Event correlation tracking
	eventHistory    []SecurityEvent
	tamperAttempts  []TamperAttempt
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

// SecurityEvent tracks individual security-related events
type SecurityEvent struct {
	Timestamp   time.Time
	Type        string
	Source      string
	Severity    string
	State       hw.TamperState
	Context     map[string]interface{}
}

// TamperAttempt tracks potential intrusion patterns
type TamperAttempt struct {
	StartTime    time.Time
	EndTime      time.Time
	EventCount   int
	Pattern      string
	Severity     string
	RelatedEvents []string
}

// StateTransition tracks security state changes
type StateTransition struct {
	Timestamp    time.Time
	FromState    hw.TamperState
	ToState      hw.TamperState
	Trigger      string
	Context      map[string]interface{}
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
		eventHistory:    make([]SecurityEvent, 0, 1000),  // Pre-allocate space for events
		tamperAttempts:  make([]TamperAttempt, 0),
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
			// Attempt recovery from errors
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

				// Log error but continue monitoring
				fmt.Printf("Security check failed: %v\n", err)
			} else {
				m.Lock()
				m.metrics.SuccessfulChecks++
				m.Unlock()
			}

			// Analyze patterns periodically
			m.analyzeTamperPatterns()
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

	// Check for state transitions
	m.detectStateTransition(state)

	// Process through policy manager
	if err := m.policyManager.HandleStateUpdate(ctx, state); err != nil {
		return fmt.Errorf("policy enforcement failed: %w", err)
	}

	return nil
}

// detectStateTransition analyzes state changes
func (m *Monitor) detectStateTransition(newState hw.TamperState) {
	m.Lock()
	defer m.Unlock()

	if len(m.stateTransitions) == 0 || !m.stateEqual(m.stateTransitions[len(m.stateTransitions)-1].ToState, newState) {
		transition := StateTransition{
			Timestamp: time.Now(),
			ToState:   newState,
			Context:   make(map[string]interface{}),
		}

		if len(m.stateTransitions) > 0 {
			transition.FromState = m.stateTransitions[len(m.stateTransitions)-1].ToState
		}

		// Determine transition trigger
		trigger := m.determineTrigger(transition.FromState, newState)
		transition.Trigger = trigger

		m.stateTransitions = append(m.stateTransitions, transition)
		m.metrics.StateTransitions++
		m.lastStateChange = time.Now()
	}
}

// stateEqual compares two TamperStates for equality
func (m *Monitor) stateEqual(a, b hw.TamperState) bool {
	return a.CaseOpen == b.CaseOpen &&
		a.MotionDetected == b.MotionDetected &&
		a.VoltageNormal == b.VoltageNormal
}

// determineTrigger analyzes what caused a state transition
func (m *Monitor) determineTrigger(from, to hw.TamperState) string {
	if from.CaseOpen != to.CaseOpen {
		return "CASE_STATE_CHANGE"
	}
	if from.MotionDetected != to.MotionDetected {
		return "MOTION_DETECTED"
	}
	if from.VoltageNormal != to.VoltageNormal {
		return "VOLTAGE_CHANGE"
	}
	return "UNKNOWN"
}

// recordSecurityEvent adds an event to the history
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

	// Trim history if needed
	if len(m.eventHistory) > 1000 {
		m.eventHistory = m.eventHistory[1:]
	}
}

// analyzeTamperPatterns looks for patterns in security events
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

	// If we see many transitions in a short time, record an attempt
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

// attemptErrorRecovery tries to recover from error states
func (m *Monitor) attemptErrorRecovery() {
	m.Lock()
	defer m.Unlock()

	if m.metrics.ErrorCount > 0 && time.Since(m.metrics.LastErrorTime) > m.retryDelay {
		// Reset error count if we've been stable
		if time.Since(m.metrics.LastErrorTime) > m.retryDelay*5 {
			m.metrics.ErrorCount = 0
			m.recordSecurityEvent("ERROR_STATE_RECOVERED", "info", nil)
		}
	}
}

// shutdown handles graceful shutdown
func (m *Monitor) shutdown(hwCancel context.CancelFunc, hwErrCh chan error) error {
	// Record shutdown event
	m.recordSecurityEvent("MONITOR_SHUTDOWN", "info", nil)

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

// GetTamperAttempts returns recent tamper attempts
func (m *Monitor) GetTamperAttempts(since time.Time) []TamperAttempt {
	m.RLock()
	defer m.RUnlock()

	var attempts []TamperAttempt
	for _, attempt := range m.tamperAttempts {
		if attempt.EndTime.After(since) {
			attempts = append(attempts, attempt)
		}
	}
	return attempts
}

// GetStateTransitions returns recent state transitions
func (m *Monitor) GetStateTransitions(since time.Time) []StateTransition {
	m.RLock()
	defer m.RUnlock()

	var transitions []StateTransition
	for _, transition := range m.stateTransitions {
		if transition.Timestamp.After(since) {
			transitions = append(transitions, transition)
		}
	}
	return transitions
}

// max returns the larger of x or y
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
