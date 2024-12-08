# Wrale Fleet Management Platform

Wrale Fleet is an enterprise-grade IoT fleet management platform designed for managing large-scale Raspberry Pi deployments. The platform provides comprehensive device management capabilities with a focus on security, multi-tenancy, and monitoring for global organizations.

## Project Status

Current development stage: **Stage 1 - Core Device Management with Enterprise Features**

This implementation provides the foundational capabilities with enterprise-grade features:

- Complete security event auditing and monitoring
- Multi-tenant architecture with strict isolation
- Continuous demonstration and validation
- Real-time status monitoring and updates

Each feature is incrementally enhanced through our demo-driven development approach, ensuring working functionality at every stage.

## Enterprise Features

### Security & Compliance
- Security event monitoring and auditing
- Multi-tenant data isolation
- Role-based access control (planned)
- Compliance reporting framework (planned)

### Monitoring & Operations
- Real-time device status tracking
- Centralized logging and monitoring
- Automated status updates
- Performance metrics collection

### Core Management
- Device registration and provisioning
- Tag-based organization
- Configuration management
- Fleet-wide status tracking

### Enterprise Readiness
- Multi-tenant architecture
- Continuous demonstration capability
- Airgapped deployment support (planned)
- Hierarchical group management (planned)

## Prerequisites

- Go 1.21 or higher
- Make
- golangci-lint (for development)
- gosec (for security checks)

## Quick Start

1. Install development tools:
   ```bash
   make install-tools
   ```

2. Build the project:
   ```bash
   make all
   ```

3. Run the demo:
   ```bash
   make run
   ```

## Continuous Demo

The platform includes a continuous demo manager (`demo_manager.go`) that demonstrates enterprise capabilities in real-time. Running `make run` shows:

```
2024-12-08T10:48:00.614-0500    INFO    device/service.go:47      registered new device     {"device_id": "6599c64f-ce91-4952-9409-6662f48afe9d", "tenant_id": "demo-tenant", "name": "Demo Raspberry Pi"}
2024-12-08T10:48:00.615-0500    INFO    device/monitor.go:62      security event   {"component": "security_monitor", "event_type": "authentication", "device_id": "6599c64f-ce91-4952-9409-6662f48afe9d", "tenant_id": "demo-tenant", "timestamp": "2024-12-08T15:48:00.615Z", "success": true, "actor": "system"}
```

The demo showcases:
- Automated device provisioning
- Real-time security monitoring
- Status tracking and updates
- Multi-tenant operations

## Development

### Build Commands

The project uses Make for build automation. Key commands include:

- `make help`: Display available commands
- `make all`: Run all checks and build the binary
- `make build`: Build the binary
- `make test`: Run tests
- `make coverage`: Generate test coverage report
- `make lint`: Run linting checks
- `make sec-check`: Run security checks
- `make run`: Build and run the application
- `make dev`: Run with hot reload (requires air)

### Project Structure

```
wrale-fleet/
├── api/            # API definitions
├── cmd/            # Application entry points
├── internal/       # Private application code
│   ├── fleet/     # Core domain logic
│   │   ├── device/    # Device management
│   │   ├── config/    # Configuration management
│   │   └── group/     # Group management
│   ├── store/     # Storage implementations
│   └── tenant/    # Multi-tenant support
├── pkg/           # Public library code
└── test/          # Additional test files
```

### Development Workflow

1. Create a new branch for your feature
2. Implement changes following Go best practices
3. Ensure all tests pass with `make test`
4. Verify code quality with `make all`
5. Submit a pull request

## API Examples

Key operations supported by the platform:

1. Device Management
```go
// Register a new device with tenant isolation
device, err := service.Register(ctx, "tenant-id", "Demo Device")

// Update device status with security auditing
err := service.UpdateStatus(ctx, "tenant-id", deviceID, device.StatusOnline)

// List tenant devices with filtering
devices, err := service.List(ctx, device.ListOptions{
    TenantID: "tenant-id",
    Status: device.StatusOnline,
})
```

2. Security Monitoring
```go
// Security events are automatically captured
monitor.RecordAuthAttempt(ctx, deviceID, tenantID, actor, success, details)
monitor.RecordStatusChange(ctx, deviceID, tenantID, oldStatus, newStatus)
```

## Roadmap

1. **Stage 1**: Core Device Management with Enterprise Features _(current)_
   - ✓ Multi-tenant architecture
   - ✓ Security monitoring
   - ✓ Status tracking
   - ✓ Configuration management

2. **Stage 2**: Fleet Organization
   - Hierarchical group management
   - Tag-based policies
   - Fleet-wide operations

3. **Stage 3**: Multi-tenant Isolation
   - Enhanced tenant isolation
   - Resource quotas
   - Tenant-specific policies

4. **Stage 4**: Airgapped Operations
   - Offline operation support
   - Secure data synchronization
   - Air-gap deployment tools

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

## Support

For questions and support, please file an issue in the GitHub repository.
