# Base template for all components
include $(MAKEFILES_DIR)/common.mk

# Default values for component-specific variables
COMPONENT_DESCRIPTION ?= $(COMPONENT_NAME)
HELP_PRE ?= @:  # No-op command
HELP_POST ?= @:

# Define the base template
define BASE_TEMPLATE

# Template targets
.PHONY: all build test clean help \
        help-pre help-post

all: clean verify build ## Clean, verify and build

# Hook points for components to extend
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