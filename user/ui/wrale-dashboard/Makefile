# UI Dashboard Makefile
# Manages the Next.js frontend dashboard application that provides
# the user interface for fleet management and monitoring.

COMPONENT_NAME := wrale-dashboard
MAKEFILES_DIR := ../../../Makefiles

include $(MAKEFILES_DIR)/templates/ui.mk
$(eval $(UI_TEMPLATE))

# Standard targets (inherited from ui.mk)
.PHONY: all clean build test dev lint deploy analyze help

all: ## Build everything (clean, verify, build)
clean: ## Clean build artifacts and dependencies
build: ## Build the Next.js application
test: ## Run all tests (unit, e2e)
test-e2e: ## Run end-to-end tests
lint: ## Run ESLint and type checking
deploy: ## Deploy to production (requires auth)
analyze: ## Analyze bundle size

# Component-specific targets
.PHONY: storybook

dev: ## Run development server with hot reload
	@echo "Starting development server..."
	$(NPM) run dev

storybook: ## Run Storybook component explorer
	@echo "Starting Storybook..."
	$(NPM) run storybook

help: ## Display available targets
	@echo "Wrale Dashboard - Available targets:"
	@echo
	@echo "Standard targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILES_DIR)/templates/ui.mk | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Component targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'