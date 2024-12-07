# Wrale Fleet Edge Agent

Edge agent for direct device management and local state control. Provides local device control with fleet coordination.

## Features

### Core Features
- Direct device control
- Local state management
- Metal daemon integration
- Brain service communication
- Real-time metrics collection
- Local policy enforcement

### Components
- Agent: Core management
- Client: Service communication
- Store: Local state

## Make Targets

### Main Targets
- `make all` - Build and verify all components
- `make build` - Build edge agent
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make verify` - Run all verifications

### Agent-Specific Targets
- `make install-deps` - Install agent dependencies
- `make package` - Create agent package
- `make verify-security` - Run security checks

Run `make help` for a complete list of available targets.

## Development

### Prerequisites
- Go 1.21+
- Access to metal daemon
- Access to brain service

### Local Setup
```bash
make all
make install-deps
```

### Configuration
See `/config` for configuration options and examples.

## Integration

### Required Services
- Metal daemon
- Brain service

### Communication
- NATS for messaging
- gRPC for metal daemon
- WebSocket for metrics

## License
See ../../LICENSE