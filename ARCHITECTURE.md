# Wrale Fleet Architecture

This document describes the high-level architecture of the Wrale Fleet system. The system follows a physical-first philosophy where hardware concerns and environmental factors drive the architecture.

## Core Components

### metal/ - Hardware Layer
The lowest level of the system, responsible for direct hardware interaction.

- **hw/** - Direct hardware control
  - gpio/ - GPIO pin management
  - power/ - Power management and monitoring
  - secure/ - Security and tamper detection
  - thermal/ - Temperature monitoring and cooling
  - diag/ - Hardware diagnostics

- **core/** - System coordination
  - cmd/metald - Main metal daemon
  - server/ - HTTP API for hardware control
  - secure/ - Security policies
  - thermal/ - Thermal management policies

- **diag/** - System-wide diagnostics and monitoring

### fleet/ - Management Layer
Coordinates multiple devices and manages fleet-wide concerns.

- **brain/** - Central coordination
  - Device coordination
  - Policy management
  - Fleet-wide decisions

- **edge/** - Edge device management
  - Communication with devices
  - State management
  - Local decision making

- **sync/** - Synchronization
  - Data synchronization between devices
  - Configuration distribution
  - State reconciliation

### user/ - Interface Layer
User-facing components for fleet management.

- **api/** - Backend API
  - REST endpoints
  - Authentication/Authorization
  - Request handling

- **ui/** - Frontend dashboard
  - React/Next.js based interface
  - Physical-first visualization
  - Real-time monitoring

### shared/ - Common Infrastructure
Shared code and utilities used across components.

- **config/** - Configuration management
- **docs/** - Documentation
- **testing/** - Common test utilities
- **tools/** - Development tools

## Key Architectural Principles

1. **Physical-First Philosophy**
   - All decisions must consider physical world implications
   - Hardware state drives software behavior
   - Environmental factors are primary concerns

2. **Hardware is First-Class**
   - Direct hardware access through metal layer
   - Real-time monitoring and response
   - Physical safety checks

3. **Environmental Awareness**
   - Temperature monitoring and management
   - Power consumption optimization
   - Physical security considerations

## Data Flow

1. Physical hardware sends signals/data to metal/hw
2. metal/core processes and coordinates hardware interactions
3. fleet/ layer manages device coordination
4. user/ layer provides visualization and control

## Security Model

- Hardware-level security through metal/hw/secure
- Policy enforcement through metal/core/secure
- Fleet-wide security coordination through fleet/brain
- User authentication and authorization in user/api

## Error Handling

1. Hardware-level errors handled by metal/hw
2. System-level diagnostics through metal/diag
3. Fleet-wide error coordination through fleet/brain
4. User notification through user/ui

## Future Considerations

1. Enhanced environmental monitoring
2. Advanced power management
3. Expanded diagnostic capabilities
4. Improved fleet coordination algorithms