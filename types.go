package wrale_fleet

import "time"

// DeviceID uniquely identifies a device in the fleet
type DeviceID string

// NodeType identifies the role of a device
type NodeType string

const (
    NodeEdge    NodeType = "EDGE"
    NodeControl NodeType = "CONTROL"
    NodeSensor  NodeType = "SENSOR"
)

// Device represents a physical device in the fleet
type Device struct {
    ID           DeviceID
    Type         NodeType
    Location     Location
    Capabilities []Capability
    Metadata     map[string]string
    LastSeen     time.Time
}

// Location represents physical location info
type Location struct {
    Latitude   float64
    Longitude  float64
    Altitude   float64
    Indoor     bool
    Zone       string
    UpdatedAt  time.Time
}

// Capability represents a device's hardware capabilities
type Capability string

const (
    CapGPIO      Capability = "GPIO"
    CapPWM       Capability = "PWM"
    CapI2C       Capability = "I2C"
    CapSPI       Capability = "SPI"
    CapAnalog    Capability = "ANALOG"
    CapMotion    Capability = "MOTION"
    CapThermal   Capability = "THERMAL"
    CapPower     Capability = "POWER"
    CapSecurity  Capability = "SECURITY"
)

// Event represents a system-wide event
type Event struct {
    Source    DeviceID
    Type      EventType
    Severity  EventSeverity
    Timestamp time.Time
    Payload   interface{}
}

// EventType identifies different types of events
type EventType string 

// EventSeverity indicates event importance
type EventSeverity string

const (
    SeverityInfo     EventSeverity = "INFO"
    SeverityWarning  EventSeverity = "WARNING"
    SeverityError    EventSeverity = "ERROR"
    SeverityCritical EventSeverity = "CRITICAL"
)

// State interface for component state management
type State interface {
    DeviceID() DeviceID
    LastUpdate() time.Time
    Validate() error
}
