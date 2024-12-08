#!/usr/bin/env bash

# Stage 1 demo: Basic device management
# This demo showcases core device management capabilities

# If running interactively with demo-magic
if [[ "${DEMO_MAGIC_INTERACTIVE:-}" == "true" ]]; then
    # Configure demo-magic
    TYPE_SPEED=20
    DEMO_PROMPT="wfdemo $ "
    
    clear
    
    p "# Welcome to the System Administrator demo - Stage 1"
    p "# We'll demonstrate basic device management capabilities"
    wait
    
    p "# First, let's register a new device"
    pe "wfdemo sysadmin device register demo-device-1"
    wait
    
    p "# Now let's check its status"
    pe "wfdemo sysadmin device status demo-device-1"
    wait
    
    p "# Let's look at the device health metrics"
    pe "wfdemo sysadmin device health demo-device-1"
    wait
    
    p "# Finally, let's examine its configuration"
    pe "wfdemo sysadmin device config get demo-device-1"
    wait
    
    p "# Demo complete! You've seen the basics of device management"
    
else
    # Non-interactive execution
    log "Starting Stage 1 demo: Basic device management"
    
    log "Registering demo device"
    wfdemo sysadmin device register demo-device-1 || exit 1
    
    log "Checking device status"
    wfdemo sysadmin device status demo-device-1 || exit 1
    
    log "Checking device health"
    wfdemo sysadmin device health demo-device-1 || exit 1
    
    log "Getting device configuration"
    wfdemo sysadmin device config get demo-device-1 || exit 1
    
    success "Stage 1 demo completed successfully"
fi
