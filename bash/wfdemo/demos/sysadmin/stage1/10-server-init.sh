#!/usr/bin/env bash

# Enable debug output for troubleshooting
set -x

# Stage 1 - Server Initialization
# This script demonstrates starting the central control plane, which is the core
# server component that manages our entire fleet of devices. The control plane
# needs to be running before any devices can connect to it.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Control Plane Initialization"

explain "As a system administrator, I need to start the central control plane"
explain "This server must be running before any devices can connect"
explain "We use dedicated port ranges to ensure consistent configuration:"
explain "  - Main API: ${WFCENTRAL_API_PORT} (Device management endpoints)"
explain "  - Management API: ${WFCENTRAL_MGMT_PORT} (Health and readiness checks)"
echo

# Set up our demo environment
setup_demo_env

step "Creating data directory for the control plane"
mkdir -p "${DEMO_TMP_DIR}/central"

step "Starting the control plane server"
# The control plane is started in the background for the demo
wfcentral start \
    --port "${WFCENTRAL_API_PORT}" \
    --management-port "${WFCENTRAL_MGMT_PORT}" \
    --data-dir "${DEMO_TMP_DIR}/central" \
    --log-level info &

WFCENTRAL_PID=$!
echo "${WFCENTRAL_PID}" > "${DEMO_TMP_DIR}/central.pid"

step "Waiting for server to be ready"
# We use the management API's readiness endpoint to verify the server is up
for i in {1..30}; do
    if curl -s "http://localhost:${WFCENTRAL_MGMT_PORT}/readyz" | grep -q '"ready":true'; then
        success "Control plane is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        error "Control plane failed to start"
        exit 1
    fi
    sleep 1
done

step "Verifying health endpoints are accessible"
if curl -s "http://localhost:${WFCENTRAL_MGMT_PORT}/healthz" | grep -q '"status":"healthy"'; then
    success "Health check endpoint verified"
else
    error "Health check endpoint not responding"
    exit 1
fi

step "Verifying no devices are registered yet"
if [ "$(wfcentral device list | wc -l)" -eq 0 ]; then
    success "Clean server state verified"
else
    error "Unexpected devices found"
    exit 1
fi

success "Control plane initialization complete"

# Export configuration for other scripts to use
cat > "${DEMO_TMP_DIR}/wfcentral.env" << EOF
export WFCENTRAL_API_PORT=${WFCENTRAL_API_PORT}
export WFCENTRAL_MGMT_PORT=${WFCENTRAL_MGMT_PORT}
export WFCENTRAL_PID=${WFCENTRAL_PID}
export WFCENTRAL_PID_FILE="${DEMO_TMP_DIR}/central.pid"
EOF