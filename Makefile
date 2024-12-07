# Root Makefile - Project orchestration
SHELL := /bin/bash

# Components in build order
COMPONENTS := shared metal sync fleet user user/ui/wrale-dashboard

# Version info
BUILD_VERSION ?= $(shell git describe --tags --always --dirty)
DOCKER_COMPOSE ?= docker compose
ENV ?= development

.PHONY: all build clean test verify docker-up docker-down setup dev-env release help $(COMPONENTS)

# Component targets
$(COMPONENTS):
	@echo "==> Building $@..."
	@$(MAKE) -C $@ $(TARGET)

# Main targets that operate on all components
all: ## Build everything
	@for dir in $(COMPONENTS); do \
		echo "==> $$dir: Building..."; \
		$(MAKE) -C $$dir all || exit 1; \
	done

build: ## Build all components
	@for dir in $(COMPONENTS); do \
		echo "==> $$dir: Building..."; \
		$(MAKE) -C $$dir build || exit 1; \
	done

clean: ## Clean all components
	@for dir in $(COMPONENTS); do \
		echo "==> $$dir: Cleaning..."; \
		$(MAKE) -C $$dir clean || exit 1; \
	done

test: ## Run all tests
	@for dir in $(COMPONENTS); do \
		echo "==> $$dir: Testing..."; \
		$(MAKE) -C $$dir test || exit 1; \
	done

verify: ## Run all verifications
	@for dir in $(COMPONENTS); do \
		echo "==> $$dir: Verifying..."; \
		$(MAKE) -C $$dir verify || exit 1; \
	done

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