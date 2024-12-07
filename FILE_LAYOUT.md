# Wrale Fleet v1.0 Layout Guide

## 1. Top-Level Structure

```
wrale-fleet/
├── go.work              # Workspace file for module coordination
├── shared/              # Shared utilities
│   └── go.mod          # github.com/wrale/wrale-fleet/shared
├── metal/              # Hardware interaction layer
│   └── go.mod          # github.com/wrale/wrale-fleet/metal
├── fleet/              # Fleet management layer
│   └── go.mod          # github.com/wrale/wrale-fleet/fleet
├── user/               # User interface layer
│   └── go.mod          # github.com/wrale/wrale-fleet/user
└── sync/               # Synchronization services
    └── go.mod          # github.com/wrale/wrale-fleet/sync
```

## 2. Detailed Layout with Module Boundaries

### Metal Layer (github.com/wrale/wrale-fleet/metal)
```
metal/
├── go.mod
├── cmd/
│   └── metald/             # Metal daemon
├── core/                   # Core metal functionality
│   ├── server/            # Metal server implementation
│   ├── secure/            # Security monitoring
│   └── thermal/           # Thermal management
├── hw/                    # Hardware abstraction
│   ├── gpio/             # GPIO management
│   ├── power/            # Power monitoring
│   ├── secure/           # Hardware security
│   └── thermal/          # Thermal control
└── types/                # Metal layer types
```

### Fleet Layer (github.com/wrale/wrale-fleet/fleet)
```
fleet/
├── go.mod
├── cmd/
│   └── fleetd/            # Fleet daemon
├── brain/                # Central coordination
│   ├── coordinator/      # Task coordination
│   ├── device/          # Device management
│   ├── engine/          # Analysis and optimization
│   ├── service/         # Brain service
│   └── types/           # Brain types
├── edge/                # Edge management
│   ├── agent/          # Edge agent
│   ├── client/         # Client implementations
│   └── store/          # Edge state storage
├── sync/                # State synchronization
│   ├── config/         # Sync configuration
│   ├── manager/        # Sync management
│   └── resolver/       # Conflict resolution
└── types/              # Fleet-wide types
```

### User Layer (github.com/wrale/wrale-fleet/user)
```
user/
├── go.mod
├── api/                # REST API
│   ├── cmd/wrale-api/
│   ├── server/        # API server
│   ├── service/       # API services
│   └── types/         # API types
└── ui/                # Web Dashboard
    └── wrale-dashboard/
        ├── src/
        │   ├── app/
        │   ├── components/
        │   ├── services/
        │   └── types/
        └── package.json
```

### Sync Layer (github.com/wrale/wrale-fleet/sync)
```
sync/
├── go.mod
├── manager/           # Sync management
├── store/            # Sync state storage
└── types/            # Sync types
```

### Shared Layer (github.com/wrale/wrale-fleet/shared)
```
shared/
├── go.mod
├── config/           # Shared configuration
├── testing/         # Test utilities
└── tools/           # Development tools
```

## 3. Module Configuration

### Root go.work
```go
go 1.21

use (
    ./metal
    ./fleet
    ./user
    ./sync
    ./shared
)
```

### Component go.mod (example for fleet)
```go
module github.com/wrale/wrale-fleet/fleet

go 1.21

require (
    github.com/wrale/wrale-fleet/metal v0.0.0
    github.com/wrale/wrale-fleet/shared v0.0.0
    github.com/wrale/wrale-fleet/sync v0.0.0
)

replace (
    github.com/wrale/wrale-fleet/metal => ../metal
    github.com/wrale/wrale-fleet/shared => ../shared
    github.com/wrale/wrale-fleet/sync => ../sync
)
```

## 4. Import Guidelines

1. **Layer-Specific Imports**
   ```go
   // Within metal layer
   import "github.com/wrale/wrale-fleet/metal/hw/gpio"
   
   // Within fleet layer
   import "github.com/wrale/wrale-fleet/fleet/brain/engine"
   
   // Within user layer
   import "github.com/wrale/wrale-fleet/user/api/types"
   ```

2. **Cross-Layer Communication**
   ```go
   // Fleet accessing metal
   import "github.com/wrale/wrale-fleet/metal/types"
   
   // User accessing fleet
   import "github.com/wrale/wrale-fleet/fleet/types"
   ```

## 5. Build Configuration

### Makefile Organization
```makefile
# Root Makefile
include Makefiles/common.mk
include Makefiles/golang.mk

# Component-specific targets
.PHONY: metal fleet user sync

metal:
    $(MAKE) -C metal all

fleet:
    $(MAKE) -C fleet all

user:
    $(MAKE) -C user all

sync:
    $(MAKE) -C sync all
```

## 6. Development Workflow

1. **Local Development**
   ```bash
   # Initial setup
   go work init
   go work use ./metal ./fleet ./user ./sync ./shared
   
   # Building specific component
   cd fleet
   go build ./...
   
   # Running tests
   go test ./...
   ```

2. **Adding New Features**
   - Add code to appropriate component
   - Update only that component's go.mod if needed
   - No need to touch other components unless API changes

## 7. Key Benefits

1. **Clear Module Boundaries**
   - Each major component is one module
   - No nested module confusion
   - Clear import paths

2. **Simplified Dependencies**
   - Direct component relationships
   - No circular dependencies
   - Clear upgrade paths

3. **Maintained Functionality**
   - All v1.0 features preserved
   - Clean integration points
   - Clear security boundaries

4. **Improved Development Experience**
   - Faster builds
   - Clearer error messages
   - Better IDE support

5. **Better CI/CD Integration**
   - Component-level testing
   - Independent deployments
   - Clear artifact boundaries

## 8. Migration Guide

1. Move files to new structure
2. Update import paths
3. Clean up go.mod files
4. Update CI/CD configurations
5. Test each component independently
6. Verify integration points
7. Update documentation

This layout maintains all v1.0 functionality while making the codebase much more maintainable and easier to understand.
