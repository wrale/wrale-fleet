#!/usr/bin/env bash

# Stage 1 - Complete System Lifecycle
# This script demonstrates the entire lifecycle of our system by running all
# Stage 1 operations in sequence. It acts as both a comprehensive demo and
# a test of the complete system.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Complete System Lifecycle"

explain "This demonstration will show the complete lifecycle:"
explain "1. Start the control plane"
explain "2. Initialize a device"
explain "3. Set up monitoring"
explain "4. Configure the device"
explain "5. Gracefully shut down"
echo

# Keep track of where we succeeded and failed
CURRENT_STEP=""
declare -A STEP_STATUS
STEPS=(
    "10-server-init.sh"
    "20-device-init.sh"
    "30-device-monitor.sh"
    "40-device-config.sh"
    "90-device-shutdown.sh"
    "99-server-shutdown.sh"
)

# Error handler
handle_error() {
    error "Failed during: ${CURRENT_STEP}"
    error "Rolling back any completed steps..."
    
    # If we failed after device init but before device shutdown
    if [[ " ${!STEP_STATUS[@]} " =~ "20-device-init.sh" ]] && \
       [[ ! " ${!STEP_STATUS[@]} " =~ "90-device-shutdown.sh" ]]; then
        warn "Running device shutdown to clean up"
        ./90-device-shutdown.sh
    fi
    
    # If we started the server but haven't shut it down
    if [[ " ${!STEP_STATUS[@]} " =~ "10-server-init.sh" ]] && \
       [[ ! " ${!STEP_STATUS[@]} " =~ "99-server-shutdown.sh" ]]; then
        warn "Running server shutdown to clean up"
        ./99-server-shutdown.sh
    fi
    
    exit 1
}

trap handle_error ERR

# Run each step in sequence
for script in "${STEPS[@]}"; do
    CURRENT_STEP="${script}"
    
    step "Running ${script}"
    if ! ./"${script}"; then
        error "Step failed: ${script}"
        handle_error
    fi
    STEP_STATUS["${script}"]=success
    echo # Add spacing between steps
done

success "Complete lifecycle demonstration successful"
log "All steps completed in sequence:"
for script in "${STEPS[@]}"; do
    if [[ "${STEP_STATUS[$script]}" == "success" ]]; then
        success "✓ ${script}"
    fi
done