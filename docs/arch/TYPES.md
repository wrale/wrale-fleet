# Wrale Fleet Type System Design

This document outlines the type system design principles and inheritance patterns across the Wrale Fleet architecture layers.

## Type System Principles

### 1. Layer Boundaries
- Each architectural layer (Metal, Fleet, Sync, User) maintains its own type definitions
- Types are transformed at layer boundaries
- No direct type sharing across non-adjacent layers
- Clear interface contracts between layers
- Shared types available through shared package

### 2. Type Hierarchy

```
User Layer (Presentation)
    ↑
    ├── TypeScript Types (UI)
    ├── API Types (REST)
    ├── View Models
    |
Sync Layer (Consensus)
    ↑
    ├── Versioned State
    ├── Sync Operations
    ├── Config Management
    |
Fleet Layer (Coordination)
    ↑
    ├── Device Models
    ├── Fleet State
    ├── Task Types
    |
Metal Layer (Hardware)
    ↑
    ├── Hardware Types
    ├── Diagnostic Types
    └── System State
```

### 3. Cross-Layer Communication

#### Metal → Fleet
- Hardware states are abstracted into device models
- Raw measurements are converted to typed metrics
- Events are transformed into structured notifications
- Errors are wrapped with context

#### Fleet → Sync
- Device states are versioned
- Changes are tracked with metadata
- Conflicts are detected and resolved
- Configuration is distributed

#### Sync → User
- Versioned states are transformed to view models
- Changes are propagated via WebSocket
- Events are enriched with sync status
- Configuration is validated

### 4. Type Safety

#### Validation Layers
1. Hardware Input Validation (Metal)
   - Range checking
   - Signal validation
   - Hardware constraints

2. Business Logic Validation (Fleet)
   - State machine rules
   - Fleet constraints
   - Operational limits

3. Sync Validation
   - State versioning
   - Conflict detection
   - Configuration validation

4. User Input Validation
   - Schema validation
   - Permission checks
   - User constraints

### 5. Interface Contracts

#### Hardware Interface (Metal)
```go
// From metal/types/types.go
type DeviceManager interface {
    GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
    ListDevices(ctx context.Context) ([]DeviceState, error)
    GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
}
```

#### Fleet Interface
```go
// From fleet/types/types.go
type StateManager interface {
    GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
    UpdateDeviceState(ctx context.Context, state DeviceState) error
    ListDevices(ctx context.Context) ([]DeviceState, error)
    RemoveDevice(ctx context.Context, deviceID DeviceID) error
    AddDevice(ctx context.Context, state DeviceState) error
}
```

#### Sync Interface
```go
// From sync/types/types.go
type StateStore interface {
    GetState(version StateVersion) (*VersionedState, error)
    SetState(deviceID DeviceID, state types.DeviceState) error
    SaveState(state *VersionedState) error
    ListVersions() ([]StateVersion, error)
    GetHistory(limit int) ([]StateChange, error)
    GetVersion() StateVersion
}
```

#### User Interface (API)
```go
// From user/api/types/types.go
type DeviceService interface {
    List(ctx context.Context) ([]Device, error)
    Get(ctx context.Context, id string) (*Device, error)
    Create(ctx context.Context, device *Device) error
    Update(ctx context.Context, device *Device) error
    Delete(ctx context.Context, id string) error
    SendCommand(ctx context.Context, id string, cmd *DeviceCommand) error
}
```

#### UI Types (TypeScript)
```typescript
// From user/ui/wrale-dashboard/src/types/device.ts
export interface Device {
    id: string
    status: string
    location: Location
    metrics: DeviceMetrics
    config: DeviceConfig
    lastUpdate: string
}

export interface DeviceMetrics {
    temperature: number
    powerUsage: number
    cpuLoad: number
    memoryUsage: number
}
```

## Type Translation

### 1. Upward Translation (→)
- Add context and metadata
- Enrich with related information
- Format for consumption
- Aggregate related data

### 2. Downward Translation (←)
- Strip to essential data
- Validate against constraints
- Transform to target format
- Apply security boundaries

### 3. Shared Types
```go
// From shared/types/types.go
type DeviceID string

type NodeType string

const (
    NodeEdge    NodeType = "EDGE"
    NodeControl NodeType = "CONTROL"
    NodeSensor  NodeType = "SENSOR"
)

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
```

## Implementation Guidelines

### 1. Type Definition Location
- Hardware types in metal/types package
- Fleet types in fleet/types package
- Sync types in sync/types package
- API types in user/api/types package
- UI types in user/ui/wrale-dashboard/src/types
- Shared types in shared/types package

### 2. Type Conversion
- Use explicit conversion functions
- Maintain conversion in single direction
- Validate during conversion
- Log conversion errors

### 3. Error Handling
- Wrap errors at layer boundaries
- Add context during propagation
- Maintain error type hierarchy
- Provide recovery mechanisms

### 4. Testing
- Test type conversions
- Verify boundary conditions
- Ensure type safety
- Validate contracts

## Evolution Guidelines

### 1. Type Evolution
- Version types at boundaries
- Maintain backward compatibility
- Plan for schema migrations
- Document type changes

### 2. Performance
- Consider serialization costs
- Optimize common paths
- Cache converted types
- Batch conversions

### 3. Security
- Validate at boundaries
- Sanitize user input
- Enforce type constraints
- Audit type usage