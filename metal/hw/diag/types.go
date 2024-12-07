package diag

import "time"

// TestType represents the type of diagnostic test
type TestType string

const (
	// TestGPIO represents GPIO diagnostic test
	TestGPIO TestType = "gpio"
	// TestPower represents power diagnostic test
	TestPower TestType = "power"
	// TestThermal represents thermal diagnostic test
	TestThermal TestType = "thermal"
)

// TestStatus represents the status of a test result
type TestStatus string

const (
	// StatusPass indicates test passed
	StatusPass TestStatus = "pass"
	// StatusFail indicates test failed
	StatusFail TestStatus = "fail"
	// StatusSkipped indicates test was skipped
	StatusSkipped TestStatus = "skipped"
)

// Config represents diagnostic configuration
type Config struct {
	EnabledTests  []TestType          `json:"enabled_tests"`
	Thresholds    map[string]float64  `json:"thresholds,omitempty"`
	RetryAttempts int                 `json:"retry_attempts"`
	TestInterval  time.Duration       `json:"test_interval"`
	GPIO          string             `json:"gpio"`
	GPIOPins      []int              `json:"gpio_pins"`
	Retries       int                `json:"retries"`
	LoadTestTime  time.Duration      `json:"load_test_time"`
	MinVoltage    float64            `json:"min_voltage"`
	TempRange     [2]float64         `json:"temp_range"`
}

// TestResult represents the result of a diagnostic test
type TestResult struct {
	Type          TestType            `json:"type"`
	Status        TestStatus          `json:"status"`
	Timestamp     int64               `json:"timestamp"`
	Message       string              `json:"message,omitempty"`
	Measurements  map[string]float64  `json:"measurements,omitempty"`
}