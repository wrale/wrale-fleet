#!/usr/bin/env bash

# Stage 1 - Device Configuration
# This script demonstrates how to manage device configuration through the control
# plane. Configuration changes are validated, applied, and then verified to ensure
# they take effect properly.

# Source our common utilities
source "../../../lib/common.sh"

begin_story "System Administrator" "Stage 1" "Device Configuration"

explain "As a system administrator, I need to configure the device"
explain "This sets up basic operational parameters"
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

step "Checking current device configuration"
# This shows what settings are currently active
wfcentral device config show ${DEVICE_NAME}

step "Creating basic configuration file"
cat > "${DEMO_TMP_DIR}/device-config.yaml" << EOF
logging:
  level: info
  format: json
monitoring:
  interval: 60s
  metrics:
    - cpu
    - memory
    - disk
EOF

step "Validating configuration file"
# The validate command checks the configuration format before we try to apply it
if ! wfcentral device config validate \
    --device ${DEVICE_NAME} \
    --file "${DEMO_TMP_DIR}/device-config.yaml"; then
    error "Configuration validation failed"
    exit 1
fi

step "Applying configuration to device"
if ! wfcentral device config apply \
    --device ${DEVICE_NAME} \
    --file "${DEMO_TMP_DIR}/device-config.yaml"; then
    error "Failed to apply configuration"
    exit 1
fi

step "Verifying configuration was applied"
if wfcentral device config show ${DEVICE_NAME} | grep -q "format: json"; then
    success "Configuration changes verified"
else
    error "Configuration verification failed"
    exit 1
fi

success "Device configuration complete"