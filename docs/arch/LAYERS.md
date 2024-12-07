# Wrale Fleet Layer Architecture

This document provides detailed architecture specifications for each system layer.

## Metal Layer

### Core Components
```
metal/
├── hw/              # Hardware abstraction
│   ├── gpio/       # GPIO management
│   ├── power/      # Power monitoring
│   ├── secure/     # Security monitoring
│   └── thermal/    # Temperature control
├── core/           # System coordination
│   ├── server/     # HTTP API
│   ├── secure/     # Security policies
│   └── thermal/    # Thermal policies
└── diag/           # Diagnostics
```

### Key Responsibilities
1. Hardware interaction
2. Physical safety monitoring
3. Environmental control
4. System diagnostics
5. Resource management

### Interface Contracts
```go
// From metal/types/types.go
type DeviceManager interface {
    GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
    ListDevices(ctx context.Context) ([]DeviceState, error)
    GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
}
```

## Fleet Layer

### Core Components
```
fleet/
├── cmd/           # Command line tools
│   └── fleetd/    # Fleet daemon
├── coordinator/   # Central coordination
│   ├── metal.go
│   ├── orchestrator.go
│   └── scheduler.go
├── device/       # Device management
│   ├── inventory.go
│   └── topology.go
├── edge/         # Edge device management
│   ├── agent/    # Edge agent
│   ├── client/   # Client implementations
│   └── store/    # Edge state storage
├── engine/       # Analysis engine
│   ├── analyzer.go
│   ├── optimizer.go
│   └── thermal.go
├── service/      # Fleet services
└── types/        # Fleet types
```

### Key Responsibilities
1. Device coordination
2. Resource management
3. Policy enforcement
4. Task scheduling
5. Edge management

### Interface Contracts
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

## Sync Layer

### Core Components
```
sync/
├── config/        # Sync configuration
│   ├── config.go
│   └── config_test.go
├── manager/       # Sync orchestration
│   ├── sync.go
│   └── manager_test.go
├── resolver/      # Conflict resolution
│   ├── resolver.go
│   └── resolver_test.go
├── store/        # State storage
│   ├── store.go
│   └── store_test.go
└── types/        # Sync types
```

### Key Responsibilities
1. State versioning
2. Configuration distribution
3. Conflict resolution
4. Consistency management
5. Update coordination

### Interface Contracts
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

## User Layer

### Core Components
```
user/
├── api/           # Backend API
│   ├── server/
│   ├── service/
│   └── types/
└── ui/            # Frontend
    └── wrale-dashboard/
        ├── components/
        ├── services/
        └── types/
```

### Key Responsibilities
1. User interface
2. API endpoints
3. Authentication
4. Real-time updates
5. Data visualization

### Interface Contracts

#### API Contracts
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

#### UI Types
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

## Shared Layer

### Core Components
```
shared/
├── config/        # Configuration
├── testing/       # Test utilities
└── types/         # Common types
```

### Key Responsibilities
1. Common utilities
2. Shared types
3. Testing tools
4. Configuration

### Shared Types
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

## Layer Communication

### Type Translation
1. **Upward Translation (↑)**
   - Add context/metadata
   - Enrich with related data
   - Format for consumption

2. **Downward Translation (↓)**
   - Strip to essential data
   - Validate constraints
   - Transform format

### Error Propagation
1. Detect at source
2. Add context
3. Transform for layer
4. Present appropriately

### State Flow
1. Hardware → Metal
2. Metal → Fleet
3. Fleet → Sync
4. Sync → User