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
   chmod +x scripts/build.sh scripts/deploy.sh
   ./scripts/build.sh 1.0.0
   ./scripts/deploy.sh 1.0.0 development
   ```

4. Access the dashboard at http://localhost:3000

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
```bash
# Build all components
./scripts/build.sh VERSION

# Build individual components
cd metal && go build ./...
cd fleet && go build ./...
cd user/api && go build ./...
cd user/ui/wrale-dashboard && npm install && npm run build
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run UI tests
cd user/ui/wrale-dashboard && npm test
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
   ./scripts/deploy.sh VERSION production
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
