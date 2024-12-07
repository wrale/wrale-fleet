# UI Dashboard Makefile
# Next.js frontend dashboard for fleet management and monitoring

COMPONENT_NAME := wrale-dashboard
COMPONENT_DESCRIPTION := Frontend Dashboard
MAKEFILES_DIR := ../../../Makefiles

# Next.js configuration
NEXT_PORT ?= 3000
NEXT_API_URL ?= http://localhost:8080
NEXT_ENV ?= development
TEST_RESULTS_DIR ?= test-results

include $(MAKEFILES_DIR)/templates/ui/next.mk
$(eval $(NEXTJS_TEMPLATE))

# Development defaults
.env.development:
	@echo "Creating default development environment..."
	@echo "NEXT_PUBLIC_API_URL=$(NEXT_API_URL)" > .env.development
	@echo "NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws" >> .env.development

# Ensure development environment exists
dev: .env.development

# Additional storybook configuration
.storybook/main.js:
	@echo "Initializing Storybook..."
	@npx sb init --builder webpack5

# Storybook with proper initialization
storybook: .storybook/main.js ## Run Storybook component explorer
	@echo "Starting Storybook..."
	@npm run storybook

# Clean should handle Next.js and Storybook artifacts
clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf .next out storybook-static $(TEST_RESULTS_DIR)
	@rm -rf node_modules/.cache
	@rm -rf coverage .coverage cypress/screenshots cypress/videos
	@rm -f .env.development