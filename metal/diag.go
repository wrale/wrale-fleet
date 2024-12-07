package metal

import (
	"context"
	"time"

	"github.com/wrale/wrale-fleet/metal/internal/types"
)

// TestType identifies the type of diagnostic test
type TestType string

const (
	TestGPIO     TestType = "GPIO"
	TestPower    TestType = "POWER"
	TestThermal  TestType = "THERMAL"
	TestSecurity TestType = "SECURITY"
)

// TestStatus represents the outcome of a diagnostic test
type TestStatus string

const (
	StatusPass    TestStatus = "PASS"
	StatusFail    TestStatus = "FAIL"
	StatusWarning TestStatus = "WARNING"
	StatusSkipped TestStatus = "SKIPPED"
)

// TestResult contains the outcome of a diagnostic test
type TestResult struct {
	types.CommonState
	Type        TestType    `json:"type"`
	Component   string      `json:"component"`
	Status      TestStatus  `json:"status"`
	Reading     float64     `json:"reading,omitempty"`
	Expected    float64     `json:"expected,omitempty"`
	Duration    time.Duration `json:"duration"`
	Description string      `json:"description"`
	Error       string      `json:"error,omitempty"`
}

// DiagnosticManager defines the interface for hardware diagnostics
type DiagnosticManager interface {
	types.Monitor

	// Individual Tests
	TestGPIO(ctx context.Context) error
	TestPower(ctx context.Context) error
	TestThermal(ctx context.Context) error
	TestSecurity(ctx context.Context) error
	
	// Test Suites
	RunAll(ctx context.Context) error
	RunSelected(ctx context.Context, types []TestType) error
	
	// Test Management
	AbortTests(ctx context.Context) error
	GetTestStatus(testID string) (*TestResult, error)
	ListTestResults() ([]TestResult, error)
	
	// Component Validation
	ValidateComponent(ctx context.Context, component string) error
	CalibrateComponent(ctx context.Context, component string) error
	
	// Resource Monitoring
	GetResourceUsage(component string) (map[string]float64, error)
	MonitorResources(ctx context.Context) (<-chan map[string]float64, error)
	
	// Events
	OnTestComplete(func(TestResult))
	OnTestStart(func(TestType, string))
}

// DiagnosticManagerConfig holds configuration for diagnostics
type DiagnosticManagerConfig struct {
	GPIO            types.GPIOController
	PowerManager    PowerManager
	ThermalManager  ThermalManager
	SecurityManager SecurityManager
	RetryAttempts   int
	LoadTestTime    time.Duration
	MinVoltage      float64
	TempRange       [2]float64
	OnTestComplete  func(TestResult)
	OnTestStart     func(TestType, string)
}

// DiagnosticOptions holds optional test configuration
type DiagnosticOptions struct {
	Retries     int
	Timeout     time.Duration
	LoadTest    bool
	SkipTests   []TestType
	Components  []string
	Thresholds  map[string]float64
}

// Common test-related functions
var (
	// WithRetries sets the retry count for tests
	WithRetries = func(retries int) Option {
		return func(v interface{}) error {
			if retries < 0 {
				return &ValidationError{Field: "retries", Value: retries, Err: ErrInvalidConfig}
			}
			if d, ok := v.(interface{ setRetries(int) }); ok {
				d.setRetries(retries)
				return nil
			}
			return ErrNotSupported
		}
	}

	// WithTimeout sets the test timeout duration
	WithTimeout = func(timeout time.Duration) Option {
		return func(v interface{}) error {
			if timeout <= 0 {
				return &ValidationError{Field: "timeout", Value: timeout, Err: ErrInvalidConfig}
			}
			if d, ok := v.(interface{ setTimeout(time.Duration) }); ok {
				d.setTimeout(timeout)
				return nil
			}
			return ErrNotSupported
		}
	}

	// WithLoadTest enables extended load testing
	WithLoadTest = func(enable bool) Option {
		return func(v interface{}) error {
			if d, ok := v.(interface{ setLoadTest(bool) }); ok {
				d.setLoadTest(enable)
				return nil
			}
			return ErrNotSupported
		}
	}
)
