define UI_TEMPLATE
include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/docker.mk

# Node/NPM settings
NODE_ENV ?= production
NPM ?= npm

.PHONY: all build clean test lint verify package deploy help

all: clean verify build docker-build ## Build everything

build: ## Build the UI application
	@echo "Building $(COMPONENT_NAME)..."
	$(NPM) ci
	$(NPM) run build

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf .next out node_modules coverage
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

test: ## Run tests
	@echo "Running tests..."
	$(NPM) run test
	$(NPM) run test:coverage

lint: ## Run linters
	@echo "Running linters..."
	$(NPM) run lint
	$(NPM) run type-check

verify: lint test ## Run all verifications

package: verify docker-build ## Create deployable package
	@echo "Creating distribution package..."
	mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz .next

deploy: package ## Deploy the application
	@echo "Deploying $(COMPONENT_NAME)..."
	./scripts/deploy.sh $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz

help: ## Show this help
	$(call HELP_FUNCTION)
endef