package diag

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/gpio"
	"github.com/wrale/wrale-fleet/metal/hw/power"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
	"github.com/wrale/wrale-fleet/metal/hw/secure"
)

// Manager handles hardware diagnostics and testing
type Manager struct {
	mux sync.RWMutex
	cfg Config

	// Test history
	results []TestResult
}

// Config contains the diagnostics manager configuration
type Config struct {
	GPIO         *gpio.Controller       // GPIO controller
	GPIOPins     map[string]int        // Map of pin names to numbers
	Power        power.Manager         // Power subsystem manager
	Thermal      thermal.Monitor       // Thermal subsystem monitor
	Security     secure.Manager        // Security subsystem manager
	Retries      int                   // Number of test retries
	LoadTestTime time.Duration         // Duration for load tests
	MinVoltage   float64              // Minimum acceptable voltage
	TempRange    [2]float64           // Acceptable temperature range [min, max]
	OnTestComplete func(TestResult)    // Callback for test completion
}

// TestResult contains the result of a diagnostic test
type TestResult struct {
	Type        TestType    // Type of test performed
	Component   string      // Component being tested
	Status      TestStatus  // Test result status
	Description string      // Test description
	Reading     float64     // Actual reading (if applicable)
	Expected    float64     // Expected value (if applicable) 
	Error       error       // Error details if test failed
	Timestamp   time.Time   // When the test was performed
}

// TestType identifies the type of diagnostic test
type TestType string

const (
	TestGPIO     TestType = "gpio"
	TestPower    TestType = "power"
	TestThermal  TestType = "thermal"
	TestSecurity TestType = "security"
)

// TestStatus represents the result status of a diagnostic test
type TestStatus string

const (
	StatusPass    TestStatus = "pass"
	StatusFail    TestStatus = "fail"
	StatusWarning TestStatus = "warning"
)

