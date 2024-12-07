# Wrale Fleet Metal Core

Core system management for Wrale Fleet Metal. Handles state coordination, event processing, and system operations.

## Core Features

### State Management
- System state tracking
- State persistence
- State synchronization
- Recovery management

### Event Processing
- Hardware events
- System events
- Error handling
- Event routing

### System Operations
- Startup/shutdown
- Recovery procedures
- Resource management
- Operation coordination

### Configuration
- System configuration
- Hardware settings
- Operational parameters
- State validation

## Make Targets

The metal core component provides the following make targets:

### Main Targets
- `make all` - Build everything including verification
- `make build` - Build the metal daemon
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make verify` - Run all verifications

### Daemon-Specific Targets
- `make package` - Create deployable daemon package
- `make install` - Install the daemon
- `make verify-security` - Run security checks

Run `make help` for a complete list of available targets.

## Integration

Integrates with:
- ../hw for hardware control
- ../diag for monitoring

## License

See ../../LICENSE