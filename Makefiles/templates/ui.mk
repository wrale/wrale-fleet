# UI-specific template (Next.js)
include $(MAKEFILES_DIR)/templates/base.mk

# Define the UI template
define UI_TEMPLATE
$(BASE_TEMPLATE)

# Node/NPM configuration
NODE ?= node
NPM ?= npm
NEXT ?= $(NPM) run
BUILD_DIR ?= .next
DIST_DIR ?= dist

.PHONY: build dev test lint clean deploy

build: ## Build the Next.js application
	$(NPM) install
	$(NEXT) build

dev: ## Run development server
	$(NPM) install
	$(NEXT) dev

test: ## Run tests
	$(NPM) test

test-e2e: ## Run end-to-end tests
	$(NPM) run test:e2e

lint: ## Run linter
	$(NPM) run lint

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) $(DIST_DIR) node_modules/.cache

deploy: build ## Deploy to production
	$(NEXT) deploy

analyze: build ## Analyze bundle size
	$(NEXT) analyze

endef