// New creates a new hardware diagnostics manager
func New(cfg Config) (*Manager, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller required")
	}

	// Set defaults
	if cfg.Retries == 0 {
		cfg.Retries = 3
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

// TestGPIO performs GPIO pin diagnostics
func (m *Manager) TestGPIO(ctx context.Context) error {
	for pinName, pinNum := range m.cfg.GPIOPins {
		// Test output mode
		pinId := fmt.Sprintf("%s_%d", pinName, pinNum)
		if err := m.cfg.GPIO.SetPinState(pinId, true); err != nil {
			m.recordResult(TestResult{
				Type:        TestGPIO,
				Component:   pinId,
				Status:      StatusFail,
				Description: "Failed to set pin HIGH",
				Error:       err,
				Timestamp:   time.Now(),
			})
			return fmt.Errorf("failed to set pin %s HIGH: %w", pinId, err)
		}

		// Verify state
		state, err := m.cfg.GPIO.GetPinState(pinId)
		if err != nil || !state {
			m.recordResult(TestResult{
				Type:        TestGPIO,
				Component:   pinId,
				Status:      StatusFail,
				Description: "Pin readback mismatch",
				Error:       err,
				Timestamp:   time.Now(),
			})
			return fmt.Errorf("pin %s state mismatch", pinId)
		}

		m.recordResult(TestResult{
			Type:        TestGPIO,
			Component:   pinId,
			Status:      StatusPass,
			Description: "GPIO pin functional",
			Timestamp:   time.Now(),
		})
	}

	return nil
}

// TestPower performs power subsystem diagnostics
func (m *Manager) TestPower(ctx context.Context) error {
	if m.cfg.Power == nil {
		return fmt.Errorf("power manager not configured")
	}

	// Test power stability
	state := m.cfg.Power.GetState()
	if state.Voltage < m.cfg.MinVoltage {
		m.recordResult(TestResult{
			Type:        TestPower,
			Component:   "voltage",
			Status:      StatusFail,
			Description: "Voltage below minimum",
			Reading:     state.Voltage,
			Expected:    m.cfg.MinVoltage,
			Error:       fmt.Errorf("voltage %v below minimum %v", state.Voltage, m.cfg.MinVoltage),
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("voltage %v below minimum %v", state.Voltage, m.cfg.MinVoltage)
	}

	m.recordResult(TestResult{
		Type:        TestPower,
		Component:   "power_system",
		Status:      StatusPass,
		Description: "Power system functional",
		Reading:     state.Voltage,
		Expected:    m.cfg.MinVoltage,
		Timestamp:   time.Now(),
	})

	return nil
}

// TestThermal performs thermal subsystem diagnostics
func (m *Manager) TestThermal(ctx context.Context) error {
	if m.cfg.Thermal == nil {
		return fmt.Errorf("thermal monitor not configured")
	}

	state := m.cfg.Thermal.GetState()

	// Verify temperature readings
	if state.CPUTemp < m.cfg.TempRange[0] || state.CPUTemp > m.cfg.TempRange[1] {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "cpu_temp",
			Status:      StatusFail,
			Description: "CPU temperature out of range",
			Reading:     state.CPUTemp,
			Expected:    (m.cfg.TempRange[0] + m.cfg.TempRange[1]) / 2,
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("CPU temp %v outside range %v-%v", state.CPUTemp,
			m.cfg.TempRange[0], m.cfg.TempRange[1])
	}

	if state.GPUTemp < m.cfg.TempRange[0] || state.GPUTemp > m.cfg.TempRange[1] {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "gpu_temp",
			Status:      StatusFail,
			Description: "GPU temperature out of range",
			Reading:     state.GPUTemp,
			Expected:    (m.cfg.TempRange[0] + m.cfg.TempRange[1]) / 2,
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("GPU temp %v outside range %v-%v", state.GPUTemp,
			m.cfg.TempRange[0], m.cfg.TempRange[1])
	}

	// Set fan to low speed for test
	if err := m.cfg.Thermal.SetFanSpeed(25); err != nil {
		m.recordResult(TestResult{
			Type:        TestThermal,
			Component:   "fan",
			Status:      StatusFail,
			Description: "Failed to control fan speed",
			Reading:     25.0,
			Expected:    25.0,
			Error:       err,
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("failed to control fan: %w", err)
	}

	m.recordResult(TestResult{
		Type:        TestThermal,
		Component:   "thermal_system",
		Status:      StatusPass,
		Description: "Thermal system functional",
		Reading:     state.CPUTemp,
		Expected:    (m.cfg.TempRange[0] + m.cfg.TempRange[1]) / 2,
		Timestamp:   time.Now(),
	})

	return nil
}

// TestSecurity performs security subsystem diagnostics
func (m *Manager) TestSecurity(ctx context.Context) error {
	if m.cfg.Security == nil {
		return fmt.Errorf("security manager not configured")
	}

	state := m.cfg.Security.GetState()

	// Verify security sensors respond
	if state.CaseOpen {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "case_sensor",
			Status:      StatusWarning,
			Description: "Case open detected",
			Reading:     1.0,
			Expected:    0.0,
			Timestamp:   time.Now(),
		})
	}

	if state.MotionDetected {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "motion_sensor",
			Status:      StatusWarning,
			Description: "Motion detected",
			Reading:     1.0,
			Expected:    0.0,
			Timestamp:   time.Now(),
		})
	}

	if !state.VoltageNormal {
		m.recordResult(TestResult{
			Type:        TestSecurity,
			Component:   "voltage_monitor",
			Status:      StatusFail,
			Description: "Abnormal voltage detected",
			Reading:     0.0,
			Expected:    1.0,
			Timestamp:   time.Now(),
		})
		return fmt.Errorf("security voltage monitor shows abnormal state")
	}

	m.recordResult(TestResult{
		Type:        TestSecurity,
		Component:   "security_system", 
		Status:      StatusPass,
		Description: "Security system functional",
		Reading:     1.0,
		Expected:    1.0,
		Timestamp:   time.Now(),
	})

	return nil
}

// RunAll performs a complete hardware diagnostic suite
func (m *Manager) RunAll(ctx context.Context) error {
	tests := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"GPIO", m.TestGPIO},
		{"Power", m.TestPower}, 
		{"Thermal", m.TestThermal},
		{"Security", m.TestSecurity},
	}

	for _, test := range tests {
		for retry := 0; retry < m.cfg.Retries; retry++ {
			err := test.fn(ctx)
			if err == nil {
				break
			}
			if retry == m.cfg.Retries-1 {
				return fmt.Errorf("%s tests failed after %d retries: %w",
					test.name, m.cfg.Retries, err)
			}
			time.Sleep(time.Second) // Wait between retries
		}
	}

	return nil
}

// GetResults returns all test results
func (m *Manager) GetResults() []TestResult {
	m.mux.RLock()
	defer m.mux.RUnlock()

	results := make([]TestResult, len(m.results))
	copy(results, m.results)
	return results
}

// recordResult stores a test result and notifies callback if configured
func (m *Manager) recordResult(result TestResult) {
	m.mux.Lock()
	m.results = append(m.results, result)
	m.mux.Unlock()

	if m.cfg.OnTestComplete != nil {
		m.cfg.OnTestComplete(result)
	}
}
