#!/usr/bin/env bash
set -euo pipefail

# test-all.sh - CI Integration Test Script
# This script verifies the complete Stage 1 functionality in a CI environment.
# Unlike all.sh which is for demonstrations, this script focuses on validation
# and produces structured output suitable for CI systems.

# Default timeouts (can be overridden via environment variables)
WFCENTRAL_START_TIMEOUT=${WFCENTRAL_START_TIMEOUT:-30}
WFMACHINE_START_TIMEOUT=${WFMACHINE_START_TIMEOUT:-30}
OPERATION_TIMEOUT=${OPERATION_TIMEOUT:-10}

# Test instance identifier (for parallel test runs)
TEST_ID=${TEST_ID:-$(head -c6 /dev/urandom | base64)}
TEST_OUTPUT_DIR=${TEST_OUTPUT_DIR:-/tmp/wrale-test-${TEST_ID}}

# Port allocation (0 means find available port)
WFCENTRAL_PORT=${WFCENTRAL_PORT:-0}
WFMACHINE_PORT=${WFMACHINE_PORT:-0}

# Logging setup
mkdir -p "${TEST_OUTPUT_DIR}/logs"
exec 1> >(tee "${TEST_OUTPUT_DIR}/logs/test.log")
exec 2> >(tee "${TEST_OUTPUT_DIR}/logs/test.err")

log() { echo "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ') $*" >&2; }
fail() { log "ERROR: $*"; exit 1; }

# Find available port
get_free_port() {
    python3 -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'
}

# Initialize test environment
test_init() {
    log "Initializing test environment ${TEST_ID}"
    
    # Allocate ports if not specified
    if [ "${WFCENTRAL_PORT}" -eq 0 ]; then
        WFCENTRAL_PORT=$(get_free_port)
    fi
    if [ "${WFMACHINE_PORT}" -eq 0 ]; then
        WFMACHINE_PORT=$(get_free_port)
    fi
    
    mkdir -p "${TEST_OUTPUT_DIR}/central"
    mkdir -p "${TEST_OUTPUT_DIR}/machine"
    
    # Export for subprocesses
    export WFCENTRAL_PORT WFMACHINE_PORT TEST_OUTPUT_DIR
}

# Test steps with timeouts and logging
test_start_central() {
    log "Starting wfcentral on port ${WFCENTRAL_PORT}"
    
    timeout "${WFCENTRAL_START_TIMEOUT}" \
        wfcentral start \
            --port "${WFCENTRAL_PORT}" \
            --data-dir "${TEST_OUTPUT_DIR}/central" \
            --log-file "${TEST_OUTPUT_DIR}/logs/central.log" &
    echo $! > "${TEST_OUTPUT_DIR}/central.pid"
    
    # Wait for server to be ready
    local deadline=$((SECONDS + WFCENTRAL_START_TIMEOUT))
    while ! wfcentral status --port "${WFCENTRAL_PORT}" | grep -q "healthy"; do
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            fail "wfcentral failed to start within ${WFCENTRAL_START_TIMEOUT} seconds"
        fi
        sleep 1
    done
}

test_start_machine() {
    log "Starting wfmachine on port ${WFMACHINE_PORT}"
    
    timeout "${WFMACHINE_START_TIMEOUT}" \
        wfmachine start \
            --port "${WFMACHINE_PORT}" \
            --data-dir "${TEST_OUTPUT_DIR}/machine" \
            --log-file "${TEST_OUTPUT_DIR}/logs/machine.log" &
    echo $! > "${TEST_OUTPUT_DIR}/machine.pid"
    
    # Wait for agent to be ready
    local deadline=$((SECONDS + WFMACHINE_START_TIMEOUT))
    while ! wfmachine status --port "${WFMACHINE_PORT}" | grep -q "ready"; do
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            fail "wfmachine failed to start within ${WFMACHINE_START_TIMEOUT} seconds"
        fi
        sleep 1
    done
}

test_register_device() {
    log "Registering device with control plane"
    
    timeout "${OPERATION_TIMEOUT}" \
        wfmachine register \
            --port "${WFMACHINE_PORT}" \
            --name "test-device-${TEST_ID}" \
            --control-plane "localhost:${WFCENTRAL_PORT}"
            
    # Verify registration
    timeout "${OPERATION_TIMEOUT}" \
        wfcentral device list --port "${WFCENTRAL_PORT}" | \
        grep -q "test-device-${TEST_ID}" || \
        fail "Device registration verification failed"
}

test_configure_monitoring() {
    log "Configuring device monitoring"
    
    timeout "${OPERATION_TIMEOUT}" \
        wfcentral device monitor \
            --port "${WFCENTRAL_PORT}" \
            --device "test-device-${TEST_ID}" \
            --interval 30s
            
    # Verify monitoring
    timeout "${OPERATION_TIMEOUT}" \
        wfcentral device health \
            --port "${WFCENTRAL_PORT}" \
            --device "test-device-${TEST_ID}" | \
        grep -q "monitoring_active: true" || \
        fail "Monitoring configuration verification failed"
}

test_shutdown() {
    log "Initiating shutdown sequence"
    
    # Stop device
    if [ -f "${TEST_OUTPUT_DIR}/machine.pid" ]; then
        timeout "${OPERATION_TIMEOUT}" \
            wfmachine stop --port "${WFMACHINE_PORT}" || \
            kill -9 "$(cat "${TEST_OUTPUT_DIR}/machine.pid")" 2>/dev/null
    fi
    
    # Stop control plane
    if [ -f "${TEST_OUTPUT_DIR}/central.pid" ]; then
        timeout "${OPERATION_TIMEOUT}" \
            wfcentral stop --port "${WFCENTRAL_PORT}" || \
            kill -9 "$(cat "${TEST_OUTPUT_DIR}/central.pid")" 2>/dev/null
    fi
}

# Run all tests
run_tests() {
    local start_time=${SECONDS}
    
    trap test_shutdown EXIT
    
    test_init
    test_start_central
    test_start_machine
    test_register_device
    test_configure_monitoring
    
    local duration=$((SECONDS - start_time))
    log "All tests passed in ${duration} seconds"
    
    # Generate JUnit-style XML report
    cat > "${TEST_OUTPUT_DIR}/junit.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="stage1" tests="4" failures="0" time="${duration}">
    <testcase name="start_central" time="${WFCENTRAL_START_TIMEOUT}" />
    <testcase name="start_machine" time="${WFMACHINE_START_TIMEOUT}" />
    <testcase name="register_device" time="${OPERATION_TIMEOUT}" />
    <testcase name="configure_monitoring" time="${OPERATION_TIMEOUT}" />
  </testsuite>
</testsuites>
EOF
}

# Main execution
run_tests