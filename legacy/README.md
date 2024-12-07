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
- **Sync Layer**: State versioning and conflict resolution
- **User Layer**: API and dashboard interface
- **Shared Layer**: Common utilities and types

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
- `make docker-build` - Build all Docker images
- `make deploy` - Deploy all services

### Component-Specific Builds
- `make metal` - Build metal layer (hardware management)
- `make fleet` - Build fleet services
- `make sync` - Build sync services
- `make user/api` - Build API service
- `make user/ui/wrale-dashboard` - Build UI dashboard

### Component Testing
- `make metal test SIMULATION=1` - Run metal tests with hardware simulation
- `make fleet integration-test` - Run fleet integration tests
- `make sync verify-consistency` - Test sync layer consistency
- `make user/api test` - Run API tests
- `make user/ui/wrale-dashboard test` - Run UI tests

Run `make help` to see all available targets for each component.

## Configuration

### Environment Variables
Key configuration variables (see .env.example for full list):
```bash
# Core settings
WRALE_ENV=development
LOG_LEVEL=debug

# Metal settings
METAL_SIMULATION=true    # Enable hardware simulation
HARDWARE_TESTING=false   # Use real hardware for tests

# Fleet settings
FLEET_EDGE_COUNT=1      # Number of edge nodes
SYNC_ENABLED=true       # Enable state sync

# Sync settings
SYNC_MODE=distributed          # Sync architecture mode
SYNC_STORE_TYPE=etcd          # State storage backend
SYNC_CONFLICT_STRATEGY=latest # Conflict resolution strategy
SYNC_MAX_BATCH_SIZE=1000     # Max batch size for updates
SYNC_NETWORK_TIMEOUT=5s      # Network operation timeout

# API settings
API_PORT=8080
JWT_SECRET=development-secret

# UI settings
NEXT_PUBLIC_API_URL=http://localhost:8080
```

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
│   ├── cmd/        # Metal daemon
│   ├── core/       # Core functionality
│   ├── diag/       # Diagnostics
│   ├── hw/         # Hardware abstraction
│   └── internal/   # Internal packages
├── fleet/          # Fleet management layer
│   ├── brain/      # Central coordination
│   ├── edge/       # Edge device management
│   ├── engine/     # Analysis engine
│   └── types/      # Fleet types
├── sync/           # Sync layer
│   ├── manager/    # Sync management
│   ├── resolver/   # Conflict resolution
│   ├── store/      # State storage
│   └── types/      # Sync types
├── user/           # User interface layer
│   ├── api/        # Backend API
│   └── ui/         # Next.js dashboard
├── shared/         # Shared utilities
│   ├── config/     # Common configuration
│   └── types/      # Shared types
└── docs/           # Documentation
```

### Hardware Simulation
The metal layer includes a comprehensive hardware simulation system for development:

```bash
# Enable simulation mode
export METAL_SIMULATION=true

# Start simulated environment
make metal sim-start

# Run tests with simulation
make metal test

# Trigger simulated events
go run cmd/hwsim/main.go trigger --event power_loss

# Stop simulation
make metal sim-stop
```

Simulated components include:
- GPIO pins and PWM
- Power states and sources
- Temperature sensors
- Security sensors
- Hardware faults

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
Standard images:
- wrale-fleet/metal - Hardware management
- wrale-fleet/fleet - Fleet coordination
- wrale-fleet/sync - State synchronization
- wrale-fleet/api - REST API
- wrale-fleet/dashboard - Web UI

## API Documentation

### Metal API
Hardware management endpoints:
```
GET    /api/v1/devices         # List all devices
GET    /api/v1/devices/{id}    # Get device details
POST   /api/v1/devices/{id}/command  # Execute device command
GET    /api/v1/devices/{id}/gpio     # Get GPIO states
POST   /api/v1/devices/{id}/gpio     # Set GPIO state
GET    /api/v1/devices/{id}/power    # Get power state
GET    /api/v1/devices/{id}/thermal  # Get thermal state
```

### Fleet API
Fleet management endpoints:
```
GET    /api/v1/fleet/state    # Get fleet state
POST   /api/v1/fleet/command  # Execute fleet command
PUT    /api/v1/fleet/config   # Update fleet config
GET    /api/v1/fleet/metrics  # Get fleet metrics
GET    /api/v1/fleet/devices  # List fleet devices
```

### Sync API
State synchronization endpoints:
```
GET    /api/v1/sync/state      # Get sync state
POST   /api/v1/sync/update     # Update state
GET    /api/v1/sync/conflicts  # List conflicts
POST   /api/v1/sync/resolve    # Resolve conflict
GET    /api/v1/sync/versions   # List versions
```

### WebSocket API
Real-time event streams:
```
WS   /api/v1/ws                    # Main event stream
WS   /api/v1/ws/devices/{id}       # Device events
WS   /api/v1/ws/fleet             # Fleet events
WS   /api/v1/ws/alerts            # Alert events
```

## Contributing
See CONTRIBUTING.md for guidelines.

## License
See LICENSE for details.

## Support
- GitHub Issues: [Issues](https://github.com/wrale/wrale-fleet/issues)
- Documentation: [Docs](docs/)

## Authors
Wrale Team