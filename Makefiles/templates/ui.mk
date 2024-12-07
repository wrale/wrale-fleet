# UI-specific template (Next.js)
# Handles npm/node targets without conflicts

# Node/NPM configuration
NODE ?= node
NPM ?= npm
NEXT ?= $(NPM) run
BUILD_DIR ?= .next
DIST_DIR ?= dist

# Override some common targets for npm usage
clean-cmd ?= rm -rf $(BUILD_DIR) $(DIST_DIR) node_modules/.cache
build-cmd ?= $(NPM) run build
test-cmd ?= $(NPM) test
lint-cmd ?= $(NPM) run lint

# Define the UI base template
define UI_BASE_TEMPLATE

.PHONY: clean build test lint dev analyze deploy storybook install-deps

# Standard UI targets with configurable commands
clean: ## Clean build artifacts
	$(clean-cmd)

build: install-deps ## Build the application
	$(build-cmd)

test: install-deps ## Run tests
	$(test-cmd)

lint: install-deps ## Run linter
	$(lint-cmd)

install-deps: ## Install dependencies
	$(NPM) install

dev: install-deps ## Run development server
	$(NEXT) dev

analyze: build ## Analyze bundle size
	$(NEXT) analyze

deploy: build ## Deploy to production
	$(NEXT) deploy

# Pre/post hooks for help target
help-pre:
	@echo "$(COMPONENT_NAME) - $(COMPONENT_DESCRIPTION)"
	@echo
	@echo "NPM Version:  $$($(NPM) -v)"
	@echo "Node Version: $$($(NODE) -v)"
	@echo

help-post:
	@if [ "$(HELP_EXTRA)" != "" ]; then \
		echo; \
		echo "$(HELP_EXTRA)"; \
	fi

endef