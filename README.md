# Wrale Fleet Management Platform

Wrale Fleet aims to be an enterprise-grade IoT fleet management platform designed for managing large-scale Raspberry Pi deployments. The platform is being built with a focus on security, multi-tenancy, and comprehensive monitoring capabilities for global organizations.

## Current Status: Stage 1 Development

We are currently implementing Stage 1 core device management capabilities, with a focus on getting the fundamentals right through our demo-driven development approach.

### What Works Now

- Core memory storage
- Foundational multi-tenant data structures
- Structured logging systems

## Build System

The project uses Make for build automation with these key commands:

```bash
make help      # Display available commands
make all       # Run all checks and build the binary
make build     # Build the binary only
make test      # Run unit tests
make lint      # Run linting checks
make sec-check # Run security checks
make run       # Build and run the application
```

## Implementation Progress Matrix

Our demo-driven development tracks progress through user personas and their stories:

### System Administrator
| Story                    | Status |
|--------------------------|--------|
| Server Initialization    |        |
| Device Registration      |        |
| Device Monitoring        |        |
| Configuration Management |        |
| Graceful Shutdown        |        |

### Security Team
| Story                    | Status |
|-------------------------|--------|
| Access Control          |        |
| Security Monitoring     |        |
| Audit Logging           |        |
| Compliance Reporting    |        |

### Operations Team
| Story                    | Status |
|-------------------------|--------|
| Fleet Overview          |        |
| Performance Monitoring  |        |
| Update Management       |        |
| Incident Response       |        |

## Architecture

The project follows standard Go project layout:

```
wrale-fleet/
├── api/            # API definitions
├── cmd/            # Application entry points
├── internal/       # Private application code
│   ├── fleet/     # Core domain logic
│   └── tenant/    # Multi-tenant support
└── web/           # Web interface assets
```

## Development

### Prerequisites

- Go 1.21 or higher
- Make
- golangci-lint (for development)
- gosec (for security checks)

### Getting Started

1. Install development tools:
   ```bash
   make install-tools
   ```

2. Build and test:
   ```bash
   make all
   ```

3. Run the demo:
   ```bash
   make run
   ```

## Future Development

We have an ambitious roadmap planned with these key stages:

1. **Stage 1**: Core Device Management _(current)_
2. **Stage 2**: Fleet Organization _(planned)_
3. **Stage 3**: Multi-tenant Isolation _(planned)_
4. **Stage 4**: Airgapped Operations _(planned)_

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.
