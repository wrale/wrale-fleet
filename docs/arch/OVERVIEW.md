# Wrale Fleet System Architecture Overview

## Core Philosophy

The Wrale Fleet system follows a physical-first architecture where hardware concerns and environmental factors drive the design decisions. The system is built to manage and coordinate distributed Raspberry Pi devices with strong emphasis on hardware interaction, environmental awareness, and physical safety.

## System Layers

```
User Layer (Presentation)
    ↑
    ├── UI (Next.js/TypeScript)
    ├── API (Go/REST)
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

## Core Components

### Metal Layer (`metal/`)
- Direct hardware interaction
- System diagnostics
- Physical safety monitoring
- Environmental control

### Fleet Layer (`fleet/`)
- Device coordination
- Resource management
- State synchronization
- Policy enforcement

### Sync Layer (`sync/`)
- State versioning
- Configuration distribution
- Conflict resolution
- Consistency management

### User Layer (`user/`)
- Web dashboard (UI)
- REST API
- WebSocket updates
- User management

### Shared Infrastructure (`shared/`)
- Common utilities
- Configuration
- Testing tools
- Type definitions

## Key Architectural Principles

1. **Physical-First Philosophy**
   - Hardware state drives software behavior
   - Environmental awareness is critical
   - Physical safety is paramount

2. **Layer Independence**
   - Clear interface contracts between layers
   - Explicit type transformations at boundaries
   - No direct dependencies between non-adjacent layers

3. **Type Safety**
   - Strong typing across all layers
   - Validation at layer boundaries
   - Clear type hierarchy

4. **Security by Design**
   - Hardware-level security
   - Policy enforcement at each layer
   - Secure communication channels

## Cross-Cutting Concerns

### Error Handling
- Hardware-level error detection
- Error propagation through layers
- User-friendly error reporting

### Monitoring
- Physical metrics collection
- Performance monitoring
- Health checking
- Alert management

### State Management
- Versioned state tracking
- Consistent state propagation
- Conflict resolution
- Recovery procedures

## Evolution Strategy

1. **Backward Compatibility**
   - Maintain interface stability
   - Version types and APIs
   - Support gradual migration

2. **Feature Development**
   - Physical safety first
   - Hardware compatibility
   - Environmental awareness
   - User experience

3. **Technical Debt**
   - Regular security updates
   - Performance optimization
   - Documentation maintenance
   - Test coverage

See additional architecture documents for detailed specifications:
- LAYERS.md - Detailed layer documentation
- SECURITY.md - Security architecture
- API.md - API documentation
- DEPLOYMENT.md - Deployment architecture