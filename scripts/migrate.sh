#!/usr/bin/env bash

# Exit on error, undefined vars, or pipe fails
set -euo pipefail

# Script version
VERSION="1.0.0"

# Default options
DRY_RUN=0
VERBOSE=0
SKIP_GIT_CHECK=0
PROGRESS=1

# Current timestamp for backup naming
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPO_ROOT=$(git rev-parse --show-toplevel)
STAGING_DIR="${REPO_ROOT}/.migration_staging_${TIMESTAMP}"
BACKUP_DIR="${REPO_ROOT}/.migration_backup_${TIMESTAMP}"
REQUIRED_SPACE_MB=1024  # 1GB minimum free space required

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

log_progress() {
    if [[ ${PROGRESS} -eq 1 ]]; then
        echo -e "${BOLD}[PROGRESS]${NC} $1" >&2
    fi
}

# Show usage information
usage() {
    cat << EOF
Usage: $0 [options]

Safely migrate to new file layout structure as defined in FILE_LAYOUT.md

Options:
    -h, --help          Show this help message
    -n, --dry-run       Show what would be done without making changes
    -v, --verbose       Increase verbosity
    --skip-git-check    Skip git working directory check
    --no-progress      Disable progress output

Version: ${VERSION}
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -n|--dry-run)
                DRY_RUN=1
                shift
                ;;
            -v|--verbose)
                VERBOSE=1
                shift
                ;;
            --skip-git-check)
                SKIP_GIT_CHECK=1
                shift
                ;;
            --no-progress)
                PROGRESS=0
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Enhanced cleanup with multiple signal handling
cleanup() {
    local exit_code=$?
    local signal=$1

    if [[ ${signal} != "EXIT" ]]; then
        log_warning "Received signal: ${signal}"
    fi

    if [[ ${exit_code} -ne 0 ]] && [[ ${DRY_RUN} -eq 0 ]]; then
        log_error "Error occurred (exit code: ${exit_code}). Rolling back changes..."
        if [[ -d "${STAGING_DIR}" ]]; then
            rm -rf "${STAGING_DIR}"
        fi
        if [[ -d "${BACKUP_DIR}" ]]; then
            log_info "Restoring from backup..."
            rsync -a --delete "${BACKUP_DIR}/" "${REPO_ROOT}/"
            rm -rf "${BACKUP_DIR}"
        fi
    else
        if [[ ${DRY_RUN} -eq 0 ]]; then
            if [[ -d "${STAGING_DIR}" ]]; then
                rm -rf "${STAGING_DIR}"
            fi
            if [[ -d "${BACKUP_DIR}" ]]; then
                rm -rf "${BACKUP_DIR}"
            fi
        fi
    fi

    # Re-kill ourselves with the original signal
    if [[ ${signal} != "EXIT" ]]; then
        trap - ${signal}
        kill "-${signal}" "$$"
    fi
}

# Set up signal handling
for sig in HUP INT QUIT TERM; do
    trap "cleanup ${sig}" ${sig}
done
trap "cleanup EXIT" EXIT

# Verify system requirements
verify_system() {
    # Check for required commands
    local required_commands=("git" "rsync" "go" "awk" "sed")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "${cmd}" &> /dev/null; then
            log_error "Required command not found: ${cmd}"
            exit 1
        fi
    done

    # Check available disk space
    local available_mb
    if [[ "$(uname)" == "Darwin" ]]; then
        available_mb=$(df -m "${REPO_ROOT}" | awk 'NR==2 {print $4}')
    else
        available_mb=$(df -m "${REPO_ROOT}" | awk 'NR==2 {print $4}')
    fi

    if [[ ${available_mb} -lt ${REQUIRED_SPACE_MB} ]]; then
        log_error "Insufficient disk space. Required: ${REQUIRED_SPACE_MB}MB, Available: ${available_mb}MB"
        exit 1
    fi

    # Verify git repository
    if [[ ${SKIP_GIT_CHECK} -eq 0 ]]; then
        if ! git rev-parse --git-dir > /dev/null 2>&1; then
            log_error "Not a git repository"
            exit 1
        fi

        if [[ -n "$(git status --porcelain)" ]]; then
            log_error "Git working directory is not clean. Commit or stash changes first."
            exit 1
        fi
    fi
}

# Verify Go installation and version
verify_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    local go_version
    go_version=$(go version | grep -oE "go[0-9]+\.[0-9]+")
    # Verify Go 1.21 or higher
    if [[ ! "${go_version}" =~ ^go1\.([2-9][1-9]|[2-9][0-9]+)$ ]]; then
        log_error "Go 1.21 or higher is required (found ${go_version})"
        exit 1
    fi
}

# Verify Go module and its dependencies
verify_go_mod() {
    local dir=$1
    if [[ ! -f "${dir}/go.mod" ]]; then
        return 0
    fi

    if ! (cd "${dir}" && go mod verify); then
        return 1
    fi

    # Verify module dependencies if go.mod exists
    if ! (cd "${dir}" && go mod tidy -v); then
        log_warning "Module tidy check failed for ${dir}"
        return 1
    fi
}

