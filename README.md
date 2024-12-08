# Wrale Fleet Management Platform

Wrale Fleet is an enterprise-grade IoT fleet management platform designed for managing large-scale Raspberry Pi deployments. The platform supports airgapped environments and provides comprehensive device management capabilities for global organizations.

## Project Status

Current development stage: **Stage 1 - Core Device Management**

This implementation provides the foundational "steel thread" of device management functionality, which will be incrementally enhanced with enterprise features through our demo-driven development approach.

## Features

- Core device management and monitoring
- Multi-tenant architecture
- Status tracking and updates
- Tag-based organization
- Configuration management
- Airgapped deployment support (planned)

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
│   └── store/     # Storage implementations
├── pkg/           # Public library code
└── test/          # Additional test files
```

### Development Workflow

1. Create a new branch for your feature
2. Implement changes following Go best practices
3. Ensure all tests pass with `make test`
4. Verify code quality with `make all`
5. Submit a pull request

## Demo Scenarios

The current implementation supports these demo scenarios:

1. Device Registration
   ```go
   // Register a new device
   device, err := service.Register(ctx, "tenant-id", "Demo Device")
   ```

2. Status Updates
   ```go
   // Update device status
   err := service.UpdateStatus(ctx, "tenant-id", deviceID, device.StatusOnline)
   ```

3. Device Listing
   ```go
   // List all devices for a tenant
   devices, err := service.List(ctx, device.ListOptions{TenantID: "tenant-id"})
   ```

## Contributing

Please read our [contributing guidelines](CONTRIBUTING.md) before submitting pull requests.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Roadmap

1. **Stage 1**: Core Device Management _(current)_
2. **Stage 2**: Fleet Organization
3. **Stage 3**: Multi-tenant Isolation
4. **Stage 4**: Airgapped Operations

## Support

For questions and support, please file an issue in the GitHub repository.