# Root Makefile - Project orchestration
SHELL := /bin/bash

# Components in build order
COMPONENTS := shared metal sync fleet user user/ui/wrale-dashboard
BUILD_VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_COMPOSE ?= docker compose

# Environment
ENV ?= development

.PHONY: all build clean test verify setup dev-env docker-up docker-down release help

# Run make in each component directory
define run_component
	@echo "==> $1: Running $2..."
	@$(MAKE) -C $1 $2 || exit 1
endef

define run_components
	@for dir in $(COMPONENTS); do \
		$(call run_component,$$dir,$1); \
	done
endef

all: ## Build everything
	$(call run_components,all)

build: ## Build all components
	$(call run_components,build)

clean: ## Clean all components
	$(call run_components,clean)

test: ## Run all tests
	$(call run_components,test)

verify: ## Run all verifications
	$(call run_components,verify)

# Development environment
setup: ## Set up development environment
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Setting up UI development..."
	@cd user/ui/wrale-dashboard && npm install
	@echo "Setup complete - Run 'make dev-env' to start development environment"

dev-env: docker-up ## Start development environment
	@echo "Starting development environment..."
	@$(MAKE) -C user/ui/wrale-dashboard dev &
	@$(MAKE) -C user run &
	@echo "Development environment ready"

# Docker management
docker-up: ## Start Docker dependencies
	@echo "Starting Docker services..."
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker dependencies
	@echo "Stopping Docker services..."
	$(DOCKER_COMPOSE) down

# Release management
release: verify ## Create a new release
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Use 'make release VERSION=v1.2.3'"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed"

# Help output
help: ## Show this help
	@echo "Wrale Fleet v$(BUILD_VERSION)"
	@echo
	@echo "Main targets:"
	@echo "  all      - Build everything"
	@echo "  build    - Build all components"
	@echo "  clean    - Clean all components"
	@echo "  test     - Run all tests"
	@echo "  verify   - Run all verifications"
	@echo
	@echo "Development:"
	@echo "  setup    - Set up development environment"
	@echo "  dev-env  - Start development environment"
	@echo
	@echo "Docker:"
	@echo "  docker-up   - Start Docker services"
	@echo "  docker-down - Stop Docker services"
	@echo
	@echo "Release:"
	@echo "  release  - Create a new release (requires VERSION=v1.2.3)"
	@echo
	@echo "Components (in build order):"
	@for dir in $(COMPONENTS); do \
		echo "  $$dir"; \
	done
	@echo
	@echo "For component-specific help: cd <component> && make help"

# Default target
.DEFAULT_GOAL := help