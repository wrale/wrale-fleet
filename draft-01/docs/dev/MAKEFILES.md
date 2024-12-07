# Wrale Fleet Build System

## Overview

The Wrale Fleet build system uses a hierarchical, template-based Make system designed for maintainability, reusability, and extensibility. It supports multiple languages, platforms, and build configurations while providing consistent behavior across all components.

## Directory Structure

```
Makefiles/
├── core/                 # Core functionality
│   ├── vars.mk          # Common variables
│   ├── targets.mk       # Base targets
│   ├── docker.mk        # Container support
│   ├── verify.mk        # Verification rules
│   └── helpers.mk       # Helper functions
├── impl/                # Language implementations
│   ├── go.mk           # Go build rules
│   ├── node.mk         # Node.js build rules
│   └── python.mk       # Python build rules
├── templates/           # Component templates
│   ├── base/           # Base templates
│   │   ├── component.mk # Base component
│   │   ├── service.mk   # Base service
│   │   └── tool.mk      # Base tool
│   ├── go/             # Go templates
│   │   ├── service.mk   # Go service
│   │   ├── daemon.mk    # Go daemon
│   │   └── agent.mk     # Go agent
│   └── ui/             # UI templates
│       ├── next.mk      # Next.js template
│       └── react.mk     # React template
└── mixins/             # Optional features
    ├── docker.mk       # Docker support
    ├── testing.mk      # Advanced testing
    └── monitoring.mk   # Monitoring support
```

## Core System

### Helper Functions (helpers.mk)

The foundation of the build system, providing essential functions for:
- Logging and output formatting
- Variable validation
- File and directory operations
- Platform detection
- Command availability checking
- Path manipulation

Example:
```make
$(call log_info, "Building component...")
$(call check_var, VERSION)
$(call ensure_dir, $(BUILD_DIR))
```

### Variables (vars.mk)

Centralizes common variable definitions:
- Build parameters
- Version information
- Tool paths
- Default configurations

### Targets (targets.mk)

Defines standard targets that all components support:
- build: Build the component
- clean: Clean build artifacts
- test: Run tests
- verify: Run verifications
- all: Full build cycle

## Language Implementations

Language-specific build rules reside in `impl/`:

### Go Implementation (go.mk)
- Go-specific build flags
- Test configuration
- Coverage reporting
- Linting settings

### Node.js Implementation (node.mk)
- NPM/Yarn handling
- Frontend build process
- Asset compilation
- Development server

### Python Implementation (python.mk)
- Virtual environment management
- Package installation
- Testing framework
- Distribution building

## Template System

### Base Templates

1. **component.mk**
   - Basic component structure
   - Common operations
   - Help system

2. **service.mk**
   - Service lifecycle management
   - Health checking
   - Logging configuration

3. **tool.mk**
   - CLI tool building
   - Installation management
   - Version handling

### Language-Specific Templates

Built on base templates, adding language-specific features:

1. **Go Templates**
   ```make
   # Example service
   include $(MAKEFILES_DIR)/templates/go/service.mk
   
   COMPONENT_NAME := my-service
   MAIN_PACKAGE := ./cmd/server
   
   $(eval $(GO_SERVICE_TEMPLATE))
   ```

2. **UI Templates**
   ```make
   # Example Next.js app
   include $(MAKEFILES_DIR)/templates/ui/next.mk
   
   COMPONENT_NAME := dashboard
   
   $(eval $(NEXTJS_TEMPLATE))
   ```

## Mixins

Optional features that can be added to any component:

### Docker Support
```make
include $(MAKEFILES_DIR)/mixins/docker.mk

# Adds targets:
# - docker-build
# - docker-push
# - docker-run
```

### Advanced Testing
```make
include $(MAKEFILES_DIR)/mixins/testing.mk

# Adds targets:
# - test-integration
# - test-performance
# - test-coverage
```

## Usage Examples

### Creating a New Go Service
```make
# service/Makefile
include $(MAKEFILES_DIR)/templates/go/service.mk

COMPONENT_NAME := my-service
COMPONENT_DESCRIPTION := My Example Service
MAIN_PACKAGE := ./cmd/server

# Optional docker support
include $(MAKEFILES_DIR)/mixins/docker.mk

$(eval $(GO_SERVICE_TEMPLATE))
```

### Creating a UI Component
```make
# ui/dashboard/Makefile
include $(MAKEFILES_DIR)/templates/ui/next.mk

COMPONENT_NAME := dashboard
COMPONENT_DESCRIPTION := Admin Dashboard

# Optional monitoring
include $(MAKEFILES_DIR)/mixins/monitoring.mk

$(eval $(NEXTJS_TEMPLATE))
```

## Best Practices

1. **Template Selection**
   - Use the most specific template available
   - Mix in optional features as needed
   - Don't override template behavior unnecessarily

2. **Variable Handling**
   - Define component-specific variables at the top
   - Use conditionals for optional features
   - Override defaults purposefully

3. **Custom Targets**
   - Add component-specific targets after template evaluation
   - Use helper functions for consistency
   - Document custom targets

4. **Error Handling**
   - Use provided validation functions
   - Add appropriate error messages
   - Fail fast on critical errors

## Extending the System

### Adding New Templates
1. Create template file in appropriate directory
2. Include base template
3. Define template-specific variables
4. Define template macro
5. Document usage

### Creating New Mixins
1. Create mixin file in mixins/
2. Define mixin-specific variables
3. Add mixin targets
4. Document features and requirements

## Common Issues

1. **Template Inheritance**
   - Ensure proper include order
   - Check variable definitions
   - Verify template evaluation

2. **Build Failures**
   - Check required variables
   - Verify tool availability
   - Review component configuration

3. **Dependency Issues**
   - Validate required tools
   - Check component dependencies
   - Verify mixin compatibility

## Contributing

1. Follow the existing structure
2. Use helper functions consistently
3. Document new features
4. Add tests for new functionality
5. Update this documentation
