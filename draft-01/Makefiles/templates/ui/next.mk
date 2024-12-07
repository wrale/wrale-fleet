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
TEST_RESULTS_DIR ?= test-results

# Build configuration
export NODE_ENV ?= $(NEXT_ENV)
export NEXT_TELEMETRY_DISABLED
export NEXT_PUBLIC_API_URL = $(NEXT_API_URL)

# Metrics configuration for monitoring mixin
METRICS_PORT ?= $(shell expr $(NEXT_PORT) + 1000)
METRICS_PATH ?= /metrics
HEALTH_PATH ?= /health

# Define the Next.js template
define NEXTJS_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: dev build-next start-next lint-next analyze-next storybook test-e2e

# Development with monitoring support
dev: node-deps ## Start development server
	$(call show_progress,"Starting Next.js development server...")
ifdef MONITORING_ENABLED
	@$(NPX) cross-env NODE_OPTIONS='-r next-metrics' \
		next dev -p $(NEXT_PORT)
else
	@$(NPX) next dev -p $(NEXT_PORT)
endif

# Production build
build-next: node-deps ## Build for production
	$(call show_progress,"Building Next.js production build...")
	@$(NPX) next build
	@if [ "$(NEXT_ANALYZE)" = "true" ]; then \
		$(MAKE) analyze-next; \
	fi

# Production server with monitoring
start-next: build-next ## Start production server
	$(call show_progress,"Starting Next.js production server...")
ifdef MONITORING_ENABLED
	@$(NPX) cross-env NODE_OPTIONS='-r next-metrics' \
		next start -p $(NEXT_PORT)
else
	@$(NPX) next start -p $(NEXT_PORT)
endif

# Linting
lint-next: node-deps ## Run Next.js linting
	$(call show_progress,"Running Next.js lint...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@$(NPX) next lint --output-file $(TEST_RESULTS_DIR)/lint-results.json
	@$(NPX) prettier --check . | tee $(TEST_RESULTS_DIR)/prettier-results.txt

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

# Testing with result output
test-e2e: node-deps ## Run end-to-end tests
	$(call show_progress,"Running E2E tests...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@if [ -d cypress ]; then \
		$(NPX) cypress run --reporter-options "reportDir=$(TEST_RESULTS_DIR)"; \
	elif [ -d playwright ]; then \
		$(NPX) playwright test --reporter=html,json \
			--report-dir=$(TEST_RESULTS_DIR); \
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

# Type checking with result output
types: node-deps ## Type check TypeScript
	$(call show_progress,"Type checking...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@$(NPX) tsc --noEmit --pretty false \
		| tee $(TEST_RESULTS_DIR)/tsc-results.txt

# Clean
clean: ## Clean build artifacts
	$(call show_progress,"Cleaning Next.js artifacts...")
	@rm -rf .next out storybook-static
	@rm -rf node_modules/.cache
	@rm -rf cypress/screenshots cypress/videos
	@rm -rf playwright-report $(TEST_RESULTS_DIR)

# Override default targets
build: build-next
lint: lint-next
test: types test-e2e

# Docker support - integrates with docker mixin
CONTAINER_BUILD_ARGS ?= \
	--build-arg NEXT_PUBLIC_API_URL=$(NEXT_PUBLIC_API_URL)
CONTAINER_RUN_ARGS ?= \
	-p $(NEXT_PORT):$(NEXT_PORT) \
	-e NODE_ENV=$(NODE_ENV) \
	-e NEXT_TELEMETRY_DISABLED=$(NEXT_TELEMETRY_DISABLED) \
	-e NEXT_PUBLIC_API_URL=$(NEXT_PUBLIC_API_URL)

# Monitoring integration
ifdef MONITORING_ENABLED
node-deps::
	@$(NPX) yarn add next-metrics prometheus-client
endif

# Health check integration
HEALTH_CHECK_CMD = curl -f http://localhost:$(NEXT_PORT)/api/health

endef # End of NEXTJS_TEMPLATE

endif # NEXTJS_MK_INCLUDED