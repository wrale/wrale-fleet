# User API Service Makefile
# Manages the REST API service that provides the interface between
# the frontend dashboard and the fleet management system.

COMPONENT_NAME := wrale-api
MAKEFILES_DIR := ../../Makefiles

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Standard targets (inherited from service.mk)
.PHONY: all clean build test docker-build docker-push help

all: ## Build everything (clean, verify, build)
clean: ## Clean build artifacts
build: ## Build the API service binary
test: ## Run all tests (unit)
docker-build: ## Build API service Docker image
docker-push: ## Push API service Docker image to registry

# Component-specific targets
.PHONY: api-spec openapi-validate

api-spec: ## Generate OpenAPI specification
	@echo "Generating API specification..."
	go run ./cmd/spec-gen

openapi-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	openapi-validator ./api/openapi.yaml

help: ## Display available targets
	@echo "Wrale API Service - Available targets:"
	@echo
	@echo "Standard targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILES_DIR)/templates/service.mk | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Component targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'