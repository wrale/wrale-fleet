# Wrale Fleet v1.0 Architecture

This document extends ARCHITECTURE.md with v1.0-specific implementation details and deployment architecture.

## v1.0 Component Status

### Metal Layer
- ✓ Hardware Control
  - GPIO management
  - Power monitoring
  - Thermal control
  - Security monitoring
- ✓ System Coordination
  - Metal daemon
  - Hardware API
  - Policy enforcement
- ✓ Diagnostics
  - Health monitoring
  - Error tracking

### Fleet Layer
- ✓ Brain
  - Device coordination
  - Fleet optimization
  - State management
- ✓ Edge
  - Device control
  - Local autonomy
  - Hardware interface
- ✓ Sync
  - State synchronization
  - Configuration management
  - Conflict resolution

### User Layer
- ✓ API
  - REST endpoints
  - WebSocket updates
  - Authentication
- ✓ Dashboard
  - Device management
  - Real-time monitoring
  - Configuration interface

## v1.0 Integration Points

### Hardware to Metal
```
Physical Hardware
    │
    ▼
metal/hw/gpio
    │
    ▼
metal/hw/{power,thermal,secure}
    │
    ▼
metal/core/server
```

### Metal to Fleet
```
metal/core/server
    │
    ▼
fleet/edge/client
    │
    ▼
fleet/brain/coordinator
    │
    ▼
fleet/sync/manager
```

### Fleet to User
```
fleet/brain/service
    │
    ▼
user/api/service
    │
    ▼
user/ui/dashboard
```

## v1.0 Deployment Architecture

### Container Structure
```
┌─────────────────┐     ┌─────────────────┐
│    Dashboard    │     │       API       │
│   (Next.js)    │────▶│      (Go)       │
└─────────────────┘     └────────┬────────┘
                               ▲  │
                               │  ▼
┌─────────────────┐     ┌─────────────────┐
│  Fleet Brain    │◀───▶│   Fleet Edge    │
│      (Go)      │     │      (Go)       │
└────────┬────────┘     └────────┬────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────────┐
│               Metal Layer               │
│                  (Go)                   │
└─────────────────────────────────────────┘
```

### Data Storage
- Metal: Local file system for device state
- Fleet: Memory-based with file backup
- API: No persistent storage (stateless)
- UI: Browser storage for preferences

### Network Flow
1. Dashboard → API (HTTPS)
2. API → Fleet Brain (HTTP/WS)
3. Brain ↔ Edge (HTTP)
4. Edge → Metal (HTTP)
5. Metal → Hardware (Direct)

## v1.0 Security Model

### Authentication
1. User → Dashboard: Session-based
2. Dashboard → API: JWT
3. API → Fleet: Service tokens
4. Edge → Metal: API keys

### Authorization
1. User roles (admin, user, viewer)
2. Resource-based access control
3. Command validation
4. Operation auditing

### Encryption
1. HTTPS for external access
2. HTTP for internal services
3. Secure WebSocket for real-time
4. Token encryption

### Hardware Security
1. Physical access monitoring
2. Tamper detection
3. Secure boot process
4. Hardware key storage

## v1.0 Error Handling

### Error Propagation
```
Hardware Error
    │
    ▼
Metal Error Handler
    │
    ▼
Edge Error Manager
    │
    ▼
Brain Coordinator
    │
    ▼
API Error Response
    │
    ▼
UI Error Display
```

### Recovery Procedures
1. Hardware: Safe mode fallback
2. Metal: State recovery
3. Fleet: Connection retry
4. API: Request retry
5. UI: Auto reconnect

## v1.0 Monitoring

### Metrics Collection
1. Hardware metrics (temp, power)
2. System metrics (CPU, memory)
3. Service metrics (latency, errors)
4. User metrics (requests, sessions)

### Health Checks
1. Hardware health
2. Service health
3. Connection health
4. System health

## Future Considerations

### v1.1 Planned Features
1. Enhanced cluster support
2. Advanced analytics
3. Predictive maintenance
4. Extended security

### Technical Debt
1. Metric aggregation
2. Error correlation
3. Testing coverage
4. Documentation