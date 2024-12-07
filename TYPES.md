# Wrale Fleet Type System Design

This document outlines the type system design principles and inheritance patterns across the Wrale Fleet architecture layers.

## Type System Principles

### 1. Layer Boundaries
- Each architectural layer (Metal, Fleet, User) maintains its own type definitions
- Types are transformed at layer boundaries
- No direct type sharing across non-adjacent layers
- Clear interface contracts between layers

### 2. Type Hierarchy

```
User Layer (Presentation)
    ↑
    ├── View Models
    ├── API DTOs
    ├── UI State Types
    |
Fleet Layer (Coordination)
    ↑
    ├── Device Models
    ├── Fleet State
    ├── Orchestration Types
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

#### Fleet → User
- Device models are mapped to view models
- State is transformed for UI consumption
- Metrics are formatted for visualization
- Events are enriched with user context

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

3. User Input Validation (User)
   - Schema validation
   - Permission checks
   - User constraints

### 5. Interface Contracts

#### Hardware Interface
```go
// Example contract pattern
type HardwareController interface {
    Initialize() error
    GetState() State
    Configure(Config) error
    Monitor() <-chan Event
}
```

#### Fleet Interface
```go
// Example contract pattern
type FleetController interface {
    ManageDevice(DeviceID) error
    GetDeviceState(DeviceID) DeviceState
    UpdateConfiguration(DeviceConfig) error
    SubscribeEvents() <-chan FleetEvent
}
```

#### User Interface
```go
// Example contract pattern
type UserInterface interface {
    GetDevices() []DeviceViewModel
    UpdateDevice(DeviceUpdateDTO) error
    SubscribeUpdates() <-chan UIEvent
}
```

## Data Flow Patterns

### 1. State Propagation
```
Hardware State → Metal State → Fleet State → View State
     (raw)         (typed)      (enriched)    (formatted)
```

### 2. Command Flow
```
User Command → Fleet Command → Metal Command → Hardware Command
   (request)     (validated)    (translated)     (executed)
```

### 3. Event Flow
```
Hardware Event → Metal Event → Fleet Event → UI Event
    (signal)      (typed)       (enriched)    (displayed)
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

## Implementation Guidelines

### 1. Type Definition Location
- Hardware types in respective hw/* packages
- Fleet types in fleet/types package
- User types in ui/types and api/types

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

## Future Considerations

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