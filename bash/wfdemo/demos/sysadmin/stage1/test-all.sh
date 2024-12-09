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
# Each service needs two ports - one for API and one for management
WFCENTRAL_API_PORT=${WFCENTRAL_API_PORT:-0}
WFCENTRAL_MGMT_PORT=${WFCENTRAL_MGMT_PORT:-0}
WFDEVICE_API_PORT=${WFDEVICE_API_PORT:-0}
WFDEVICE_MGMT_PORT=${WFDEVICE_MGMT_PORT:-0}

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
    if [ "${WFCENTRAL_API_PORT}" -eq 0 ]; then
        WFCENTRAL_API_PORT=$(get_free_port)
    fi
    if [ "${WFCENTRAL_MGMT_PORT}" -eq 0 ]; then
        WFCENTRAL_MGMT_PORT=$(get_free_port)
    fi
    if [ "${WFDEVICE_API_PORT}" -eq 0 ]; then
        WFDEVICE_API_PORT=$(get_free_port)
    fi
    if [ "${WFDEVICE_MGMT_PORT}" -eq 0 ]; then
        WFDEVICE_MGMT_PORT=$(get_free_port)
    fi
    
    mkdir -p "${TEST_OUTPUT_DIR}/central"
    mkdir -p "${TEST_OUTPUT_DIR}/device"
    
    # Export for subprocesses
    export WFCENTRAL_API_PORT WFCENTRAL_MGMT_PORT 
    export WFDEVICE_API_PORT WFDEVICE_MGMT_PORT 
    export TEST_OUTPUT_DIR
    
    log "Port allocation:"
    log "  wfcentral API port: ${WFCENTRAL_API_PORT}"
    log "  wfcentral Management port: ${WFCENTRAL_MGMT_PORT}"
    log "  wfdevice API port: ${WFDEVICE_API_PORT}"
    log "  wfdevice Management port: ${WFDEVICE_MGMT_PORT}"
}

# Test steps with timeouts and logging
test_start_central() {
    log "Starting wfcentral on ports ${WFCENTRAL_API_PORT}(API) and ${WFCENTRAL_MGMT_PORT}(Management)"
    
    run_with_timeout "${WFCENTRAL_START_TIMEOUT}" \
        wfcentral start \
            --port "${WFCENTRAL_API_PORT}" \
            --management-port "${WFCENTRAL_MGMT_PORT}" \
            --data-dir "${TEST_OUTPUT_DIR}/central" \
            --log-file "${TEST_OUTPUT_DIR}/logs/central.log" &
    echo $! > "${TEST_OUTPUT_DIR}/central.pid"
    
    # Wait for server to be ready by checking management endpoint
    local deadline=$((SECONDS + WFCENTRAL_START_TIMEOUT))
    while ! curl -s "http://localhost:${WFCENTRAL_MGMT_PORT}/readyz" | grep -q '"ready":true'; do
        if [ "${SECONDS}" -ge "${deadline}" ]; then
            fail "wfcentral failed to start within ${WFCENTRAL_START_TIMEOUT} seconds"
        fi
        sleep 1
    done
}

test_start_device() {
    log "Starting wfdevice on ports ${WFDEVICE_API_PORT}(API) and ${WFDEVICE_MGMT_PORT}(Management)"
    
    run_with_timeout "${WFDEVICE_START_TIMEOUT}" \
        wfdevice start \
            --port "${WFDEVICE_API_PORT}" \
            --management-port "${WFDEVICE_MGMT_PORT}" \
            --data-dir "${TEST_OUTPUT_DIR}/device" \
            --log-file "${TEST_OUTPUT_DIR}/logs/device.log" &
    echo $! > "${TEST_OUTPUT_DIR}/device.pid"
    
    # Wait for agent to be ready by checking management endpoint
    local deadline=$((SECONDS + WFDEVICE_START_TIMEOUT))
    while ! curl -s "http://localhost:${WFDEVICE_MGMT_PORT}/readyz" | grep -q '"ready":true'; do
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
            --port "${WFDEVICE_API_PORT}" \
            --name "test-device-${TEST_ID}" \
            --control-plane "localhost:${WFCENTRAL_API_PORT}"
            
    # Verify registration
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device list --port "${WFCENTRAL_API_PORT}" | \
        grep -q "test-device-${TEST_ID}" || \
        fail "Device registration verification failed"
}

test_configure_monitoring() {
    log "Configuring device monitoring"
    
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device monitor \
            --port "${WFCENTRAL_API_PORT}" \
            --device "test-device-${TEST_ID}" \
            --interval 30s
            
    # Verify monitoring by checking health endpoint
    run_with_timeout "${OPERATION_TIMEOUT}" \
        wfcentral device health \
            --port "${WFCENTRAL_API_PORT}" \
            --device "test-device-${TEST_ID}" | \
        grep -q "monitoring_active: true" || \
        fail "Monitoring configuration verification failed"
}

test_verify_health_endpoints() {
    log "Verifying health endpoints are properly secured"
    
    # Check central health endpoint
    if ! curl -s "http://localhost:${WFCENTRAL_MGMT_PORT}/healthz" | grep -q '"status":"healthy"'; then
        fail "Central health endpoint not responding correctly"
    fi
    
    # Check device health endpoint
    if ! curl -s "http://localhost:${WFDEVICE_MGMT_PORT}/healthz" | grep -q '"status":"healthy"'; then
        fail "Device health endpoint not responding correctly"
    fi
    
    log "Health endpoints verified successfully"
}

test_shutdown() {
    log "Initiating shutdown sequence"
    
    # Stop device
    if [ -f "${TEST_OUTPUT_DIR}/device.pid" ]; then
        run_with_timeout "${OPERATION_TIMEOUT}" \
            wfdevice stop --port "${WFDEVICE_API_PORT}" || \
            kill -9 "$(cat "${TEST_OUTPUT_DIR}/device.pid")" 2>/dev/null
    fi
    
    # Stop control plane
    if [ -f "${TEST_OUTPUT_DIR}/central.pid" ]; then
        run_with_timeout "${OPERATION_TIMEOUT}" \
            wfcentral stop --port "${WFCENTRAL_API_PORT}" || \
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
    test_verify_health_endpoints
    
    local duration=$((SECONDS - start_time))
    log "All tests passed in ${duration} seconds"
    
    # Generate JUnit-style XML report
    cat > "${TEST_OUTPUT_DIR}/junit.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="stage1" tests="5" failures="0" time="${duration}">
    <testcase name="start_central" time="${WFCENTRAL_START_TIMEOUT}" />
    <testcase name="start_device" time="${WFDEVICE_START_TIMEOUT}" />
    <testcase name="register_device" time="${OPERATION_TIMEOUT}" />
    <testcase name="configure_monitoring" time="${OPERATION_TIMEOUT}" />
    <testcase name="verify_health_endpoints" time="${OPERATION_TIMEOUT}" />
  </testsuite>
</testsuites>
EOF
}

# Main execution
run_tests