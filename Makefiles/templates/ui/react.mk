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

# Build configuration
export NODE_ENV ?= $(REACT_ENV)
export BROWSER
export REACT_APP_API_URL = $(REACT_API_URL)
export PORT = $(REACT_PORT)

# Define the React template
define REACT_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: dev build-react start-react lint-react analyze-react storybook test-react

# Development
dev: node-deps ## Start development server
	$(call show_progress,"Starting React development server...")
	@$(NPM) start

# Production build
build-react: node-deps ## Build for production
	$(call show_progress,"Building React production build...")
	@$(NPM) run build

# Production server
start-react: build-react ## Start production server
	$(call show_progress,"Starting React production server...")
	@if command -v serve >/dev/null 2>&1; then \
		serve -s build -l $(REACT_PORT); \
	else \
		$(call log_error,"serve not installed. Run: npm install -g serve"); \
		exit 1; \
	fi

# Linting
lint-react: node-deps ## Run React linting
	$(call show_progress,"Running React lint...")
	@$(NPM) run lint
	@$(NPX) prettier --check .

# Testing
test-react: node-deps ## Run React tests
	$(call show_progress,"Running React tests...")
	@$(NPM) test

test-react-watch: node-deps ## Run React tests in watch mode
	@$(NPM) test -- --watch

test-react-coverage: node-deps ## Run React tests with coverage
	@$(NPM) test -- --coverage

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

# Type checking
types: node-deps ## Type check TypeScript
	$(call show_progress,"Type checking...")
	@if [ -f tsconfig.json ]; then \
		$(NPX) tsc --noEmit; \
	else \
		echo "$(YELLOW)No TypeScript configuration found$(RESET)"; \
	fi

# Bundle analysis
analyze-react: ## Analyze bundle size
	$(call show_progress,"Analyzing bundle size...")
	@if command -v source-map-explorer >/dev/null 2>&1; then \
		$(NPM) run build && \
		source-map-explorer 'build/static/js/*.js'; \
	else \
		$(call log_error,"source-map-explorer not installed"); \
		$(call log_error,"Run: npm install -g source-map-explorer"); \
		exit 1; \
	fi

# Component scaffolding
define COMPONENT_DIR
src/components/$(shell echo $(1) | sed 's/[A-Z]/-\l&/g;s/^-//')
endef

define COMPONENT_FILE
import React from 'react';
import './$(1).css';

export interface $(1)Props {
  // Add props here
}

export const $(1): React.FC<$(1)Props> = (props) => {
  return (
    <div className="$(shell echo $(1) | sed 's/[A-Z]/-\l&/g;s/^-//')">
      {/* Add component content */}
    </div>
  );
};

export default $(1);
endef

define COMPONENT_TEST
import React from 'react';
import { render, screen } from '@testing-library/react';
import { $(1) } from './$(1)';

describe('$(1)', () => {
  it('renders successfully', () => {
    render(<$(1) />);
    // Add your test assertions
  });
});
endef

define COMPONENT_CSS
.$(shell echo $(1) | sed 's/[A-Z]/-\l&/g;s/^-//') {
  /* Add component styles */
}
endef

define COMPONENT_STORY
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
    // Add default props
  },
};
endef

create-component: ## Create new component (make create-component name=MyComponent)
	$(call check_var,name)
	$(call show_progress,"Creating component $(name)...")
	@mkdir -p $(call COMPONENT_DIR,$(name))
	@echo "$$COMPONENT_FILE" | sed 's/$(1)/$(name)/g' > $(call COMPONENT_DIR,$(name))/$(name).tsx
	@echo "$$COMPONENT_TEST" | sed 's/$(1)/$(name)/g' > $(call COMPONENT_DIR,$(name))/$(name).test.tsx
	@echo "$$COMPONENT_CSS" | sed 's/$(1)/$(name)/g' > $(call COMPONENT_DIR,$(name))/$(name).css
	@echo "$$COMPONENT_STORY" | sed 's/$(1)/$(name)/g' > $(call COMPONENT_DIR,$(name))/$(name).stories.tsx
	@echo "$(GREEN)Component created:$(RESET) $(call COMPONENT_DIR,$(name))"

# Override default targets
build: build-react
lint: lint-react
test: test-react types

# Docker support
DOCKER_PORTS = -p $(REACT_PORT):$(REACT_PORT)
DOCKER_ENV = \
	-e NODE_ENV=$(NODE_ENV) \
	-e REACT_APP_API_URL=$(REACT_APP_API_URL)

endef # End of REACT_TEMPLATE

endif # REACT_MK_INCLUDED