# Create directory structure exactly matching FILE_LAYOUT.md
create_structure() {
    log_info "Creating new directory structure in staging..."
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "[DRY RUN] Would create directory structure in ${STAGING_DIR}"
        return 0
    fi

    # Create top-level directories
    mkdir -p "${STAGING_DIR}"/{shared,metal,fleet,user,sync}

    # Metal layer structure
    mkdir -p "${STAGING_DIR}/metal/"{cmd/metald,core/{server,secure,thermal},hw/{gpio,power,secure,thermal},types}

    # Fleet layer structure
    mkdir -p "${STAGING_DIR}/fleet/"{cmd/fleetd,brain/{coordinator,device,engine,service,types},edge/{agent,client,store},sync/{config,manager,resolver},types}

    # User layer structure
    mkdir -p "${STAGING_DIR}/user/"{api/{cmd/wrale-api,server,service,types},ui/wrale-dashboard/src/{app,components,services,types}}

    # Sync layer structure
    mkdir -p "${STAGING_DIR}/sync/"{manager,store,types}

    # Shared layer structure
    mkdir -p "${STAGING_DIR}/shared/"{config,testing,tools}
}

# Function to calculate and verify checksums
verify_checksums() {
    local source=$1
    local dest=$2
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        return 0
    fi

    log_progress "Verifying checksums..."
    
    local source_sum
    local dest_sum
    
    if [[ -d "${source}" ]]; then
        source_sum=$(find "${source}" -type f -exec sha256sum {} \; | sort | sha256sum)
        dest_sum=$(find "${dest}" -type f -exec sha256sum {} \; | sort | sha256sum)
    else
        source_sum=$(sha256sum "${source}" | cut -d' ' -f1)
        dest_sum=$(sha256sum "${dest}" | cut -d' ' -f1)
    fi
    
    if [[ "${source_sum}" != "${dest_sum}" ]]; then
        log_error "Checksum verification failed"
        return 1
    fi
}

# Move files to staging with progress and verification
stage_files() {
    log_info "Moving files to staging area..."
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "[DRY RUN] Would stage files to ${STAGING_DIR}"
        return 0
    fi

    # Create backup first
    log_progress "Creating backup..."
    rsync -a "${REPO_ROOT}/" "${BACKUP_DIR}/"
    
    # Stage files with progress
    local components=(shared metal fleet user sync)
    local total=${#components[@]}
    local current=0
    
    for component in "${components[@]}"; do
        ((current++))
        log_progress "Staging ${component} (${current}/${total})"
        
        rsync -a "${REPO_ROOT}/${component}/" "${STAGING_DIR}/${component}/"
        verify_checksums "${REPO_ROOT}/${component}" "${STAGING_DIR}/${component}"
    done

    # Copy root files
    log_progress "Copying root files..."
    local root_files=(LICENSE README.md go.work Makefile)
    for file in "${root_files[@]}"; do
        if [[ -f "${REPO_ROOT}/${file}" ]]; then
            cp "${REPO_ROOT}/${file}" "${STAGING_DIR}/"
            verify_checksums "${REPO_ROOT}/${file}" "${STAGING_DIR}/${file}"
        fi
    done
}

# Update go.work file
update_go_work() {
    local work_file="${STAGING_DIR}/go.work"
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "[DRY RUN] Would update go.work"
        return 0
    fi

    cat > "${work_file}" << EOF
go 1.21

use (
    ./metal
    ./fleet
    ./user
    ./sync
    ./shared
)
EOF
}

# Verify all modules
verify_modules() {
    log_info "Verifying Go modules..."
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "[DRY RUN] Would verify modules"
        return 0
    }

    local modules=(shared metal fleet user sync)
    local total=${#modules[@]}
    local current=0
    
    for module in "${modules[@]}"; do
        ((current++))
        log_progress "Verifying ${module} (${current}/${total})"
        
        if ! verify_go_mod "${STAGING_DIR}/${module}"; then
            log_error "Module verification failed for ${module}"
            return 1
        fi
    done
}

# Deploy from staging to final location
deploy_changes() {
    log_info "Deploying changes..."
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "[DRY RUN] Would deploy from ${STAGING_DIR} to ${REPO_ROOT}"
        return 0
    }

    # Sync staging to repo root with progress
    rsync -a --delete --progress "${STAGING_DIR}/" "${REPO_ROOT}/"
    
    # Final verification
    if ! verify_modules; then
        log_error "Final verification failed"
        return 1
    fi
    
    log_info "Migration completed successfully"
}

# Main execution
main() {
    parse_args "$@"
    
    if [[ ${DRY_RUN} -eq 1 ]]; then
        log_info "Running in dry-run mode. No changes will be made."
    fi

    log_info "Starting migration process..."
    
    verify_system
    verify_go
    create_structure
    stage_files
    update_go_work
    verify_modules
    deploy_changes
}

main "$@"
