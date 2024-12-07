# Wrale Fleet Sync Architecture

## Overview

The sync layer provides state synchronization, configuration distribution, and conflict resolution across the Wrale Fleet system. It ensures consistency between the fleet brain, edge devices, and hardware state while handling network partitions and physical-world constraints.

## Core Components

### State Management

#### State Types
```go
type VersionedState struct {
    Version     StateVersion             `json:"version"`
    State       DeviceState              `json:"state"`
    Timestamp   time.Time                `json:"timestamp"`
    Source      StateSource              `json:"source"`
    Metadata    map[string]interface{}   `json:"metadata"`
}

type StateChange struct {
    Version     StateVersion             `json:"version"`
    Delta       StateDelta               `json:"delta"`
    Timestamp   time.Time                `json:"timestamp"`
    Source      StateSource              `json:"source"`
    Reason      ChangeReason             `json:"reason"`
}

type StateDelta struct {
    Added       map[string]interface{}   `json:"added"`
    Modified    map[string]interface{}   `json:"modified"`
    Removed     []string                 `json:"removed"`
}
```

#### State Store
```go
type StateStore interface {
    // State operations
    GetState(version StateVersion) (*VersionedState, error)
    SetState(deviceID DeviceID, state DeviceState) error
    SaveState(state *VersionedState) error
    
    // Version management
    ListVersions() ([]StateVersion, error)
    GetHistory(limit int) ([]StateChange, error)
    GetVersion() StateVersion
    
    // Snapshot management
    CreateSnapshot() (*StateSnapshot, error)
    RestoreSnapshot(snapshot *StateSnapshot) error
    ListSnapshots() ([]StateSnapshot, error)
}
```

### Recovery Framework

#### Recovery Types
```go
type RecoveryType string

const (
    RecoveryHardware    RecoveryType = "HARDWARE"
    RecoveryState       RecoveryType = "STATE"
    RecoveryConfig      RecoveryType = "CONFIG"
    RecoveryNetwork     RecoveryType = "NETWORK"
    RecoverySecurity    RecoveryType = "SECURITY"
)

type RecoveryEvent struct {
    ID           RecoveryID              `json:"id"`
    Type         RecoveryType            `json:"type"`
    Status       RecoveryStatus          `json:"status"`
    Component    ComponentID             `json:"component"`
    Timestamp    time.Time               `json:"timestamp"`
    Details      map[string]interface{}  `json:"details"`
}
```

#### Recovery Manager
```go
type RecoveryManager interface {
    // Generic recovery
    InitiateRecovery(event RecoveryEvent) error
    MonitorRecovery(recoveryID RecoveryID) (<-chan RecoveryEvent, error)
    CompleteRecovery(recoveryID RecoveryID) error
    
    // State recovery
    RecoverState(stateID StateID) error
    ValidateStateRecovery(stateID StateID) error
    
    // Network recovery
    HandlePartition(partition NetworkPartition) error
    ReconnectNode(nodeID NodeID) error
    
    // Hardware recovery
    RecoverDevice(deviceID DeviceID) error
    ValidateDeviceRecovery(deviceID DeviceID) error
}
```

### Conflict Resolution

#### Conflict Detection
```go
type ConflictDetector interface {
    // Detection
    DetectConflicts(states []VersionedState) ([]Conflict, error)
    ValidateStateConsistency(state VersionedState) error
    
    // Analysis
    AnalyzeConflict(conflict Conflict) (*ConflictAnalysis, error)
    GetResolutionStrategy(analysis *ConflictAnalysis) (ResolutionStrategy, error)
}

type Conflict struct {
    ID          ConflictID              `json:"id"`
    Type        ConflictType            `json:"type"`
    States      []VersionedState        `json:"states"`
    Timestamp   time.Time               `json:"timestamp"`
    Status      ConflictStatus          `json:"status"`
}
```

