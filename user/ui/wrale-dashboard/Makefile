# UI Dashboard Makefile
# Manages the Next.js frontend dashboard application that provides
# the user interface for fleet management and monitoring.

COMPONENT_NAME := wrale-dashboard
COMPONENT_DESCRIPTION := Frontend Dashboard - Device management and monitoring UI
MAKEFILES_DIR := ../../../Makefiles

# Help customization
HELP_EXTRA := "Development Note: Run 'make dev' to start development server with hot reload"

# Include UI base template
include $(MAKEFILES_DIR)/templates/ui.mk
$(eval $(UI_TEMPLATE))

# Component-specific targets
.PHONY: storybook

storybook: install-deps ## Run Storybook component explorer
	@echo "Starting Storybook..."
	$(NPM) run storybook