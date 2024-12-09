# Demo Strategy

## Introduction

This document outlines our strategy for maintaining a continuously working demonstration of the Wrale Fleet Management Platform throughout its development cycle. Through carefully structured demos, we validate functionality, showcase capabilities, and provide practical examples for users.

## Core Principles

Our demonstration strategy operates at the intersection of three dimensions:

1. **Stages** - What capabilities are available
2. **Personas** - Who is using the system
3. **Stories** - What real-world tasks they need to accomplish

This three-dimensional approach ensures demos remain focused and meaningful while progressively showcasing the system's growing capabilities.

## Demo Structure

Demonstrations are organized in a clear hierarchy that reflects these intersecting concerns:

```
wfdemo/
├── demos/                    # All demonstration scenarios
│   ├── sysadmin/            # System Administrator perspectives
│   │   └── stage1/          # Stage 1 capabilities
│   │       ├── 10-server-init.sh     # Start wfcentral
│   │       ├── 20-device-init.sh     # Start wfdevice
│   │       ├── 30-device-monitor.sh  # Configure monitoring
│   │       ├── 40-device-config.sh   # Apply configuration
│   │       ├── 90-device-shutdown.sh # Stop device
│   │       └── 99-server-shutdown.sh # Stop server
│   ├── security/            # Security Team perspectives
│   └── operations/          # Operations Team perspectives
└── lib/                     # Shared demonstration utilities
    ├── common.sh            # Common functions
    └── personas/            # Persona-specific implementations
```

## Script Organization

Each demo script follows consistent patterns to be both educational and reliable:

### Naming Convention
Scripts are numbered in groups to ensure proper sequencing:
- 10-19: Infrastructure initialization
- 20-29: Basic setup and configuration
- 30-49: Core operations and management
- 90-99: Shutdown and cleanup

### Script Structure
```bash
#!/usr/bin/env bash

# Stage and persona identification
# Brief description of what this script demonstrates

# Source common utilities
source "../../../lib/common.sh"

begin_story "Persona" "Stage" "Story Name"

explain "What this story shows"
explain "Why it matters"

# Environment setup
setup_demo_env

# Main demonstration sequence
step "Starting the control plane"
command_to_run [options]

step "Verifying readiness"
verify_command

# Additional steps with verification
step "Next operation"
if ! command_to_run; then
    error "Operation failed"
    exit 1
fi

success "Story complete"
```

## Stage 1 Demonstrations

Our Stage 1 demonstrations showcase fundamental capabilities:

### Control Plane Setup (10-server-init.sh)
- Starting wfcentral
- Verifying server readiness
- Basic health monitoring

### Device Registration (20-device-init.sh)
- Starting wfdevice
- Registering with control plane
- Establishing basic connectivity

### Device Monitoring (30-device-monitor.sh)
- Setting up health checks
- Configuring metrics collection
- Verifying monitoring status

### Configuration Management (40-device-config.sh)
- Creating device configurations
- Applying settings
- Verifying configuration state

### Graceful Shutdown (90/99-shutdown.sh)
- Clean device disconnection
- Proper resource cleanup
- Server shutdown

## Test Integration

Our demonstration system serves both interactive learning and automated testing needs:

### Interactive Mode
```bash
./wfdemo run sysadmin/stage1/10-server-init.sh --interactive
```
Features:
- Step-by-step progression
- Clear explanations
- Visual feedback
- Learning opportunities

### Automated Testing
```bash
TEST_OUTPUT_DIR=/path/to/artifacts ./test-all.sh
```
Features:
- Structured output
- Clear pass/fail status
- Detailed logging
- CI/CD integration

## Best Practices

When writing demonstration scripts:

1. Always verify command success:
```bash
if ! wfcentral status | grep "healthy"; then
    error "Server health check failed"
    exit 1
fi
```

2. Use clear, descriptive step messages:
```bash
step "Registering device with control plane"
```

3. Provide context through explanations:
```bash
explain "We verify server health before device registration"
explain "This ensures reliable device connections"
```

4. Clean up resources reliably:
```bash
trap cleanup_demo_env EXIT
```

## Success Criteria

A demonstration is considered successful when:
- All steps complete without error
- Expected state changes are verified
- Resources are properly cleaned up
- Clear educational value is provided
- Both interactive and automated modes work
- New users can understand and follow along

## Future Evolution

As we develop additional stages, demonstrations will expand to showcase:
- Multi-site operations
- Regional deployments
- Trust relationships
- Mesh networking
- Enterprise features

Each new capability will be demonstrated through our three-dimensional approach:
1. The stage that enables it
2. The persona who needs it
3. The story that shows its value

## Educational Value

The demonstration system should:
1. Teach through example
2. Show best practices
3. Illustrate common patterns
4. Handle error cases gracefully
5. Provide clear explanations

## Maintenance Requirements

To keep demonstrations working through development:
1. Run demos as part of CI/CD
2. Update when interfaces change
3. Maintain compatibility with all stages
4. Keep documentation current
5. Test both interactive and automated modes
