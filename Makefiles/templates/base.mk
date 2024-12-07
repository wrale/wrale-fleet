# Base template for all components
include $(MAKEFILES_DIR)/common.mk

# Default values for component-specific variables
COMPONENT_DESCRIPTION ?= $(COMPONENT_NAME)
HELP_PRE ?= @:  # No-op command
HELP_POST ?= @:

# Default build commands that components should override
BUILD_CMD ?= @echo "No build command defined for $(COMPONENT_NAME)"
TEST_CMD ?= @echo "No test command defined for $(COMPONENT_NAME)"
CLEAN_CMD ?= rm -rf $(BUILD_DIR) $(DIST_DIR)
VERIFY_CMD ?= @:  # No-op by default

# Define the base template
define BASE_TEMPLATE

# Template targets
.PHONY: all build test clean verify help \
        help-pre help-post

build: ## Build the component
	$(BUILD_CMD)

test: ## Run component tests
	$(TEST_CMD)

clean: ## Clean build artifacts
	$(CLEAN_CMD)

verify: ## Verify the component
	$(VERIFY_CMD)

all: clean verify build ## Clean, verify and build

# Help system hooks
help-pre:
	$(HELP_PRE)

help-post:
	$(HELP_POST)

# Main help target that components shouldn't override
help: help-pre ## Show available targets
	@echo "$(COMPONENT_NAME) - $(COMPONENT_DESCRIPTION)"
	@echo
	@echo "Available targets:"
	@echo
	$(HELP_FUNCTION)
	@if [ "$(HELP_EXTRA)" != "" ]; then \
		echo; \
		echo "$(HELP_EXTRA)"; \
	fi
	$(MAKE) help-post

endef