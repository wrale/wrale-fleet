# Base template for all components
include $(MAKEFILES_DIR)/common.mk

# Define the base template
define BASE_TEMPLATE
.PHONY: all test build help

all: clean verify build ## Clean, verify and build

help: ## Show available targets
	$(HELP_FUNCTION)

endef