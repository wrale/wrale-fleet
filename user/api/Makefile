# User API Service Makefile
# Manages the REST API service that provides the interface between
# the frontend dashboard and the fleet management system.

COMPONENT_NAME := wrale-api
COMPONENT_DESCRIPTION := API Service - Frontend integration and service gateway
MAKEFILES_DIR := ../../Makefiles

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Component-specific targets
.PHONY: api-spec openapi-validate

api-spec: ## Generate OpenAPI specification
	@echo "Generating API specification..."
	go run ./cmd/spec-gen

openapi-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	openapi-validator ./api/openapi.yaml