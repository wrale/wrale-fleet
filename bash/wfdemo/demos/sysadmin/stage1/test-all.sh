#!/usr/bin/env bash
set -euo pipefail

# test-all.sh - CI Integration Test Script
# This script verifies the complete Stage 1 functionality in a CI environment.
# Unlike all.sh which is for demonstrations, this script focuses on validation
# and produces structured output suitable for CI systems.

# Default timeouts (can be overridden via environment variables)
WFCENTRAL_START_TIMEOUT=${WFCENTRAL_START_TIMEOUT:-30}
WFDEVICE_START_TIMEOUT=${WFDEVICE_START_TIMEOUT:-30}
OPERATION_TIMEOUT=${OPERATION_TIMEOUT:-10}

# Test instance identifier (for parallel test runs)
TEST_ID=${TEST_ID:-$(head -c6 /dev/urandom | base64)}
TEST_OUTPUT_DIR=${TEST_OUTPUT_DIR:-/tmp/wrale-test-${TEST_ID}}

# Port allocation (0 means find available port)
WFCENTRAL_PORT=${WFCENTRAL_PORT:-0}
WFDEVICE_PORT=${WFDEVICE_PORT:-0}

# Logging setup
mkdir -p "${TEST_OUTPUT_DIR}/logs"
exec 1> >(tee "${TEST_OUTPUT_DIR}/logs/test.log")
exec 2> >(tee "${TEST_OUTPUT_DIR}/logs/test.err")

log() { echo "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ') $*" >&2; }
fail() { log "ERROR: $*"; exit 1; }

# Cross-platform timeout implementation
run_with_timeout() {
    local timeout=$1
    shift
    
    # Run the command in background
    ("$@") & local pid=$!
    
    # Start a timer in background
    (
        sleep "$timeout"
        # If process is still running after timeout
        if kill -0 $pid 2>/dev/null; then
            kill -TERM $pid 2>/dev/null || kill -9 $pid 2>/dev/null
            fail "Command timed out after ${timeout} seconds: $*"
        fi
    ) & local timer_pid=$!
    
    # Wait for command to finish
    if wait $pid 2>/dev/null; then
        kill -TERM $timer_pid 2>/dev/null || kill -9 $timer_pid 2>/dev/null
        wait $timer_pid 2>/dev/null || true
        return 0
    else
        local ret=$?
        kill -TERM $timer_pid 2>/dev/null || kill -9 $timer_pid 2>/dev/null
        wait $timer_pid 2>/dev/null || true
        return $ret
    fi
}

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
    if [ "${WFDEVICE_PORT}" -eq 0 ]; then
        WFDEVICE_PORT=$(get_free_port)
    fi
    
    mkdir -p "${TEST_OUTPUT_DIR}/central"
    mkdir -p "${TEST_OUTPUT_DIR}/device"
    
    # Export for subprocesses
    export WFCENTRAL_PORT WFDEVICE_PORT TEST_OUTPUT_DIR
}

# Test steps with timeouts and logging
test_start_central() {
    log "Starting wfcentral on port ${WFCENTRAL_PORT}"
    
    run_with_timeout "${WFCENTRAL_START_TIMEOUT}" \
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

test_start_device() {
    log "Starting wfdevice on port ${WFDEVICE_PORT}"
    
    run_with_timeout "${WFDEVICE_START_TIMEOUT}" \
        wfdevice start \
            --port "${WFDEVICE_PORT}" \
            --data-dir "${TEST_OUTPUT_DIR}/device" \
            --log-file "${TEST_OUTPUT_DIR}/logs/device.log" &
    echo $! > "${TEST_OUTPUT_DIR}/device.pid"
    
    # Wait for agent to be ready
    local deadline=$((SECONDS + WFDEVICE_START_TIMEOUT))
    while ! wfdevice status --port "${WFDEVICE_PORT}" | grep -q "ready"; do
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            fail "wfdevice failed to start within ${WFDEVICE_START_TIMEOUT} seconds"
        fi
        sleep 1
    done
}

test_register_device() {
    log "Registering device with control plane"
    
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfdevice register \
            --port "${WFDEVICE_PORT}" \
            --name "test-device-${TEST_ID}" \
            --control-plane "localhost:${WFCENTRAL_PORT}"
            
    # Verify registration
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device list --port "${WFCENTRAL_PORT}" | \
        grep -q "test-device-${TEST_ID}" || \
        fail "Device registration verification failed"
}

test_configure_monitoring() {
    log "Configuring device monitoring"
    
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device monitor \
            --port "${WFCENTRAL_PORT}" \
            --device "test-device-${TEST_ID}" \
            --interval 30s
            
    # Verify monitoring
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device health \
            --port "${WFCENTRAL_PORT}" \
            --device "test-device-${TEST_ID}" | \
        grep -q "monitoring_active: true" || \
        fail "Monitoring configuration verification failed"
}

test_shutdown() {
    log "Initiating shutdown sequence"
    
    # Stop device
    if [ -f "${TEST_OUTPUT_DIR}/device.pid" ]; then
        run_with_timeout "${OPERATION_TIMEOUT}" \
            wfdevice stop --port "${WFDEVICE_PORT}" || \
            kill -9 "$(cat "${TEST_OUTPUT_DIR}/device.pid")" 2>/dev/null
    fi
    
    # Stop control plane
    if [ -f "${TEST_OUTPUT_DIR}/central.pid" ]; then
        run_with_timeout "${OPERATION_TIMEOUT}" \
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
    test_start_device
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
    <testcase name="start_device" time="${WFDEVICE_START_TIMEOUT}" />
    <testcase name="register_device" time="${OPERATION_TIMEOUT}" />
    <testcase name="configure_monitoring" time="${OPERATION_TIMEOUT}" />
  </testsuite>
</testsuites>
EOF
}

# Main execution
run_tests