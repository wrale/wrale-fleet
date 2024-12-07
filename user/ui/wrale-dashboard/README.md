# Wrale Fleet Dashboard

Web-based dashboard for fleet management and monitoring. Built with Next.js and React.

## Features

### Core Features
- Real-time fleet monitoring
- Device management interface
- Analytics and metrics visualization
- Configuration management
- Alert management
- Interactive device map

### Components
- Device management
- Analytics dashboards
- Configuration panels
- Interactive maps
- Maintenance tools

## Make Targets

### Main Targets
- `make all` - Build and verify all components
- `make build` - Build dashboard
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make verify` - Run all verifications

### UI-Specific Targets
- `make dev` - Run development server
- `make storybook` - Run Storybook
- `make docker-build` - Build Docker image
- `make docker-push` - Push Docker image

Run `make help` for a complete list of available targets.

## Docker Build

The dashboard is containerized for deployment:
```dockerfile
# Image: wrale-fleet/dashboard
# Tag: version-${VERSION}
make docker-build
```

## Development

### Prerequisites
- Node.js 18+
- npm or yarn
- Docker

### Local Setup
```bash
make dev
```

Development server runs at http://localhost:3000

### Storybook
```bash
make storybook
```

Storybook runs at http://localhost:6006

## Integration

### Required Services
- API service
- Authentication service
- WebSocket service

### Environment Variables
See `.env.example` for required configurations.

## License
See ../../../../LICENSE