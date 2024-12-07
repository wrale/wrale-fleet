# UI Dashboard Makefile
COMPONENT_NAME := wrale-dashboard
MAKEFILES_DIR := ../../../Makefiles

include $(MAKEFILES_DIR)/templates/ui.mk
$(eval $(UI_TEMPLATE))

# Component-specific targets
.PHONY: dev storybook

dev: ## Run development server
	@echo "Starting development server..."
	$(NPM) run dev

storybook: ## Run Storybook
	@echo "Starting Storybook..."
	$(NPM) run storybook