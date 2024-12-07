# Wrale Fleet API Service

RESTful API service providing fleet management interface and real-time data access.

## Features

### Core Features
- REST API endpoints
- WebSocket real-time updates
- Authentication and authorization
- Fleet management interface
- Device control API
- Metric aggregation

### API Groups
- Fleet management
- Device control
- User management
- Analytics
- Configuration

## Make Targets

### Main Targets
- `make all` - Build and verify all components
- `make build` - Build API service
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make verify` - Run all verifications

### API-Specific Targets
- `make api-spec` - Generate OpenAPI specification
- `make openapi-validate` - Validate OpenAPI specification
- `make docker-build` - Build Docker image
- `make docker-push` - Push Docker image

Run `make help` for a complete list of available targets.

## Docker Build

The API service is containerized for deployment:
```dockerfile
# Image: wrale-fleet/api
# Tag: version-${VERSION}
make docker-build
```

## Development

### Prerequisites
- Go 1.21+
- Docker
- OpenAPI tools

### Local Setup
```bash
make all
make api-spec
```

### API Documentation
Generated OpenAPI specification available at:
- Development: http://localhost:8080/swagger/
- Production: https://api.wrale.com/swagger/

## Integration

### Required Services
- Fleet service
- Authentication service
- Database

### Client Libraries
- JavaScript/TypeScript
- Go
- Python

## License
See ../../LICENSE