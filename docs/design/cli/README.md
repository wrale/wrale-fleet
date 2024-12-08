# CLI Design

## Introduction

The Wrale Fleet Platform provides two core command-line tools that evolve through carefully designed stages. These tools serve both enterprise fleet management needs and comprehensive system demonstration, with each tool having distinct responsibilities in the overall architecture.

## Core Tools

### wfcentral
The control plane command providing global fleet management capabilities:
- Device registration and inventory
- Configuration management
- Monitoring and health checks
- Multi-region coordination
- Enterprise security features

### wfdevice
The device agent command providing local device management:
- Local device operations
- Status reporting
- Configuration application
- Health monitoring
- Secure communication with control plane

## Command Evolution

The command-line tools grow in capability through six distinct stages, with each stage building upon previous functionality while maintaining backward compatibility.

### Stage 1: Basic Device Management

#### wfcentral
```bash
# Server Lifecycle
wfcentral start              
    --port PORT              # Server port (default: 8080)
    --data-dir DIR          # Data directory (default: /var/lib/wfcentral)
    --log-level LEVEL       # Logging level (default: info)

wfcentral stop              # Stop control plane gracefully
wfcentral status            # Show server status and health

# Device Management
wfcentral device list       # List all registered devices
wfcentral device status NAME  # Show device status
wfcentral device health NAME  # Show device health metrics

# Configuration Management
wfcentral device config show NAME     # Show current configuration
wfcentral device config validate NAME # Validate configuration file
wfcentral device config apply NAME    # Apply new configuration
```

#### wfdevice
```bash
# Agent Lifecycle
wfdevice start
    --port PORT             # Agent port (default: 9090)
    --data-dir DIR         # Data directory
    --log-level LEVEL      # Logging level (default: info)

wfdevice stop             # Stop agent gracefully
wfdevice status           # Show agent status

# Registration
wfdevice register
    --name NAME           # Device name
    --control-plane HOST  # Control plane address
    --tags KEY=VALUE      # Device tags (multiple allowed)

wfdevice notify-shutdown  # Signal planned shutdown
```

### Stage 2: Multi-Site Operations

#### wfcentral
```bash
# Cluster Management
wfcentral cluster init             # Initialize cluster
wfcentral cluster join NODE        # Join existing cluster
wfcentral cluster status           # Show cluster health
wfcentral cluster members          # List cluster members

# Group Management
wfcentral group create NAME        # Create device group
wfcentral group list               # List all groups
wfcentral group add NAME DEVICE    # Add device to group
wfcentral group deploy NAME CONFIG # Deploy config to group
```

#### wfdevice
```bash
# Site Operations
wfdevice site status              # Show site connectivity
wfdevice site sync                # Sync with local site
wfdevice site failover            # Switch to alternate site
```

### Stage 3: Regional Operations

#### wfcentral
```bash
# Regional Management
wfcentral region list              # List all regions
wfcentral region status NAME       # Show region status
wfcentral region create NAME       # Create new region

# Regional Deployment
wfcentral deploy
    --region NAME                 # Target region
    --config FILE                # Configuration file
    --rollout STRATEGY           # Rollout strategy
```

#### wfdevice
```bash
# Regional Operations
wfdevice region set NAME          # Set device region
wfdevice region status            # Show regional status
wfdevice region metrics           # Report regional metrics
```

### Stage 4: Trust Relationships

#### wfcentral
```bash
# Trust Management
wfcentral trust establish NODE     # Establish trust relationship
wfcentral trust list               # List trust relationships
wfcentral trust verify NODE        # Verify trust status

# Security Operations
wfcentral security audit           # Run security audit
wfcentral security scan            # Scan for vulnerabilities
wfcentral certs rotate             # Rotate certificates
```

#### wfdevice
```bash
# Trust Operations
wfdevice trust status             # Show trust status
wfdevice trust verify             # Verify trust chain
wfdevice certs renew              # Renew certificates
```

### Stage 5: Advanced Mesh Operations

#### wfcentral
```bash
# Mesh Management
wfcentral mesh init                # Initialize mesh network
wfcentral mesh join                # Join existing mesh
wfcentral mesh status              # Show mesh status
wfcentral mesh topology            # Display mesh topology

# Performance Operations
wfcentral route optimize           # Optimize routing
wfcentral latency measure NODE     # Measure latency
```

#### wfdevice
```bash
# Mesh Operations
wfdevice mesh connect             # Connect to mesh
wfdevice mesh peers               # List mesh peers
wfdevice mesh metrics             # Report mesh metrics
wfdevice route update             # Update routing table
```

### Stage 6: Enterprise Features

#### wfcentral
```bash
# Enterprise Management
wfcentral enterprise audit         # Run enterprise audit
wfcentral compliance verify        # Verify compliance status
wfcentral maintenance schedule     # Schedule maintenance

# Integration Management
wfcentral integrate
    --system NAME                 # Target system
    --config FILE                # Integration config
    --validate                   # Validate only
```

#### wfdevice
```bash
# Enterprise Operations
wfdevice compliance check         # Check compliance
wfdevice backup create            # Create device backup
wfdevice maintenance prepare      # Prepare for maintenance
wfdevice diagnose                 # Run diagnostics
```

## Command Design Principles

1. Consistent Structure
   - Common pattern for lifecycle commands
   - Similar flags across both tools
   - Consistent help formatting
   - Progressive feature revelation

2. Clear Separation
   - Control plane operations in wfcentral
   - Device operations in wfdevice
   - Explicit boundaries between stages
   - Clear capability progression

3. Error Handling
   - Descriptive error messages
   - Clear status reporting
   - Graceful degradation
   - Stage-aware capabilities

4. Documentation
   - Built-in help for all commands
   - Examples in usage text
   - Clear stage requirements
   - Deprecation notices when needed

## Success Criteria

A CLI design is considered successful when:
- Commands are intuitive and consistent
- Help text is clear and comprehensive
- Error messages are actionable
- Staged evolution is clear
- Enterprise operations are streamlined
- Demo scenarios are supported effectively

## Future Considerations

Future CLI enhancements may include:
- Advanced automation capabilities
- Enhanced scripting support
- Extended enterprise integrations
- Advanced mesh operations
- Extended offline capabilities
