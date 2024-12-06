package secure

import (
	"context"
	"fmt"
	"time"
)

// Monitor starts continuous security monitoring of hardware
func (m *Manager) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.checkSecurity(ctx); err != nil {
				return fmt.Errorf("security check failed: %w", err)
			}
		}
	}
}

// checkSecurity performs raw hardware security checks
func (m *Manager) checkSecurity(ctx context.Context) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Check case sensor
	caseOpen, err := m.gpio.GetPinState(m.caseSensor)
	if err != nil {
		return fmt.Errorf("failed to check case sensor: %w", err)
	}

	// Check motion sensor
	motion, err := m.gpio.GetPinState(m.motionSensor)
	if err != nil {
		return fmt.Errorf("failed to check motion sensor: %w", err)
	}

	// Check voltage sensor
	voltageOK, err := m.gpio.GetPinState(m.voltSensor)
	if err != nil {
		return fmt.Errorf("failed to check voltage sensor: %w", err)
	}

	// Update state
	m.state = TamperState{
		CaseOpen:       caseOpen,
		MotionDetected: motion,
		VoltageNormal:  voltageOK,
		LastCheck:      time.Now(),
	}

	// Notify of raw state changes through callback
	if m.onTamper != nil && (caseOpen || motion || !voltageOK) {
		m.onTamper(m.state)
	}

	return nil
}
