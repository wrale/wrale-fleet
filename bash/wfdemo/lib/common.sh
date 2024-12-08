#!/usr/bin/env bash

# Common utilities for wfdemo story demonstrations
# Each story exists at the intersection of:
# - A stage (what capabilities are available)
# - A persona (who is using the system)
# - A story (what they're trying to accomplish)

# Colorized output for clear demo steps
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Story framework functions
function begin_story() {
    local persona="$1"
    local stage="$2"
    local story="$3"
    
    echo -e "${BLUE}=== Story: ${story} ===${NC}"
    echo -e "${BLUE}Persona: ${persona}${NC}"
    echo -e "${BLUE}Stage: ${stage}${NC}"
    echo
}

function step() {
    local message="$1"
    echo -e "${YELLOW}→ ${message}${NC}"
}

function explain() {
    local message="$1"
    echo -e "${BLUE}  ${message}${NC}"
}

# Basic output functions
function log() { echo -e "${BLUE}[INFO]${NC} $*" >&2; }
function warn() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
function error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
function success() { echo -e "${GREEN}[SUCCESS]${NC} $*" >&2; }

# Verification helpers
function verify_port_available() {
    local port="$1"
    if lsof -i ":${port}" >/dev/null 2>&1; then
        error "Port ${port} is already in use"
        return 1
    fi
    return 0
}

function wait_for_service() {
    local name="$1"
    local port="$2"
    local max_attempts="${3:-30}"
    local attempt=1
    
    while ! curl -s "http://localhost:${port}/health" >/dev/null 2>&1; do
        if ((attempt >= max_attempts)); then
            error "${name} failed to start after ${max_attempts} attempts"
            return 1
        fi
        log "Waiting for ${name} to start (attempt ${attempt}/${max_attempts})..."
        sleep 1
        ((attempt++))
    done
    success "${name} is ready"
}

# Clean up helpers for story completion
function cleanup_service() {
    local pid_file="$1"
    if [[ -f "${pid_file}" ]]; then
        local pid
        pid=$(cat "${pid_file}")
        if kill -0 "${pid}" 2>/dev/null; then
            kill "${pid}"
            rm -f "${pid_file}"
        fi
    fi
}

# Temporary directory management
DEMO_TMP_DIR=""

function setup_demo_env() {
    DEMO_TMP_DIR=$(mktemp -d)
    export DEMO_TMP_DIR
}

function cleanup_demo_env() {
    if [[ -n "${DEMO_TMP_DIR}" && -d "${DEMO_TMP_DIR}" ]]; then
        rm -rf "${DEMO_TMP_DIR}"
    fi
}

# Ensure cleanup runs even if script fails
trap cleanup_demo_env EXIT