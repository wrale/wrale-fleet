define UI_TEMPLATE
include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/docker.mk

# Node/NPM settings
NODE_ENV ?= production
NPM ?= npm

# Check if Docker is available
DOCKER_AVAILABLE := $(shell which docker 2>/dev/null)

.PHONY: all build clean test lint verify package deploy help

all: clean verify build ## Build everything
ifneq ($(DOCKER_AVAILABLE),)
	$(MAKE) docker-build
else
	@echo "Docker not found - skipping container build"
endif

build: ## Build the UI application
	@echo "Building $(COMPONENT_NAME)..."
	$(NPM) ci
	$(NPM) run build

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf .next out node_modules coverage
ifneq ($(DOCKER_AVAILABLE),)
	-docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
endif

test: ## Run tests
	@echo "Running tests..."
	$(NPM) run test
	$(NPM) run test:coverage

lint: ## Run linters
	@echo "Running linters..."
	$(NPM) run lint
	$(NPM) run type-check

verify: lint test ## Run all verifications

package: verify ## Create deployable package
ifneq ($(DOCKER_AVAILABLE),)
	$(MAKE) docker-build
endif
	@echo "Creating distribution package..."
	mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz .next

deploy: package ## Deploy the application
	@echo "Deploying $(COMPONENT_NAME)..."
	./scripts/deploy.sh $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz

help: ## Show this help
	$(call HELP_FUNCTION)
endef