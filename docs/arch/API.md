# Wrale Fleet API Architecture

This document describes the API contracts and interface specifications across all system layers.

## Layer APIs

### Metal Layer API

#### Device Management
```go
type DeviceManager interface {
    GetDevice(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
    ListDevices(ctx context.Context) ([]DeviceState, error)
    GetDevicesInZone(ctx context.Context, zone string) ([]DeviceState, error)
}
```

#### Hardware Management
- GPIO control and monitoring
- Power state management
- Thermal control interfaces
- Security monitoring

### Fleet Layer API

#### State Management
```go
type StateManager interface {
    GetDeviceState(ctx context.Context, deviceID DeviceID) (*DeviceState, error)
    UpdateDeviceState(ctx context.Context, state DeviceState) error
    ListDevices(ctx context.Context) ([]DeviceState, error)
    RemoveDevice(ctx context.Context, deviceID DeviceID) error
    AddDevice(ctx context.Context, state DeviceState) error
}
```

#### Edge Management
- Device state synchronization
- Local resource management
- Command processing
- Event handling

### Sync Layer API

#### State Store
```go
type StateStore interface {
    GetState(version StateVersion) (*VersionedState, error)
    SetState(deviceID DeviceID, state types.DeviceState) error
    SaveState(state *VersionedState) error
    ListVersions() ([]StateVersion, error)
    GetHistory(limit int) ([]StateChange, error)
    GetVersion() StateVersion
}
```

#### Sync Management
- Configuration distribution
- State synchronization
- Conflict resolution
- Version control

### User Layer API

#### Device Service
```go
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
interface Device {
    id: string
    status: string
    location: Location
    metrics: DeviceMetrics
    config: DeviceConfig
    lastUpdate: string
}

interface DeviceMetrics {
    temperature: number
    powerUsage: number
    cpuLoad: number
    memoryUsage: number
}
```

## Cross-Layer Communication

### Metal → Fleet
1. Hardware state updates
2. Physical metrics
3. Security events
4. Environmental data

### Fleet → Sync
1. Device state changes
2. Configuration updates
3. Operation events
4. Resource allocation

### Sync → User
1. State updates
2. Configuration changes
3. Event notifications
4. Resource metrics

## Real-Time APIs

### WebSocket Events
1. State changes
2. Metric updates
3. Alert notifications
4. Command responses

### Event Streams
1. Hardware events
2. State transitions
3. Metric streams
4. Alert streams

## API Security

### Authentication
1. Service-to-service auth
2. User authentication
3. Device authentication
4. Token management

### Authorization
1. Role-based access control
2. Resource permissions
3. Operation validation
4. Audit logging

## Error Handling

### Error Types
1. Hardware errors
2. State errors
3. Operation errors
4. Validation errors

### Error Propagation
1. Error context addition
2. Layer-specific wrapping
3. Client-friendly messages
4. Error recovery hints

## API Versioning

### Version Management
1. API versioning strategy
2. Compatibility guarantees
3. Deprecation policies
4. Migration support

### Backward Compatibility
1. Type compatibility
2. Interface stability
3. Default behaviors
4. Optional fields