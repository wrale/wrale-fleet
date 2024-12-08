#!/usr/bin/env bash

# Basic logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

function log() { echo -e "${BLUE}[INFO]${NC} $*" >&2; }
function warn() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
function error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
function success() { echo -e "${GREEN}[SUCCESS]${NC} $*" >&2; }

# Basic validation
function validate_name() {
    local name="$1"
    if [[ ! "$name" =~ ^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$ ]]; then
        error "Invalid name: $name"
        error "Names must contain only letters, numbers, and hyphens"
        error "They must start and end with alphanumeric characters"
        return 1
    fi
}
