package types

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/power"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
	"github.com/wrale/wrale-fleet/metal/hw/secure"
)

// GPIOController defines interface for GPIO operations
type GPIOController interface {
	Initialize() error
	SetMode(pin int, mode string) error
	Write(pin int, value bool) error
	Read(pin int) (bool, error)
}

// Config defines hardware diagnostics configuration
type Config struct {
	GPIO          GPIOController
	Power         power.Manager
	Thermal       thermal.Monitor
	Security      secure.Manager
	GPIOPins      map[string]int
	RetryAttempts int
	LoadTestTime  time.Duration
	MinVoltage    float64
	TempRange     [2]float64
	OnTestComplete func(TestResult)
}

// TestType identifies the type of diagnostic test
type TestType string

const (
	TestGPIO     TestType = "gpio"
	TestPower    TestType = "power"
	TestThermal  TestType = "thermal"
	TestSecurity TestType = "security"
)

// TestStatus represents the outcome of a diagnostic test
type TestStatus string

const (
	StatusPass    TestStatus = "pass"
	StatusFail    TestStatus = "fail"
	StatusWarning TestStatus = "warning"
)

// TestResult contains the outcome of a diagnostic test
type TestResult struct {
	Type        TestType
	Component   string
	Status      TestStatus
	Description string
	Reading     float64
	Expected    float64
	Error       error
	Timestamp   time.Time
}