#### Resolution Engine
```go
type ConflictResolver interface {
    // Resolution
    ResolveConflict(conflict Conflict) (*Resolution, error)
    ApplyResolution(resolution Resolution) error
    ValidateResolution(resolution Resolution) error
    
    // History
    GetConflictHistory(deviceID DeviceID) ([]Conflict, error)
    GetResolutionHistory(deviceID DeviceID) ([]Resolution, error)
}

type Resolution struct {
    ConflictID  ConflictID              `json:"conflict_id"`
    Strategy    ResolutionStrategy      `json:"strategy"`
    Result      *VersionedState         `json:"result"`
    Timestamp   time.Time               `json:"timestamp"`
    Metadata    map[string]interface{}  `json:"metadata"`
}
```

### Configuration Distribution

#### Config Management
```go
type ConfigManager interface {
    // Distribution
    DistributeConfig(config DistributedConfig) error
    ValidateDistribution(config DistributedConfig) error
    RollbackDistribution(version ConfigVersion) error
    
    // Status
    GetDistributionStatus(version ConfigVersion) (*DistributionStatus, error)
    MonitorDistribution(version ConfigVersion) (<-chan DistributionEvent, error)
}

type DistributedConfig struct {
    Version     ConfigVersion           `json:"version"`
    Config      interface{}             `json:"config"`
    Target      ConfigTarget            `json:"target"`
    Policy      DistributionPolicy      `json:"policy"`
}

type ConfigTarget struct {
    Layers      []Layer                 `json:"layers"`
    Devices     []DeviceID              `json:"devices"`
    Groups      []GroupID               `json:"groups"`
}
```

## Physical Considerations

### Network Constraints
```go
type NetworkPolicy struct {
    // Bandwidth management
    MaxBandwidth    Bandwidth           `json:"max_bandwidth"`
    Priority        SyncPriority        `json:"priority"`
    Schedule        SyncSchedule        `json:"schedule"`
    
    // Resource constraints
    CPULimit        float64             `json:"cpu_limit"`
    MemoryLimit     ByteSize            `json:"memory_limit"`
    DiskLimit       ByteSize            `json:"disk_limit"`
    
    // Network constraints
    Latency         Duration            `json:"latency"`
    Reliability     float64             `json:"reliability"`
    QoS             QualityOfService    `json:"qos"`
}
```

### Physical Sync Patterns

#### State Sync
1. Rate limiting based on bandwidth
2. Priority-based synchronization
3. Delta compression for state updates
4. Resource-aware scheduling

#### Config Distribution
1. Staged rollouts for safety
2. Physical location awareness
3. Resource availability checks
4. Network condition adaptation

## Error Handling

### Sync Errors
```go
type SyncError struct {
    Code        ErrorCode              `json:"code"`
    Message     string                 `json:"message"`
    Source      ErrorSource            `json:"source"`
    Recoverable bool                   `json:"recoverable"`
}

type ErrorSource struct {
    Layer       Layer                  `json:"layer"`
    Component   string                 `json:"component"`
    Operation   string                 `json:"operation"`
}
```

### Error Recovery
```go
type ErrorHandler interface {
    // Error handling
    HandleError(err SyncError) error
    GetRecoveryAction(err SyncError) (*RecoveryAction, error)
    ExecuteRecovery(action RecoveryAction) error
    
    // Status tracking
    GetErrorStatus(errorID ErrorID) (*ErrorStatus, error)
    MonitorRecovery(errorID ErrorID) (<-chan RecoveryEvent, error)
}
```

## Health Monitoring

### Sync Health
```go
type SyncHealthCheck interface {
    // Health monitoring
    CheckSyncHealth() (*SyncHealthStatus, error)
    ValidateSyncStatus() error
    
    // Metrics
    GetSyncMetrics() (*SyncMetrics, error)
    MonitorSyncHealth() (<-chan HealthEvent, error)
}

type SyncMetrics struct {
    StateChanges     uint64             `json:"state_changes"`
    Conflicts        uint64             `json:"conflicts"`
    SyncLatency      Duration           `json:"sync_latency"`
    DataTransferred  ByteSize           `json:"data_transferred"`
}
```

## Best Practices

### Implementation Guidelines
1. Always version state changes
2. Handle network partitions gracefully
3. Implement conflict detection early
4. Maintain audit trails
5. Consider physical constraints
6. Prioritize safety over consistency

### Performance Guidelines
1. Batch small updates
2. Use delta compression
3. Implement backoff strategies
4. Monitor resource usage
5. Cache frequently accessed state
6. Optimize for physical constraints