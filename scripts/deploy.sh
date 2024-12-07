#!/bin/bash
set -e

# Configuration
VERSION=${1:-"1.0.0"}
ENV=${2:-"production"}
REGISTRY=${DOCKER_REGISTRY:-"localhost:5000"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Deploying Wrale Fleet v${VERSION} to ${ENV}"

# Check required tools
command -v docker >/dev/null 2>&1 || { echo -e "${RED}Error: docker is required but not installed.${NC}" >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo -e "${RED}Error: docker-compose is required but not installed.${NC}" >&2; exit 1; }

# Check environment file
ENV_FILE=".env.${ENV}"
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: Environment file $ENV_FILE not found${NC}"
    exit 1
fi

# Pull images
echo -e "\n${GREEN}Pulling images...${NC}"
docker pull ${REGISTRY}/wrale-fleet/metal:${VERSION}
docker pull ${REGISTRY}/wrale-fleet/fleet:${VERSION}
docker pull ${REGISTRY}/wrale-fleet/api:${VERSION}
docker pull ${REGISTRY}/wrale-fleet/dashboard:${VERSION}

# Stop existing services
echo -e "\n${GREEN}Stopping existing services...${NC}"
docker-compose down || true

# Start services with new images
echo -e "\n${GREEN}Starting services...${NC}"
ENV_FILE=$ENV_FILE docker-compose up -d

# Wait for services to be healthy
echo -e "\n${GREEN}Waiting for services to be healthy...${NC}"
sleep 30

# Check health endpoints
services=("metal:8080" "fleet:8081" "api:8083")
for service in "${services[@]}"; do
    if curl -f "http://${service}/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ ${service} is healthy${NC}"
    else
        echo -e "${RED}✗ ${service} is not healthy${NC}"
        echo -e "${YELLOW}Logs for ${service%%:*}:${NC}"
        docker-compose logs ${service%%:*}
        
        echo -e "${RED}Deployment failed. Rolling back...${NC}"
        docker-compose down
        exit 1
    fi
done

echo -e "\n${GREEN}Deployment complete!${NC}"
echo "Services are running at:
- Metal Service: http://localhost:8080
- Fleet Service: http://localhost:8081
- API Service: http://localhost:8083
- Dashboard: http://localhost:3000"
