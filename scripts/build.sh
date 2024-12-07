#!/bin/bash
set -e

# Configuration
VERSION=${1:-"1.0.0"}
REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "Building Wrale Fleet v${VERSION}"

# Check required tools
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Error: docker is required but not installed.${NC}" >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo -e "${RED}Error: docker-compose is required but not installed.${NC}" >&2; exit 1; }

# Build all services
echo -e "\n${GREEN}Building services...${NC}"
docker-compose build

# Tag images
echo -e "\n${GREEN}Tagging images...${NC}"
docker tag wrale-fleet_metal ${REGISTRY}/wrale-fleet/metal:${VERSION}
docker tag wrale-fleet_fleet ${REGISTRY}/wrale-fleet/fleet:${VERSION}
docker tag wrale-fleet_api ${REGISTRY}/wrale-fleet/api:${VERSION}
docker tag wrale-fleet_dashboard ${REGISTRY}/wrale-fleet/dashboard:${VERSION}

# Basic verification
echo -e "\n${GREEN}Running basic verification...${NC}"

echo "Starting services..."
docker-compose up -d

# Wait for services to be healthy
echo "Waiting for services to be healthy..."
sleep 30

# Check health endpoints
services=("metal:8080" "fleet:8081" "api:8083")
for service in "${services[@]}"; do
    if curl -f "http://${service}/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ ${service} is healthy${NC}"
    else
        echo -e "${RED}✗ ${service} is not healthy${NC}"
        docker-compose logs ${service%%:*}
        docker-compose down
        exit 1
    fi
done

# Stop services
docker-compose down

echo -e "\n${GREEN}Build complete!${NC}"
echo "Images are ready to be pushed:
- ${REGISTRY}/wrale-fleet/metal:${VERSION}
- ${REGISTRY}/wrale-fleet/fleet:${VERSION}
- ${REGISTRY}/wrale-fleet/api:${VERSION}
- ${REGISTRY}/wrale-fleet/dashboard:${VERSION}"
