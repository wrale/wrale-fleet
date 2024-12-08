#!/usr/bin/env bash

# System Administrator persona implementation

# Command handler for sysadmin persona
function handle_sysadmin_command() {
    local command="$1"
    shift

    case "$command" in
        device)
            handle_device_command "$@"
            ;;
        config)
            handle_config_command "$@"
            ;;
        *)
            error "Unknown sysadmin command: $command"
            return 1
            ;;
    esac
}

# Handle device-related commands
function handle_device_command() {
    local operation="$1"
    shift

    case "$operation" in
        register)
            register_device "$@"
            ;;
        status)
            device_status "$@"
            ;;
        health)
            device_health "$@"
            ;;
        config)
            device_config "$@"
            ;;
        *)
            error "Unknown device operation: $operation"
            echo "Available operations: register, status, health, config"
            return 1
            ;;
    esac
}

# Device management operations
function register_device() {
    local name="$1"
    if [[ -z "$name" ]]; then
        error "Device name required"
        return 1
    fi

    validate_device_name "$name" || return 1
    
    log "Registering device: $name"
    if fleet_command device register "$name"; then
        success "Device registered successfully: $name"
        return 0
    else
        error "Failed to register device: $name"
        return 1
    fi
}

function device_status() {
    local name="$1"
    if [[ -z "$name" ]]; then
        error "Device name required"
        return 1
    fi

    validate_device_name "$name" || return 1
    
    log "Checking device status: $name"
    fleet_command device status "$name"
}

function device_health() {
    local name="$1"
    if [[ -z "$name" ]]; then
        error "Device name required"
        return 1
    fi

    validate_device_name "$name" || return 1
    
    log "Checking device health: $name"
    fleet_command device health "$name"
}

function device_config() {
    local operation="$1"
    shift
    local name="$1"
    
    if [[ -z "$name" ]]; then
        error "Device name required"
        return 1
    fi

    validate_device_name "$name" || return 1

    case "$operation" in
        get)
            log "Getting device configuration: $name"
            fleet_command device config get "$name"
            ;;
        set)
            shift
            local config_file="$1"
            if [[ -z "$config_file" ]]; then
                error "Config file required"
                return 1
            fi
            if [[ ! -f "$config_file" ]]; then
                error "Config file not found: $config_file"
                return 1
            fi
            log "Setting device configuration: $name"
            fleet_command device config set "$name" --file "$config_file"
            ;;
        *)
            error "Unknown config operation: $operation"
            echo "Available operations: get, set"
            return 1
            ;;
    esac
}

# Handle configuration-related commands
function handle_config_command() {
    local operation="$1"
    shift

    case "$operation" in
        list)
            list_configs "$@"
            ;;
        deploy)
            deploy_config "$@"
            ;;
        *)
            error "Unknown config operation: $operation"
            echo "Available operations: list, deploy"
            return 1
            ;;
    esac
}
