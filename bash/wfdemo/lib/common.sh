#!/usr/bin/env bash

# Common functions and utilities for the wfdemo tool

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
function log() {
    echo -e "${BLUE}[INFO]${NC} $*" >&2
}

function warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
}

function error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

function success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" >&2
}

# Validation helpers
function validate_device_name() {
    local name="$1"
    if [[ ! "$name" =~ ^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$ ]]; then
        error "Invalid device name: $name"
        error "Device names must contain only letters, numbers, and hyphens"
        error "They must start and end with alphanumeric characters"
        return 1
    fi
}

# Demo scenario handling
function run_demo_scenario() {
    local persona="$1"
    shift
    local scenario="$1"
    shift

    local demo_file="${SCRIPT_DIR}/demos/${persona}/${scenario}.sh"
    if [[ ! -f "$demo_file" ]]; then
        error "Demo scenario not found: ${persona}/${scenario}"
        return 1
    fi

    # Source demo-magic for interactive demos if requested
    if [[ "${1:-}" == "--interactive" ]]; then
        source "${LIB_DIR}/demo-magic.sh"
        export DEMO_MAGIC_INTERACTIVE=true
    fi

    source "$demo_file"
}

# Common fleet management operations
function fleet_command() {
    # This is a placeholder for actual fleet management commands
    # In a real implementation, this would interact with the fleet management system
    echo "Executing: $*"
    # Simulate command execution
    sleep 1
    return 0
}
