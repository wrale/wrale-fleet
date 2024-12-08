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

step "Creating data directory for device agent"
mkdir -p "${DEMO_TMP_DIR}/device"

step "Starting device agent"
# Note: Device name will be set during registration
wfdevice start \
    --port 9090 \
    --data-dir "${DEMO_TMP_DIR}/device" \
    --log-level info &

WFdevice_PID=$!
echo "${WFdevice_PID}" > "${DEMO_TMP_DIR}/device.pid"

step "Waiting for device agent to be ready"
for i in {1..30}; do
    if wfdevice status | grep -q "ready"; then
        success "Device agent is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        error "Device agent failed to start"
        exit 1
    fi
    sleep 1
done

step "Registering device with control plane"
if ! wfdevice register \
    --name "first-device" \
    --control-plane "localhost:${WFCENTRAL_PORT}"; then
    error "Failed to register device"
    exit 1
fi

step "Verifying registration"
if wfcentral device list | grep -q "first-device"; then
    success "Device appears in control plane"
else
    error "Device not found in control plane"
    exit 1
fi

success "Device initialization complete"

# Export configuration for other scripts
cat > "${DEMO_TMP_DIR}/wfdevice.env" << EOF
export WFdevice_PORT=9090
export WFdevice_PID=${WFdevice_PID}
export WFdevice_PID_FILE="${DEMO_TMP_DIR}/device.pid"
export DEVICE_NAME="first-device"
EOF