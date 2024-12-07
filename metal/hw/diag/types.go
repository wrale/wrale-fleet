package diag

import (
	"time"
	
	"github.com/wrale/wrale-fleet/metal/hw/gpio"
	"github.com/wrale/wrale-fleet/metal/hw/power"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
	"github.com/wrale/wrale-fleet/metal/hw/secure"
)

// TestType represents the type of diagnostic test
type TestType string

const (
	// TestGPIO represents GPIO diagnostic test
	TestGPIO TestType = "gpio"
	// TestPower represents power diagnostic test 
	TestPower TestType = "power"
	// TestThermal represents thermal diagnostic test
	TestThermal TestType = "thermal"
	// TestSecurity represents security diagnostic test
	TestSecurity TestType = "security"
)

// TestStatus represents the status of a test result
type TestStatus string

const (
	// StatusPass indicates test passed
	StatusPass TestStatus = "pass"
	// StatusFail indicates test failed 
	StatusFail TestStatus = "fail"
	// StatusWarning indicates test warning
	StatusWarning TestStatus = "warning"
	// StatusSkipped indicates test was skipped
	StatusSkipped TestStatus = "skipped"
)

// Config represents diagnostic configuration
type Config struct {
	// Hardware controllers
	GPIO         *gpio.Controller       // GPIO controller
	GPIOPins     map[string]int        // Map of pin names to numbers
	Power        power.Manager         // Power subsystem manager
	Thermal      thermal.Monitor       // Thermal subsystem monitor
	Security     secure.Manager        // Security subsystem manager

	// Test configuration
	EnabledTests  []TestType           // List of enabled test types
	Thresholds    map[string]float64   // Test-specific thresholds
	RetryAttempts int                  // Number of test retries
	TestInterval  time.Duration        // Interval between tests

	// Hardware specific settings
	LoadTestTime  time.Duration        // Duration for load tests
	MinVoltage    float64             // Minimum acceptable voltage
	TempRange     [2]float64          // Acceptable temperature range [min, max]

	// Callbacks
	OnTestComplete func(TestResult)    // Callback for test completion
}

// TestResult represents the result of a diagnostic test
type TestResult struct {
	Type          TestType            // Type of test performed
	Status        TestStatus          // Test result status
	Timestamp     time.Time           // When the test was performed
	Component     string              // Component being tested
	Description   string              // Test description
	Reading       float64             // Actual reading (if applicable)
	Expected      float64             // Expected value (if applicable)
	Error         error               // Error details if test failed
	Measurements  map[string]float64  // Additional measurements
}