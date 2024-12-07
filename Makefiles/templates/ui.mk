# UI-specific template (Next.js)
include $(MAKEFILES_DIR)/templates/base.mk

# Node/NPM configuration
NODE ?= node
NPM ?= npm
NEXT ?= $(NPM) run
BUILD_DIR ?= .next
DIST_DIR ?= dist

# Override base commands with npm equivalents
CLEAN_CMD ?= rm -rf $(BUILD_DIR) $(DIST_DIR) node_modules/.cache
BUILD_CMD ?= $(NPM) install && $(NPM) run build
TEST_CMD ?= $(NPM) test
VERIFY_CMD ?= $(NPM) run lint

# Define the UI template with npm-specific targets
define UI_TEMPLATE
$(BASE_TEMPLATE)

.PHONY: dev analyze deploy install-deps lint-fix

install-deps: ## Install dependencies
	$(NPM) install

dev: install-deps ## Run development server
	$(NEXT) dev

analyze: build ## Analyze bundle size
	$(NEXT) analyze

deploy: build ## Deploy to production
	$(NEXT) deploy

lint-fix: install-deps ## Run linter with auto-fix
	$(NPM) run lint -- --fix

endef