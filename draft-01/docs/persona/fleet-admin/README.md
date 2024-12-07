# Fleet Administrator Guide

## Role Overview
As a Fleet Administrator, you manage the overall fleet operations, configuration, and optimization. Your focus is on fleet-wide policies, resource allocation, and system health.

## Make Targets for Fleet Administration

As a Fleet Administrator, you'll primarily use these make targets for system management:

### Core Operations
- `make all` - Build and verify all components
- `make verify` - Run system-wide verification
- `make deploy` - Deploy fleet components

### Component Management
- `make fleet` - Build fleet service
- `make metal/core` - Build metal daemon
- `make user/api` - Build API service
- `make user/ui/wrale-dashboard` - Build UI dashboard

### Verification
- `make verify-security` - Run security checks
- `make verify-deps` - Verify dependencies
- `make verify-package` - Verify package contents

Run `make help` to see all available targets.

[Previous content remains unchanged...]