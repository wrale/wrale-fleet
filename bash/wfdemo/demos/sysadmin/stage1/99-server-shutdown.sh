#!/usr/bin/env bash

# Stage 1 - Server Shutdown
# This script demonstrates how to properly shut down the control plane server.
# It's important to ensure all devices are disconnected before stopping the server
# to prevent any orphaned or inconsistent states.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Control Plane Shutdown"

explain "As a system administrator, I need to gracefully stop the control plane"
explain "This should only happen after all devices are shut down"
echo

# Set up our demo environment
setup_demo_env

step "Loading server configuration"
if [[ -f "${DEMO_TMP_DIR}/wfcentral.env" ]]; then
    source "${DEMO_TMP_DIR}/wfcentral.env"
else
    error "Server configuration not found"
    exit 1
fi

step "Checking for active devices"
if wfcentral device list | grep -q "online"; then
    error "There are still active devices connected"
    error "Please run 90-device-shutdown.sh first"
    exit 1
fi

step "Preparing for shutdown"
# Disable new connections
if ! wfcentral maintenance enable --message "Server shutting down"; then
    warn "Failed to enable maintenance mode"
fi

step "Stopping control plane"
if wfcentral stop; then
    success "Control plane stopped cleanly"
else
    error "Failed to stop control plane gracefully"
    
    if [[ -f "${WFCENTRAL_PID_FILE}" ]]; then
        warn "Attempting forced shutdown"
        WFCENTRAL_PID=$(cat "${WFCENTRAL_PID_FILE}")
        kill -9 "${WFCENTRAL_PID}" 2>/dev/null
        rm -f "${WFCENTRAL_PID_FILE}"
    fi
fi

success "Control plane shutdown complete"

# Clean up all environment files
rm -f "${DEMO_TMP_DIR}"/*.env