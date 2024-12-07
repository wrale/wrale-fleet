# Wrale Fleet Service

Central fleet management service responsible for device coordination, fleet-wide orchestration, and state management.

## Features

### Core Features
- Device coordination through brain service
- Fleet-wide state management
- Edge agent orchestration
- Real-time telemetry processing
- Configuration management
- Policy enforcement

### Components
- Brain: Central coordination
- Edge: Device management
- Sync: State synchronization

## Make Targets

### Main Targets
- `make all` - Build and verify all components
- `make build` - Build fleet service
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make verify` - Run all verifications

### Service-Specific Targets
- `make integration-test` - Run integration tests
- `make docker-build` - Build Docker image
- `make docker-push` - Push Docker image
- `make package` - Create deployable package

Run `make help` for a complete list of available targets.

## Docker Build

The fleet service is containerized for deployment:
```dockerfile
# Image: wrale-fleet/fleet
# Tag: version-${VERSION}
make docker-build
```

## Development

### Prerequisites
- Go 1.21+
- Docker
- Access to brain and edge services

### Local Setup
```bash
make all
make integration-test
```

### Configuration
See `fleet/config/` for configuration options and examples.

## Integration

### Required Services
- Metal daemon
- Edge agents
- API service

### Service Dependencies
- etcd (for state)
- NATS (for messaging)
- Prometheus (for metrics)

## License
See ../LICENSE