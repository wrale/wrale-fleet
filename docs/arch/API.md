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
```go
type HardwareManager interface {
    // GPIO management
    GetGPIOState(pin GPIOPin) (GPIOState, error)
    SetGPIOState(pin GPIOPin, state GPIOState) error
    MonitorGPIO(pin GPIOPin) (<-chan GPIOEvent, error)

    // Power management
    GetPowerState() (PowerState, error)
    SetPowerMode(mode PowerMode) error
    MonitorPower() (<-chan PowerEvent, error)

    // Thermal management
    GetTemperature() (Temperature, error)
    SetCoolingMode(mode CoolingMode) error
    MonitorTemperature() (<-chan TempEvent, error)
}
```

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

#### Configuration Management
```go
type ConfigManager interface {
    // Layer configuration
    GetLayerConfig(layer Layer) (*LayerConfig, error)
    UpdateLayerConfig(layer Layer, config *LayerConfig) error
    ValidateConfig(config *LayerConfig) error

    // Device configuration
    GetDeviceConfig(deviceID DeviceID) (*DeviceConfig, error)
    UpdateDeviceConfig(deviceID DeviceID, config *DeviceConfig) error
    ValidateDeviceConfig(config *DeviceConfig) error

    // Policy configuration
    GetPolicyConfig() (*PolicyConfig, error)
    UpdatePolicyConfig(config *PolicyConfig) error
    ValidatePolicyConfig(config *PolicyConfig) error

    // Config versioning
    GetConfigVersion() (Version, error)
    RollbackConfig(version Version) error
    ListConfigVersions() ([]Version, error)
}

// Configuration types
type LayerConfig struct {
    Layer       Layer                    `json:"layer"`
    Settings    map[string]interface{}   `json:"settings"`
    Constraints map[string]Constraint    `json:"constraints"`
    Version     Version                  `json:"version"`
}

type DeviceConfig struct {
    DeviceID    DeviceID                 `json:"device_id"`
    Hardware    HardwareConfig           `json:"hardware"`
    Network     NetworkConfig            `json:"network"`
    Security    SecurityConfig           `json:"security"`
    Version     Version                  `json:"version"`
}

type PolicyConfig struct {
    Safety      SafetyPolicy             `json:"safety"`
    Resource    ResourcePolicy           `json:"resource"`
    Security    SecurityPolicy           `json:"security"`
    Version     Version                  `json:"version"`
}
```

#### Recovery Management
```go
type RecoveryManager interface {
    // State recovery
    DetectStateInconsistency() error
    InitiateStateRecovery(deviceID DeviceID) error
    ValidateStateRecovery(deviceID DeviceID) error
    CompleteStateRecovery(deviceID DeviceID) error

    // Config recovery
    DetectConfigInconsistency() error
    InitiateConfigRecovery(version Version) error
    ValidateConfigRecovery(version Version) error
    CompleteConfigRecovery(version Version) error

    // System recovery
    DetectSystemIssue() error
    InitiateSystemRecovery(component Component) error
    ValidateSystemRecovery(component Component) error
    CompleteSystemRecovery(component Component) error

    // Recovery monitoring
    GetRecoveryStatus(recoveryID RecoveryID) (*RecoveryStatus, error)
    MonitorRecovery(recoveryID RecoveryID) (<-chan RecoveryEvent, error)
    ListActiveRecoveries() ([]RecoveryStatus, error)
}

// Recovery types
type RecoveryStatus struct {
    ID          RecoveryID               `json:"recovery_id"`
    Type        RecoveryType             `json:"type"`
    State       RecoveryState            `json:"state"`
    Progress    float64                  `json:"progress"`
    StartTime   time.Time                `json:"start_time"`
    LastUpdate  time.Time                `json:"last_update"`
    Error       string                   `json:"error,omitempty"`
}

type RecoveryEvent struct {
    ID          RecoveryID               `json:"recovery_id"`
    Type        RecoveryEventType        `json:"type"`
    Message     string                   `json:"message"`
    Timestamp   time.Time                `json:"timestamp"`
    Data        map[string]interface{}   `json:"data,omitempty"`
}
```

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

#### Conflict Resolution
```go
type ConflictResolver interface {
    // Conflict detection
    DetectConflicts(states []VersionedState) ([]Conflict, error)
    ValidateStateConsistency(state VersionedState) error
    
    // Resolution
    ResolveConflict(conflict Conflict) (*Resolution, error)
    ApplyResolution(resolution *Resolution) error
    ValidateResolution(resolution *Resolution) error

    // History
    GetConflictHistory(deviceID DeviceID) ([]Conflict, error)
    GetResolutionHistory(deviceID DeviceID) ([]Resolution, error)
}
```

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

#### Real-time Events
```go
type EventService interface {
    // Event streams
    SubscribeDeviceEvents(deviceID string) (<-chan DeviceEvent, error)
    SubscribeStateEvents() (<-chan StateEvent, error)
    SubscribeAlertEvents() (<-chan AlertEvent, error)
    
    // Event management
    PublishEvent(event Event) error
    GetEventHistory(deviceID string) ([]Event, error)
    AcknowledgeEvent(eventID string) error
}
```

## Error Handling

### Error Types
```go
type Error interface {
    error
    Code() ErrorCode
    Layer() Layer
    Details() map[string]interface{}
    Recoverable() bool
}

// Error implementations
type HardwareError struct {
    code        ErrorCode
    message     string
    component   Component
    recoverable bool
    details     map[string]interface{}
}

type StateError struct {
    code        ErrorCode
    message     string
    stateID     StateID
    version     Version
    recoverable bool
    details     map[string]interface{}
}

type OperationError struct {
    code        ErrorCode
    message     string
    operation   Operation
    recoverable bool
    details     map[string]interface{}
}
```

### Error Recovery
```go
type ErrorRecovery interface {
    // Recovery handling
    CanRecover(err error) bool
    GetRecoveryPlan(err error) (*RecoveryPlan, error)
    ExecuteRecovery(plan *RecoveryPlan) error
    ValidateRecovery(plan *RecoveryPlan) error

    // Recovery monitoring
    GetRecoveryStatus(planID string) (*RecoveryStatus, error)
    MonitorRecovery(planID string) (<-chan RecoveryEvent, error)
}
```

## Version Management

### API Versioning
```go
type VersionManager interface {
    // Version control
    GetAPIVersion() Version
    ValidateVersion(version Version) error
    IsCompatible(clientVersion Version) bool

    // Migration
    RequiresMigration(fromVersion Version) bool
    GetMigrationPath(fromVersion Version) ([]Migration, error)
    ExecuteMigration(migration Migration) error
}
```

### Compatibility Checking
```go
type CompatibilityChecker interface {
    // Interface compatibility
    ValidateInterface(interface{} interface{}) error
    CheckCompatibility(oldVersion, newVersion Version) error
    
    // Type compatibility
    ValidateType(typeName string, version Version) error
    GetCompatibleTypes(version Version) ([]string, error)
}
```