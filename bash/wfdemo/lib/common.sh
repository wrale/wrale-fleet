#!/usr/bin/env bash

# Common configuration and utilities for demo scripts
# This file establishes consistent behavior across our demo environment

# Demo temporary directory for runtime files
export DEMO_TMP_DIR="/tmp/wfdemo"

# Port allocations for demo environment
# We use dedicated port ranges for each service to avoid conflicts
# These defaults can be overridden by environment variables
export WFCENTRAL_API_PORT=${WFCENTRAL_API_PORT:-8600}     # Main API for device management
export WFCENTRAL_MGMT_PORT=${WFCENTRAL_MGMT_PORT:-8601}   # Health and management endpoints
export WFDEVICE_API_PORT=${WFDEVICE_API_PORT:-8700}       # Device agent API
export WFDEVICE_MGMT_PORT=${WFDEVICE_MGMT_PORT:-8701}     # Device health endpoints

# Print debug information about port configuration
echo "DEBUG: Using ports:"
echo "  WFCENTRAL_API_PORT=${WFCENTRAL_API_PORT}"
echo "  WFCENTRAL_MGMT_PORT=${WFCENTRAL_MGMT_PORT}"
echo "  WFDEVICE_API_PORT=${WFDEVICE_API_PORT}"
echo "  WFDEVICE_MGMT_PORT=${WFDEVICE_MGMT_PORT}"

# Text formatting for prettier output
BOLD="\033[1m"
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
BLUE="\033[34m"
RESET="\033[0m"

# begin_story sets up the context for a demo scenario
begin_story() {
    local persona=$1
    local stage=$2
    local story=$3

    echo -e "${BOLD}Demo Story: $story${RESET}"
    echo -e "${BLUE}Persona:${RESET} $persona"
    echo -e "${BLUE}Stage:${RESET} $stage"
    echo
}

# explain provides context about what we're demonstrating
explain() {
    echo -e "${BLUE}→${RESET} $1"
}

# step indicates progress through the demo
step() {
    echo -e "\n${YELLOW}▶${RESET} $1"
}

# success indicates a successful outcome
success() {
    echo -e "${GREEN}✓${RESET} $1"
}

# error indicates a failure condition
error() {
    echo -e "${RED}✗${RESET} $1"
}

# setup_demo_env prepares the environment for demo execution
setup_demo_env() {
    # Create clean demo directory
    rm -rf "${DEMO_TMP_DIR}"
    mkdir -p "${DEMO_TMP_DIR}"

    # Create example configuration
    cat > "${DEMO_TMP_DIR}/wfcentral.yaml" << EOF
# wfcentral configuration for demo environment
# This shows recommended settings for development/testing

# Main API endpoint for device management
port: ${WFCENTRAL_API_PORT}

# Health and management endpoint 
# Separated for security best practices
management:
  port: ${WFCENTRAL_MGMT_PORT}
  # Control how much information is exposed
  # Options: minimal, standard, full
  exposure_level: standard

# Data storage location
data_dir: ${DEMO_TMP_DIR}/central

# Logging configuration
log_level: info

# Stage 1 specific settings
stage1:
  device_storage: memory
EOF

    # Create example device configuration
    cat > "${DEMO_TMP_DIR}/wfdevice.yaml" << EOF
# wfdevice configuration for demo environment
# This shows recommended settings for development/testing

# Main API endpoint for device operations
port: ${WFDEVICE_API_PORT}

# Health and management endpoint
# Separated for security best practices
management:
  port: ${WFDEVICE_MGMT_PORT}
  exposure_level: standard

# Data storage location
data_dir: ${DEMO_TMP_DIR}/device

# Connection to control plane
control_plane:
  host: localhost
  port: ${WFCENTRAL_API_PORT}

# Logging configuration  
log_level: info
EOF

    # Print the effective configuration for debugging
    echo "DEBUG: Created configuration files with ports:"
    echo "  API Port: ${WFCENTRAL_API_PORT}"
    echo "  Management Port: ${WFCENTRAL_MGMT_PORT}"
}

# cleanup_demo_env ensures proper cleanup after demo execution
cleanup_demo_env() {
    local pid_file="${DEMO_TMP_DIR}/wfcentral.pid"
    if [ -f "$pid_file" ]; then
        if kill -0 "$(cat "$pid_file")" 2>/dev/null; then
            kill "$(cat "$pid_file")"
        fi
    fi
    rm -rf "${DEMO_TMP_DIR}"
}

# Set up cleanup trap
trap cleanup_demo_env EXIT
