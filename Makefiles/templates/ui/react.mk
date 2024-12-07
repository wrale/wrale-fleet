# React template
# Extends base component template with React-specific functionality
ifndef REACT_MK_INCLUDED
REACT_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/component.mk
include $(MAKEFILES_DIR)/impl/node.mk

# React configuration
REACT_PORT ?= 3000
REACT_API_URL ?= http://localhost:8080
REACT_ENV ?= development
BROWSER ?= none
TEST_RESULTS_DIR ?= test-results

# Build configuration
export NODE_ENV ?= $(REACT_ENV)
export BROWSER
export REACT_APP_API_URL = $(REACT_API_URL)
export PORT = $(REACT_PORT)

# Metrics configuration for monitoring mixin
METRICS_PORT ?= $(shell expr $(REACT_PORT) + 1000)
METRICS_PATH ?= /metrics
HEALTH_PATH ?= /health

# Define the React template
define REACT_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: dev build-react start-react lint-react analyze-react storybook test-react component-scaffold

# Development with monitoring support
dev: node-deps ## Start development server with optional monitoring
	$(call show_progress,"Starting React development server...")
ifdef MONITORING_ENABLED
	@$(NPX) cross-env NODE_OPTIONS='-r prom-client-metrics' \
		react-scripts start
else
	@react-scripts start
endif

# Production build with PWA support
build-react: node-deps ## Build for production
	$(call show_progress,"Building React production build...")
	@react-scripts build
	@if [ -f src/serviceWorker.ts ]; then \
		$(NPX) workbox generateSW; \
	fi

# Production server with monitoring
start-react: build-react ## Start production server
	$(call show_progress,"Starting React production server...")
	@if ! command -v serve >/dev/null 2>&1; then \
		$(NPM) install -g serve; \
	fi
ifdef MONITORING_ENABLED
	@$(NPX) cross-env NODE_OPTIONS='-r prom-client-metrics' \
		serve -s build -l $(REACT_PORT)
else
	@serve -s build -l $(REACT_PORT)
endif

# Linting with result output
lint-react: node-deps ## Run comprehensive linting
	$(call show_progress,"Running React lint...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@$(NPX) eslint . --format json \
		--output-file $(TEST_RESULTS_DIR)/eslint-results.json || true
	@$(NPX) prettier --check . \
		| tee $(TEST_RESULTS_DIR)/prettier-results.txt

# Testing with result output
test-react: node-deps ## Run tests with coverage and reporting
	$(call show_progress,"Running React tests...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@JEST_JUNIT_OUTPUT_DIR=$(TEST_RESULTS_DIR) \
		react-scripts test --coverage \
		--coverageDirectory=$(TEST_RESULTS_DIR)/coverage \
		--ci --watchAll=false \
		--reporters=default --reporters=jest-junit

test-react-watch: node-deps ## Run tests in watch mode
	@react-scripts test

# Storybook
storybook: node-deps ## Run Storybook development server
	$(call show_progress,"Starting Storybook...")
	@if [ -f .storybook/main.js ]; then \
		$(NPX) storybook dev -p 6006; \
	else \
		$(call log_error,"Storybook not configured"); \
		exit 1; \
	fi

build-storybook: node-deps ## Build static Storybook
	$(call show_progress,"Building Storybook...")
	@if [ -f .storybook/main.js ]; then \
		$(NPX) storybook build -o storybook-static; \
	else \
		$(call log_error,"Storybook not configured"); \
		exit 1; \
	fi

# Type checking with result output
types: node-deps ## Type check TypeScript
	$(call show_progress,"Type checking...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@if [ -f tsconfig.json ]; then \
		$(NPX) tsc --noEmit --pretty false \
			| tee $(TEST_RESULTS_DIR)/tsc-results.txt; \
	else \
		echo "$(YELLOW)No TypeScript configuration found$(RESET)"; \
	fi

# Bundle analysis
analyze-react: ## Analyze bundle size
	$(call show_progress,"Analyzing bundle size...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@if command -v source-map-explorer >/dev/null 2>&1; then \
		npm run build && \
		source-map-explorer 'build/static/js/*.js' \
			--html $(TEST_RESULTS_DIR)/bundle-analysis.html; \
	else \
		$(call log_error,"source-map-explorer not installed"); \
		$(call log_error,"Run: npm install -g source-map-explorer"); \
		exit 1; \
	fi

# Component scaffolding templates
define COMPONENT_TEMPLATE
import React from 'react';
import { FC } from 'react';
import './$(1).css';

export interface $(1)Props {
  // Define component props here
}

export const $(1): FC<$(1)Props> = (props) => {
  return (
    <div className="$(shell echo $(1) | sed 's/[A-Z]/-\l&/g;s/^-//')">
      {/* Component content */}
    </div>
  );
};
endef

define TEST_TEMPLATE
import { render, screen } from '@testing-library/react';
import { $(1) } from './$(1)';

describe('$(1)', () => {
  it('renders component', () => {
    render(<$(1) />);
    // Add test assertions
  });
});
endef

define STORY_TEMPLATE
import type { Meta, StoryObj } from '@storybook/react';
import { $(1) } from './$(1)';

const meta: Meta<typeof $(1)> = {
  component: $(1),
  title: 'Components/$(1)',
};

export default meta;
type Story = StoryObj<typeof $(1)>;

export const Default: Story = {
  args: {
    // Default props
  },
};
endef

# Component scaffolding
component-scaffold: ## Create new component (make component-scaffold name=Button)
	$(call check_var,name)
	$(call show_progress,"Creating component $(name)...")
	@mkdir -p src/components/$(name)
	@echo "$$COMPONENT_TEMPLATE" | sed 's/\$$(1)/$(name)/g' > src/components/$(name)/$(name).tsx
	@echo "$$TEST_TEMPLATE" | sed 's/\$$(1)/$(name)/g' > src/components/$(name)/$(name).test.tsx
	@echo "$$STORY_TEMPLATE" | sed 's/\$$(1)/$(name)/g' > src/components/$(name)/$(name).stories.tsx
	@touch src/components/$(name)/$(name).css
	@echo "$(GREEN)Created component$(RESET) src/components/$(name)"

# Override default targets
build: build-react
lint: lint-react
test: types test-react

# Docker integration
CONTAINER_BUILD_ARGS ?= \
	--build-arg REACT_APP_API_URL=$(REACT_APP_API_URL)
CONTAINER_RUN_ARGS ?= \
	-p $(REACT_PORT):$(REACT_PORT) \
	-e NODE_ENV=$(NODE_ENV) \
	-e REACT_APP_API_URL=$(REACT_APP_API_URL)

# Monitoring integration
ifdef MONITORING_ENABLED
node-deps::
	@$(NPX) yarn add prom-client express
endif

# Health check integration
HEALTH_CHECK_CMD = curl -f http://localhost:$(REACT_PORT)/health

# Clean
clean: ## Clean build artifacts
	$(call show_progress,"Cleaning React artifacts...")
	@rm -rf build storybook-static node_modules/.cache $(TEST_RESULTS_DIR)
	@rm -rf coverage junit.xml

endef # End of REACT_TEMPLATE

endif # REACT_MK_INCLUDED