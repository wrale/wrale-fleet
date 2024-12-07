package metal

import "time"

// CommonState contains fields shared by all hardware states
type CommonState struct {
    DeviceID  string    `json:"device_id"`  // Device identifier
    UpdatedAt time.Time `json:"updated_at"` // Last update timestamp
}

// TestStatus indicates test result status
type TestStatus string

const (
    StatusPass    TestStatus = "PASS"
    StatusFail    TestStatus = "FAIL"
    StatusSkipped TestStatus = "SKIPPED"
)

// TestType identifies diagnostic test type
type TestType string

const (
    TestGPIO     TestType = "GPIO"
    TestPower    TestType = "POWER"
    TestThermal  TestType = "THERMAL"
    TestSecurity TestType = "SECURITY"
)

// TestResult holds test execution results
type TestResult struct {
    Type        TestType    `json:"type"`        // Test type
    Component   string      `json:"component"`   // Component tested
    Status      TestStatus  `json:"status"`      // Test status
    Reading     float64     `json:"reading"`     // Measured value
    Expected    float64     `json:"expected"`    // Expected value
    Description string      `json:"description"` // Test description
    Error       string      `json:"error"`       // Error if any
    Timestamp   time.Time   `json:"timestamp"`   // Test time
}

// PowerSource identifies power sources
type PowerSource string 

const (
    MainPower    PowerSource = "MAIN"
    BatteryPower PowerSource = "BATTERY"
    SolarPower   PowerSource = "SOLAR"
)

// PowerState represents power subsystem state
type PowerState struct {
    CommonState
    BatteryLevel     float64                `json:"battery_level"`
    Charging         bool                   `json:"charging"`
    Voltage          float64                `json:"voltage"`
    CurrentSource    PowerSource            `json:"current_source"`
    AvailablePower   map[PowerSource]bool   `json:"available_power"`
    PowerConsumption float64                `json:"power_consumption"`
}
