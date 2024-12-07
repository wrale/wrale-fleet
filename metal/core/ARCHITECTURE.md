# Core Layer Architecture

The core layer (`metal/core/`) serves as the central coordination point for the metal service, mediating between hardware control and fleet management while enforcing physical safety constraints and policies.

## Core Components

### Server (metald)
The main daemon process coordinating all metal operations.

- **HTTP Server** (`server/server.go`)
  - RESTful API endpoints
  - WebSocket connections for real-time updates
  - Health checks
  - Metrics exposure

- **Request Handling** (`server/handlers.go`)
  - Input validation
  - Command routing
  - Response formatting
  - Error handling

### Hardware Integration

- **Hardware Manager**
  - Subsystem coordination
  - Hardware state tracking
  - Command validation
  - Safety interlocks

- **Subsystem Controllers**
  - Power management interface
  - Thermal control interface
  - Security system interface
  - GPIO management

### State Management

- **Device State**
  - Current hardware status
  - Operational parameters
  - Environmental conditions
  - Alert conditions

- **State Synchronization**
  - State updates to fleet layer
  - Local state persistence
  - State recovery procedures
  - Consistency checks

### Policy Engine

- **Policy Manager**
  - Physical constraint enforcement
  - Operational rules
  - Safety policies
  - Resource limits

- **Rule Engine**
  - Policy evaluation
  - Action validation
  - Condition monitoring
  - Violation handling

### Security Layer

- **Authentication & Authorization**
  - API authentication
  - Permission management
  - Token validation
  - Session management

- **Audit Trail**
  - Operation logging
  - Access logging
  - Security events
  - Compliance tracking

### Event System

- **Event Bus**
  - Hardware events
  - State changes
  - Policy triggers
  - Alert conditions

- **Event Handlers**
  - Event processing
  - Action triggers
  - Notification dispatch
  - Event persistence

## Integration Patterns

### Hardware Layer Integration
1. Direct hardware control through metal/hw
2. Hardware state monitoring
3. Physical safety enforcement
4. Real-time metrics collection

### Fleet Layer Integration
1. Device state reporting
2. Command acceptance
3. Policy synchronization
4. Event propagation

### User API Integration
1. Command validation
2. State exposure
3. Event notification
4. Metric reporting

## Physical Considerations

### Safety Critical Operations
1. Pre-operation validation
2. Physical constraint checking
3. Operation monitoring
4. Emergency shutdown capabilities

### Environmental Awareness
1. Temperature monitoring
2. Power conditions
3. Physical security
4. Environmental policy enforcement

## State Management

### State Transitions
1. Safe state transitions
2. State persistence
3. Recovery procedures
4. Consistency maintenance

### State Synchronization
1. Fleet synchronization
2. Local caching
3. Conflict resolution
4. State recovery

## Error Handling

### Hardware Errors
1. Error detection
2. Safe state maintenance
3. Recovery procedures
4. Error reporting

### System Errors
1. Process monitoring
2. Resource monitoring
3. Service recovery
4. Error escalation

## Testing Considerations

### Component Testing
1. Server testing
2. Policy testing
3. State management testing
4. Integration testing

### System Testing
1. End-to-end testing
2. Load testing
3. Failure testing
4. Recovery testing

## Future Considerations

1. Enhanced policy engine
2. Advanced state management
3. Improved security features
4. Extended monitoring capabilities