# Wrale Fleet Layer Architecture

This document provides detailed architecture specifications for each system layer. For detailed API contracts and interfaces, see [API.md](API.md).

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
1. Hardware interaction and control
2. Physical safety monitoring
3. Environmental control
4. System diagnostics
5. Resource management

### Core Services
- Hardware abstraction layer
- Physical device management
- Real-time monitoring
- Safety enforcement
- Resource coordination

### Key Patterns
- Direct hardware access
- Real-time event processing
- Safety-first operations
- Resource management

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

### Core Services
- Fleet orchestration
- Device management
- Resource allocation
- State coordination
- Policy enforcement

### Key Patterns
- Distributed coordination
- State management
- Policy enforcement
- Resource optimization

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

### Core Services
- State synchronization
- Configuration management
- Version control
- Conflict resolution
- Distribution coordination

### Key Patterns
- Version control
- State replication
- Conflict detection
- Consistency checking

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
1. User interface presentation
2. API endpoint management
3. Authentication/Authorization
4. Real-time updates
5. Data visualization

### Core Services
- Web dashboard
- REST API
- WebSocket updates
- User management
- Data presentation

### Key Patterns
- Responsive design
- Real-time updates
- User authentication
- Data visualization

## Shared Layer

### Core Components
```
shared/
├── config/        # Configuration
├── testing/       # Test utilities
└── types/        # Common types
```

### Key Responsibilities
1. Common utilities
2. Shared types
3. Testing tools
4. Configuration

### Core Services
- Type definitions
- Utility functions
- Test frameworks
- Configuration management

### Key Patterns
- Type sharing
- Configuration management
- Test utilities
- Common functionality

## Layer Communication

### State Flow Patterns

#### State Transformation
1. **Hardware to Metal**
   - Raw data → Typed metrics
   - Signal data → Events
   - Physical state → Device state

2. **Metal to Fleet**
   - Device state → Fleet state
   - Events → Coordination
   - Metrics → Analysis

3. **Fleet to Sync**
   - Fleet state → Versioned state
   - Configuration → Distribution
   - Changes → Synchronization

4. **Sync to User**
   - Versioned state → View models
   - Events → Notifications
   - Metrics → Visualizations

### Error Handling Patterns

#### Error Flow
1. **Detection**
   - Hardware errors
   - State errors
   - Operation errors
   - Validation errors

2. **Propagation**
   - Add context
   - Transform for layer
   - Enrich metadata
   - Handle recovery

3. **Presentation**
   - User-friendly messages
   - Error classification
   - Recovery options
   - Guidance

### Recovery Patterns

#### Recovery Flow
1. **Error Detection**
   - Identify source
   - Classify error
   - Assess impact
   - Determine scope

2. **State Recovery**
   - Save current state
   - Roll back changes
   - Validate state
   - Resume operations

3. **Error Resolution**
   - Apply fixes
   - Verify resolution
   - Update state
   - Resume normal operation