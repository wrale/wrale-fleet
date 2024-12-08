#!/usr/bin/env bash

# Stage 1 - Device Monitoring
# Sets up basic monitoring for the device

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Device Monitoring"

explain "As a system administrator, I need to monitor device health"
explain "This ensures we can detect and respond to issues"
echo

# Set up our demo environment
setup_demo_env

step "Loading system configuration"
for env_file in "${DEMO_TMP_DIR}/wfcentral.env" "${DEMO_TMP_DIR}/wfdevice.env"; do
    if [[ -f "${env_file}" ]]; then
        source "${env_file}"
    else
        error "Missing configuration. Run previous initialization scripts first."
        exit 1
    fi
done

step "Verifying device status"
if ! wfcentral device status ${DEVICE_NAME}; then
    error "Device not found or not responding"
    exit 1
fi

step "Enabling device monitoring"
if ! wfcentral device monitor ${DEVICE_NAME} \
        --interval 30s \
        --metrics cpu,memory,disk; then
    error "Failed to enable monitoring"
    exit 1
fi

step "Verifying monitoring is active"
if wfcentral device health ${DEVICE_NAME} | grep -q "monitoring_active: true"; then
    success "Monitoring is properly configured"
else
    error "Monitoring verification failed"
    exit 1
fi

success "Device monitoring enabled"

# Export monitoring configuration
cat > "${DEMO_TMP_DIR}/monitoring.env" << EOF
export MONITORING_ENABLED=true
export HEALTH_CHECK_INTERVAL=30
EOF