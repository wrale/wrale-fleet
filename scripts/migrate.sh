#!/usr/bin/env bash

# Exit on error, undefined vars, or pipe fails
set -euo pipefail

# Current timestamp for backup naming
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPO_ROOT=$(git rev-parse --show-toplevel)
STAGING_DIR="${REPO_ROOT}/.migration_staging_${TIMESTAMP}"
BACKUP_DIR="${REPO_ROOT}/.migration_backup_${TIMESTAMP}"

# Logging functions
log_info() {
    echo "[INFO] $1" >&2
}

log_error() {
    echo "[ERROR] $1" >&2
}

# Cleanup function registered as trap
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "Error occurred. Rolling back changes..."
        if [ -d "$STAGING_DIR" ]; then
            rm -rf "$STAGING_DIR"
        fi
        if [ -d "$BACKUP_DIR" ]; then
            log_info "Restoring from backup..."
            rsync -a --delete "${BACKUP_DIR}/" "${REPO_ROOT}/"
            rm -rf "$BACKUP_DIR"
        fi
    else
        if [ -d "$STAGING_DIR" ]; then
            rm -rf "$STAGING_DIR"
        fi
        if [ -d "$BACKUP_DIR" ]; then
            rm -rf "$BACKUP_DIR"
        fi
    fi
}

trap cleanup EXIT

# Verify Go installation and version
verify_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    local go_version
    go_version=$(go version | grep -oP "go\d+\.\d+")
    if [[ ! "$go_version" =~ ^go1\.([2-9][1-9]|[2-9][0-9]+)$ ]]; then
        log_error "Go 1.21 or higher is required"
        exit 1
    }
}

# Verify Go module
verify_go_mod() {
    local dir=$1
    if [ -f "${dir}/go.mod" ]; then
        (cd "$dir" && go mod verify) || return 1
    fi
}

# Create directory structure
create_structure() {
    log_info "Creating new directory structure in staging..."
    
    mkdir -p "${STAGING_DIR}"/{shared,metal,fleet,user,sync}
    mkdir -p "${STAGING_DIR}/metal/"{core,hw,types}
    mkdir -p "${STAGING_DIR}/fleet/"{brain,edge,sync,types}
    mkdir -p "${STAGING_DIR}/user/"{api,ui}
    mkdir -p "${STAGING_DIR}/sync/"{manager,store,types}
    mkdir -p "${STAGING_DIR}/shared/"{config,testing,tools}
}

# Move files to staging
stage_files() {
    log_info "Moving files to staging area..."
    
    # Create backup
    log_info "Creating backup..."
    rsync -a "${REPO_ROOT}/" "${BACKUP_DIR}/"
    
    # Shared components
    rsync -a "${REPO_ROOT}/shared/" "${STAGING_DIR}/shared/"
    
    # Metal layer
    rsync -a "${REPO_ROOT}/metal/" "${STAGING_DIR}/metal/"
    
    # Fleet layer
    rsync -a "${REPO_ROOT}/fleet/" "${STAGING_DIR}/fleet/"
    
    # User layer
    rsync -a "${REPO_ROOT}/user/" "${STAGING_DIR}/user/"
    
    # Sync layer
    rsync -a "${REPO_ROOT}/sync/" "${STAGING_DIR}/sync/"
    
    # Root files
    for file in LICENSE README.md go.work Makefile; do
        if [ -f "${REPO_ROOT}/$file" ]; then
            cp "${REPO_ROOT}/$file" "${STAGING_DIR}/"
        fi
    done
}

# Verify module structure
verify_modules() {
    log_info "Verifying Go modules..."
    
    local modules=("shared" "metal" "fleet" "user" "sync")
    
    for module in "${modules[@]}"; do
        if ! verify_go_mod "${STAGING_DIR}/${module}"; then
            log_error "Module verification failed for ${module}"
            return 1
        fi
    done
}

# Deploy from staging to final location
deploy_changes() {
    log_info "Deploying changes..."
    
    # Sync staging to repo root
    rsync -a --delete "${STAGING_DIR}/" "${REPO_ROOT}/"
    
    # Verify deployment
    if ! verify_modules; then
        log_error "Final verification failed"
        return 1
    fi
    
    log_info "Migration completed successfully"
}

# Main execution
main() {
    log_info "Starting migration process..."
    
    verify_go
    create_structure
    stage_files
    verify_modules
    deploy_changes
}

main