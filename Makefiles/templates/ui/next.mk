# Next.js template
# Extends base component template with Next.js-specific functionality
ifndef NEXTJS_MK_INCLUDED
NEXTJS_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/component.mk
include $(MAKEFILES_DIR)/impl/node.mk

# Next.js configuration
NEXT_PORT ?= 3000
NEXT_API_URL ?= http://localhost:8080
NEXT_ENV ?= development
NEXT_ANALYZE ?= false
NEXT_TELEMETRY_DISABLED ?= 1

# Build configuration
export NODE_ENV ?= $(NEXT_ENV)
export NEXT_TELEMETRY_DISABLED
export NEXT_PUBLIC_API_URL = $(NEXT_API_URL)

# Define the Next.js template
define NEXTJS_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: dev build-next start-next lint-next analyze-next storybook test-e2e

# Development
dev: node-deps ## Start development server
	$(call show_progress,"Starting Next.js development server...")
	@$(NPX) next dev -p $(NEXT_PORT)

# Production build
build-next: node-deps ## Build for production
	$(call show_progress,"Building Next.js production build...")
	@$(NPX) next build
	@if [ "$(NEXT_ANALYZE)" = "true" ]; then \
		$(MAKE) analyze-next; \
	fi

# Production server
start-next: build-next ## Start production server
	$(call show_progress,"Starting Next.js production server...")
	@$(NPX) next start -p $(NEXT_PORT)

# Linting
lint-next: node-deps ## Run Next.js linting
	$(call show_progress,"Running Next.js lint...")
	@$(NPX) next lint
	@$(NPX) prettier --check .

# Bundle analysis
analyze-next: ## Analyze bundle size
	$(call show_progress,"Analyzing bundle size...")
	@ANALYZE=true $(NPX) next build

# Storybook
storybook: node-deps ## Run Storybook
	$(call show_progress,"Starting Storybook...")
	@if [ -f .storybook/main.js ]; then \
		$(NPX) storybook dev -p 6006; \
	else \
		$(call log_error,"Storybook not configured"); \
		exit 1; \
	fi

build-storybook: node-deps ## Build Storybook
	$(call show_progress,"Building Storybook...")
	@if [ -f .storybook/main.js ]; then \
		$(NPX) storybook build -o storybook-static; \
	else \
		$(call log_error,"Storybook not configured"); \
		exit 1; \
	fi

# Testing
test-e2e: node-deps ## Run end-to-end tests
	$(call show_progress,"Running E2E tests...")
	@if [ -d cypress ]; then \
		$(NPX) cypress run; \
	elif [ -d playwright ]; then \
		$(NPX) playwright test; \
	else \
		$(call log_error,"No E2E test framework found"); \
		exit 1; \
	fi

test-e2e-ui: node-deps ## Run E2E tests with UI
	@if [ -d cypress ]; then \
		$(NPX) cypress open; \
	elif [ -d playwright ]; then \
		$(NPX) playwright test --ui; \
	else \
		$(call log_error,"No E2E test framework found"); \
		exit 1; \
	fi

# Type checking
types: node-deps ## Type check TypeScript
	$(call show_progress,"Type checking...")
	@$(NPX) tsc --noEmit

# Clean
clean: ## Clean build artifacts
	$(call show_progress,"Cleaning Next.js artifacts...")
	@rm -rf .next out storybook-static
	@rm -rf node_modules/.cache
	@rm -rf cypress/screenshots cypress/videos
	@rm -rf playwright-report test-results

# Override default targets
build: build-next
lint: lint-next
test: types

# Docker support
DOCKER_PORTS = -p $(NEXT_PORT):$(NEXT_PORT)
DOCKER_ENV = \
	-e NODE_ENV=$(NODE_ENV) \
	-e NEXT_TELEMETRY_DISABLED=$(NEXT_TELEMETRY_DISABLED) \
	-e NEXT_PUBLIC_API_URL=$(NEXT_PUBLIC_API_URL)

endef # End of NEXTJS_TEMPLATE

endif # NEXTJS_MK_INCLUDED