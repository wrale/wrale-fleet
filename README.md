# Wrale Fleet v1.0

Wrale Fleet is a comprehensive hardware fleet management system with a physical-first philosophy. It provides real-time monitoring, control, and optimization of hardware devices with emphasis on physical safety and environmental awareness.

## Features

### Core Features
- Physical device management and monitoring
- Real-time metrics and telemetry
- Temperature and power optimization
- Security monitoring and policy enforcement
- Fleet-wide coordination and synchronization
- Web-based dashboard with real-time updates

### Key Components
- **Metal Layer**: Direct hardware interaction and control
- **Fleet Layer**: Device coordination and fleet management
- **User Layer**: API and dashboard interface

## Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for development)
- Node.js 18+ (for UI development)
- Git

## Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/wrale/wrale-fleet.git
   cd wrale-fleet
   ```

2. Set up environment:
   ```bash
   cp .env.development .env
   # Edit .env with your configurations
   ```

3. Build and start services:
   ```bash
   make all
   ```

4. Access the dashboard at http://localhost:3000

## Make Targets

The project uses a standardized build system with make targets across all components:

### Core Targets
- `make all` - Build all components
- `make clean` - Clean all build artifacts
- `make test` - Run tests across all components
- `make verify` - Run verification checks

### Component-Specific Builds
- `make fleet` - Build fleet service
- `make fleet/edge` - Build edge agent
- `make metal/core` - Build metal daemon
- `make user/api` - Build API service
- `make user/ui/wrale-dashboard` - Build UI dashboard

Run `make help` to see all available targets for each component.

## Configuration

### Environment Variables
See `.env.production` and `.env.development` for available configurations:
- Service ports and endpoints
- Security settings
- Resource limits
- Feature flags

### Security
- Set strong JWT_SECRET and METAL_API_KEY in production
- Configure proper authentication for all services
- Use HTTPS in production environments
- Set appropriate resource limits

## Development

### Project Structure
```
.
├── metal/          # Hardware interaction layer
├── fleet/          # Fleet management layer
│   ├── brain/      # Central coordination
│   ├── edge/       # Device management
│   └── sync/       # State synchronization
├── user/           # User interface layer
│   ├── api/        # Backend API
│   └── ui/         # Web dashboard
└── shared/         # Shared utilities
```

### Building Components
Use make targets for building:
```bash
# Build all components
make all

# Build individual components
make fleet
make metal/core
make user/api
make user/ui/wrale-dashboard
```

### Running Tests
```bash
# Run all tests
make test

# Run component-specific tests
make fleet test
make user/ui/wrale-dashboard test
```

## Deployment

### Production Deployment
1. Configure production environment:
   ```bash
   cp .env.production .env
   # Edit .env with production settings
   ```

2. Deploy services:
   ```bash
   make all
   make deploy
   ```

### Docker Images
- wrale-fleet/metal
- wrale-fleet/fleet
- wrale-fleet/api
- wrale-fleet/dashboard

## API Documentation

### Metal API
- `GET /api/v1/devices` - List all devices
- `GET /api/v1/devices/{id}` - Get device details
- `POST /api/v1/devices/{id}/command` - Execute device command

### Fleet API
- `GET /api/v1/fleet/metrics` - Get fleet metrics
- `POST /api/v1/fleet/command` - Execute fleet-wide command
- `PUT /api/v1/fleet/config` - Update fleet configuration

### WebSocket API
- Connect to `/api/v1/ws` for real-time updates
- Subscribe to specific devices with `?device=id1,id2`

## Contributing
See CONTRIBUTING.md for guidelines.

## License
See LICENSE for details.

## Support
- GitHub Issues: [Issues](https://github.com/wrale/wrale-fleet/issues)
- Documentation: [Docs](docs/)

## Authors
Wrale Team