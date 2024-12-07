# Fleet Service Makefile
# Manages the fleet coordination service that handles device orchestration
# and fleet-wide operations.

COMPONENT_NAME := wrale-fleet
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/fleetd

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Standard targets (inherited from service.mk)
.PHONY: all clean build test docker-build docker-push help

all: ## Build everything (clean, verify, build)
clean: ## Clean build artifacts
build: ## Build the fleet service binary
test: ## Run all tests (unit, integration)
docker-build: ## Build fleet service Docker image
docker-push: ## Push fleet service Docker image to registry

# Component-specific targets
.PHONY: integration-test

integration-test: ## Run integration tests (requires running fleet service)
	@echo "Running $(COMPONENT_NAME) integration tests..."
	go test -v -tags=integration ./integration/...

help: ## Display available targets
	@echo "Wrale Fleet Service - Available targets:"
	@echo
	@echo "Standard targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILES_DIR)/templates/service.mk | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Component targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'