package diag

import (
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/gpio"
	"github.com/wrale/wrale-fleet/metal/hw/power"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
	"github.com/wrale/wrale-fleet/metal/hw/secure"
)

// TestType identifies the type of diagnostic test
type TestType int

const (
	TestGPIO TestType = iota
	TestPower
	TestThermal
	TestSecurity
)

// TestStatus represents the result status of a test
type TestStatus int

const (
	StatusPass TestStatus = iota
	StatusFail
	StatusWarning
)

// TestResult captures the outcome of a diagnostic test
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

// Config defines the configuration for the diagnostics manager
type Config struct {
	GPIO          gpio.Controller
	Power         *power.Manager
	Thermal       *thermal.Monitor
	Security      *secure.Manager
	GPIOPins      map[string]int
	RetryAttempts int
	LoadTestTime  time.Duration
	MinVoltage    float64
	TempRange     [2]float64
	OnTestComplete func(TestResult)
}
