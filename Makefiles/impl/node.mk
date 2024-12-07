# Node.js implementation
ifndef NODE_MK_INCLUDED
NODE_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk
include $(MAKEFILES_DIR)/core/vars.mk

# Check Node.js installation
ifeq ($(HAS_NODE),)
    $(call log_error,"Node.js is not installed")
    $(error Node.js is required)
endif

# Node configuration
NODE ?= node
NPM ?= npm
YARN ?= yarn
NPX ?= npx
NODE_ENV ?= development

# Package manager detection (prefer yarn if both are available)
YARN_LOCK := $(wildcard yarn.lock)
PACKAGE_LOCK := $(wildcard package-lock.json)

ifdef YARN_LOCK
    PKG_MANAGER := yarn
    PKG_INSTALL := $(YARN) install
    PKG_ADD := $(YARN) add
    PKG_REMOVE := $(YARN) remove
    PKG_BUILD := $(YARN) build
    PKG_START := $(YARN) start
    PKG_TEST := $(YARN) test
    PKG_LINT := $(YARN) lint
else
    PKG_MANAGER := npm
    PKG_INSTALL := $(NPM) install
    PKG_ADD := $(NPM) install
    PKG_REMOVE := $(NPM) uninstall
    PKG_BUILD := $(NPM) run build
    PKG_START := $(NPM) start
    PKG_TEST := $(NPM) test
    PKG_LINT := $(NPM) run lint
endif

# Common directories
NODE_MODULES := node_modules
NEXT_DIR := .next
DIST_DIR := dist

node-info: ## Show Node.js information
	$(call show_progress,"Node.js version:")
	@$(NODE) --version
	@echo "NPM version: $$($(NPM) --version)"
	@echo "Package manager: $(PKG_MANAGER)"
	@echo "Environment: $(NODE_ENV)"

node-deps: package.json ## Install dependencies
	$(call show_progress,"Installing dependencies...")
	@$(PKG_INSTALL)

node-deps-dev: package.json ## Install dependencies including devDependencies
	$(call show_progress,"Installing dev dependencies...")
	@NODE_ENV=development $(PKG_INSTALL)

node-build: ## Build the application
	$(call show_build_progress,$(COMPONENT_NAME))
	@NODE_ENV=production $(PKG_BUILD)

node-dev: ## Start development server
	$(call show_progress,"Starting development server...")
	@$(PKG_START)

node-test: ## Run tests
	$(call show_test_progress,$(COMPONENT_NAME))
	@$(PKG_TEST)

node-test-watch: ## Run tests in watch mode
	@$(PKG_TEST) --watch

node-test-coverage: ## Run tests with coverage
	@$(PKG_TEST) --coverage

node-lint: ## Run linter
	$(call show_progress,"Running linter...")
	@$(PKG_LINT)

node-lint-fix: ## Run linter with auto-fix
	@$(PKG_LINT) --fix

node-clean: ## Clean build artifacts
	$(call show_progress,"Cleaning build artifacts...")
	@rm -rf $(NODE_MODULES) $(NEXT_DIR) $(DIST_DIR)
	@rm -f coverage.out coverage.html .eslintcache

node-audit: ## Run security audit
	$(call show_progress,"Running security audit...")
	@$(PKG_MANAGER) audit

node-outdated: ## Check for outdated dependencies
	$(call show_progress,"Checking for outdated dependencies...")
	@$(PKG_MANAGER) outdated

node-update: ## Update dependencies
	$(call show_progress,"Updating dependencies...")
ifeq ($(PKG_MANAGER),yarn)
	@$(YARN) upgrade
else
	@$(NPM) update
endif

# Function to add a dependency
# Usage: $(call node_add_dep,package-name[,dev])
define node_add_dep
	$(call show_progress,"Adding $(1)...")
	@if [ "$(2)" = "dev" ]; then \
		$(PKG_ADD) -D $(1); \
	else \
		$(PKG_ADD) $(1); \
	fi
endef

# Function to remove a dependency
# Usage: $(call node_remove_dep,package-name)
define node_remove_dep
	$(call show_progress,"Removing $(1)...")
	@$(PKG_REMOVE) $(1)
endef

# Standard Node.js targets that components can use
node-all: node-deps node-lint node-test node-build ## Full Node.js build cycle

.PHONY: node-info node-deps node-deps-dev node-build node-dev node-test \
        node-test-watch node-test-coverage node-lint node-lint-fix node-clean \
        node-audit node-outdated node-update node-all

endif # NODE_MK_INCLUDED
