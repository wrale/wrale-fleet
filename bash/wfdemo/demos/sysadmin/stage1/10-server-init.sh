#!/usr/bin/env bash

# Stage 1 - Server Initialization
# This script demonstrates starting the central control plane, which is the core
# server component that manages our entire fleet of devices. The control plane
# needs to be running before any devices can connect to it.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Control Plane Initialization"

explain "As a system administrator, I need to start the central control plane"
explain "This server must be running before any devices can connect"
echo

# Set up our demo environment
setup_demo_env

step "Creating data directory for the control plane"
mkdir -p "${DEMO_TMP_DIR}/central"

step "Starting the control plane server"
# The --foreground flag is omitted since we want to run in background for the demo
wfcentral start \
    --port 8080 \
    --data-dir "${DEMO_TMP_DIR}/central" \
    --log-level info &

WFCENTRAL_PID=$!
echo "${WFCENTRAL_PID}" > "${DEMO_TMP_DIR}/central.pid"

step "Waiting for server to be ready"
# We use the status command to verify the server is up and responding
for i in {1..30}; do
    if wfcentral status | grep -q "healthy"; then
        success "Control plane is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        error "Control plane failed to start"
        exit 1
    fi
    sleep 1
done

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
export WFCENTRAL_PORT=8080
export WFCENTRAL_PID=${WFCENTRAL_PID}
export WFCENTRAL_PID_FILE="${DEMO_TMP_DIR}/central.pid"
EOF