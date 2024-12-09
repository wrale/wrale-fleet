#!/usr/bin/env bash

# Stage 1 - Device Initialization
# This script shows how to start a device agent and register it with the control
# plane. A device agent is the local service that runs on each managed device,
# allowing the control plane to monitor and manage it.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Device Initialization"

explain "As a system administrator, I need to start and register a device agent"
explain "This connects our first device to the control plane"
explain "We use dedicated port ranges to ensure consistent configuration:"
explain "  - Main API: ${WFDEVICE_API_PORT} (Device operations)"
explain "  - Management API: ${WFDEVICE_MGMT_PORT} (Health checks)"
echo

# Set up our demo environment
setup_demo_env

step "Loading control plane configuration"
if [[ -f "${DEMO_TMP_DIR}/wfcentral.env" ]]; then
    source "${DEMO_TMP_DIR}/wfcentral.env"
else
    error "Control plane not initialized. Run 10-server-init.sh first."
    exit 1
fi

# Validate control plane port is available
if [[ -z "${WFCENTRAL_API_PORT}" ]]; then
    error "Control plane API port not set in environment"
    exit 1
fi

step "Constructing control plane address"
CONTROL_PLANE_ADDR="localhost:${WFCENTRAL_API_PORT}"
explain "Using control plane address: ${CONTROL_PLANE_ADDR}"

step "Creating data directory for device agent"
mkdir -p "${DEMO_TMP_DIR}/device"

step "Starting device agent"
explain "Connecting to control plane at ${CONTROL_PLANE_ADDR}"
wfdevice start \
    --port ${WFDEVICE_API_PORT} \
    --management-port ${WFDEVICE_MGMT_PORT} \
    --data-dir "${DEMO_TMP_DIR}/device" \
    --log-level info \
    --control-plane "${CONTROL_PLANE_ADDR}" \
    --name "first-device" &

WFDEVICE_PID=$!
echo "${WFDEVICE_PID}" > "${DEMO_TMP_DIR}/device.pid"

step "Waiting for device agent to be ready"
# Check readiness through management port
for i in {1..30}; do
    if curl -s "http://localhost:${WFDEVICE_MGMT_PORT}/readyz" | grep -q '"ready":true'; then
        success "Device agent is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        error "Device agent failed to start"
        exit 1
    fi
    sleep 1
done

step "Verifying device status"
if wfcentral device list --port "${WFCENTRAL_API_PORT}" | grep -q "first-device"; then
    success "Device appears in control plane"
else
    error "Device not found in control plane"
    exit 1
fi

success "Device initialization complete"

# Export configuration for other scripts
cat > "${DEMO_TMP_DIR}/wfdevice.env" << EOF
export WFDEVICE_API_PORT=${WFDEVICE_API_PORT}
export WFDEVICE_MGMT_PORT=${WFDEVICE_MGMT_PORT}
export WFDEVICE_PID=${WFDEVICE_PID}
export WFDEVICE_PID_FILE="${DEMO_TMP_DIR}/device.pid"
export DEVICE_NAME="first-device"
EOF
