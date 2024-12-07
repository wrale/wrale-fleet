# Wrale Fleet Build Guide

## Prerequisites

### System Requirements
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- npm or yarn
- Git

### Environment Setup
```bash
# Clone repository
git clone https://github.com/wrale/wrale-fleet.git
cd wrale-fleet

# Set up development environment
cp .env.development .env
```

## Core Make Targets

The project uses a standardized build system with consistent make targets across all components:

```bash
# Build everything
make all

# Clean all build artifacts
make clean

# Run tests across all components
make test

# Run all verifications
make verify
```

## Component-Specific Builds

### Metal Layer
```bash
# Build all metal components
make metal

# Build metal daemon
make metal/core

# Run metal tests
make metal test

# Hardware-specific targets
make metal hardware-test     # Run hardware-specific tests
make metal simulation       # Run in simulation mode
make metal calibrate       # Run sensor calibration
```

### Fleet Layer
```bash
# Build fleet service
make fleet

# Build edge agent
make fleet/edge

# Run fleet tests
make fleet test

# Fleet-specific targets
make fleet integration-test  # Run integration tests
make fleet package          # Create deployable package
```

### Sync Layer
```bash
# Build sync service
make sync

# Run sync tests
make sync test

# Sync-specific targets
make sync integration-test    # Run integration tests
make sync verify-consistency # Verify state consistency
make sync verify-conflict   # Test conflict resolution
make sync benchmark        # Run sync benchmarks

# Development targets
make sync dev           # Run with development config
make sync simulation    # Run with simulated network
```

### User Layer - API
```bash
cd user/api

# Build API service
make build

# Generate OpenAPI spec
make api-spec

# Validate OpenAPI spec
make openapi-validate

# Run API tests
make test
```

### User Layer - Dashboard
```bash
cd user/ui/wrale-dashboard

# Install dependencies
npm install

# Development server
make dev        # Runs at http://localhost:3000

# Production build
make build

# Run UI tests
make test

# Run Storybook
make storybook  # Runs at http://localhost:6006
```

## Docker Builds

### Building Images
```bash
# Build all containers
make docker-build

# Component-specific builds
make metal docker-build    # wrale-fleet/metal:${VERSION}
make fleet docker-build    # wrale-fleet/fleet:${VERSION}
make user/api docker-build # wrale-fleet/api:${VERSION}
make user/ui docker-build  # wrale-fleet/dashboard:${VERSION}

# Push images
make docker-push
```

### Running with Docker Compose
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down
```

## Development Workflow

### Standard Development Cycle
```bash
# 1. Start development environment
make dev

# 2. Run verifications
make verify

# 3. Run tests
make test

# 4. Build components
make build

# 5. Create deployable artifacts
make package
```

### Component Development

#### Metal Development
```bash
cd metal

# Build with simulation enabled
make build SIMULATION=1

# Run tests with hardware
make test HARDWARE=1

# Package for deployment
make package
```

#### Fleet Development
```bash
cd fleet

# Build brain service
make brain

# Build edge agent
make edge

# Run integration tests
make integration-test
```

#### UI Development
```bash
cd user/ui/wrale-dashboard

# Start dev server with hot reload
make dev

# Build for production
make build

# Run component tests
make test
```

## Verification Targets

```bash
# Run all verifications
make verify

# Security checks
make verify-security

# Type checking
make verify-types

# Lint checking
make verify-lint

# License verification
make verify-license
```

## Build Artifacts

### Output Locations
```
./tmp/metal/       # Metal build outputs
./tmp/fleet/       # Fleet build outputs
./tmp/sync/        # Sync build outputs
./tmp/user/api/    # API build outputs
./tmp/user/ui/     # UI build outputs
```

### Packages
```bash
# Create deployment packages
make package

# Outputs:
./dist/metal-${VERSION}.tar.gz
./dist/fleet-${VERSION}.tar.gz
./dist/sync-${VERSION}.tar.gz
./dist/api-${VERSION}.tar.gz
./dist/dashboard-${VERSION}.tar.gz
```

## Common Issues

### Build Failures
1. Check Go version matches required 1.21+
2. Ensure all dependencies are installed
3. Verify environment variables are set
4. Clean build artifacts and retry

### Test Failures
1. Ensure simulation mode for hardware tests
2. Check network connectivity for integration tests
3. Verify database is running for API tests
4. Confirm Node.js version for UI tests

## Environment Variables

See .env.example for all required configurations:
```bash
# Core settings
WRALE_ENV=development
LOG_LEVEL=debug

# Metal settings
METAL_SIMULATION=true
HARDWARE_TESTING=false

# Fleet settings
FLEET_EDGE_COUNT=1
SYNC_ENABLED=true

# Sync settings
SYNC_MODE=distributed
SYNC_STORE_TYPE=etcd
SYNC_CONFLICT_STRATEGY=latest
SYNC_MAX_BATCH_SIZE=1000
SYNC_NETWORK_TIMEOUT=5s

# API settings
API_PORT=8080
JWT_SECRET=development-secret

# UI settings
NEXT_PUBLIC_API_URL=http://localhost:8080
```