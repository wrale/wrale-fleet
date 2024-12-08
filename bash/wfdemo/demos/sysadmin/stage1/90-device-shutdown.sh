#!/usr/bin/env bash

# Stage 1 - Device Shutdown
# This script demonstrates how to gracefully stop a device agent. Proper shutdown
# ensures the control plane is notified and can update its state accordingly.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Device Shutdown"

explain "As a system administrator, I need to gracefully stop the device agent"
explain "This ensures clean disconnection from the control plane"
echo

# Set up our demo environment
setup_demo_env

step "Loading device configuration"
if [[ -f "${DEMO_TMP_DIR}/wfdevice.env" ]]; then
    source "${DEMO_TMP_DIR}/wfdevice.env"
else
    error "Device configuration not found"
    exit 1
fi

step "Preparing for shutdown"
# Notify the control plane we're about to stop
if ! wfdevice notify-shutdown; then
    warn "Failed to notify control plane of pending shutdown"
fi

step "Stopping device agent"
if wfdevice stop; then
    success "Device agent stopped cleanly"
else
    error "Failed to stop device agent gracefully"
    
    if [[ -f "${WFdevice_PID_FILE}" ]]; then
        warn "Attempting forced shutdown"
        WFdevice_PID=$(cat "${WFdevice_PID_FILE}")
        kill -9 "${WFdevice_PID}" 2>/dev/null
        rm -f "${WFdevice_PID_FILE}"
    fi
fi

step "Verifying device state in control plane"
if wfcentral device status ${DEVICE_NAME} | grep -q "offline"; then
    success "Control plane shows device as offline"
else
    warn "Device state may not be properly updated in control plane"
fi

success "Device shutdown complete"