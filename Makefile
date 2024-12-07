# Root Makefile - Project orchestration
SHELL := /bin/bash

# Component directories
COMPONENTS := fleet fleet/edge metal/core user/api user/ui/wrale-dashboard
BUILD_VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_COMPOSE ?= docker compose

# Environment
ENV ?= development

.PHONY: all build clean test verify setup dev-env docker-up docker-down release help $(COMPONENTS)

all: $(COMPONENTS) ## Build all components

build: $(COMPONENTS) ## Build all components (alias for 'all')

$(COMPONENTS):
	$(MAKE) -C $@

clean test verify: ## Run command for all components
	@for dir in $(COMPONENTS); do \
		echo "Running $@ in $$dir..."; \
		$(MAKE) -C $$dir $@ || exit 1; \
	done

setup: ## Set up development environment
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing UI dependencies..."
	@cd user/ui/wrale-dashboard && npm install
	@echo "Setup complete. Run 'make dev-env' to start development environment"

dev-env: docker-up ## Start development environment
	@echo "Development environment ready"

docker-up: ## Start Docker services
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker services
	$(DOCKER_COMPOSE) down

release: verify ## Create a new release
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Use 'make release VERSION=v1.2.3'"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed"

dev-tools: ## Install development tools
	@go install github.com/golangci/golint/cmd/golangci-lint@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

# Help system
help: ## Show available targets
	@echo "Wrale Fleet v$(BUILD_VERSION) - Main targets:"
	@echo
	@echo "Component Management:"
	@echo "  all          - Build all components"
	@echo "  clean        - Clean all components"
	@echo "  test         - Test all components"
	@echo "  verify       - Verify all components"
	@echo
	@echo "Development:"
	@echo "  setup        - Set up development environment"
	@echo "  dev-env      - Start development environment"
	@echo "  dev-tools    - Install development tools"
	@echo
	@echo "Docker:"
	@echo "  docker-up    - Start Docker services"
	@echo "  docker-down  - Stop Docker services"
	@echo
	@echo "Release:"
	@echo "  release      - Create a new release (requires VERSION=v1.2.3)"
	@echo
	@echo "Components:"
	@echo "The following components are available:"
	@for dir in $(COMPONENTS); do \
		echo "  $$dir"; \
	done
	@echo
	@echo "For component-specific help, run 'make help' in the component directory"
	@echo "Example: cd fleet && make help"

# Default target
.DEFAULT_GOAL := help