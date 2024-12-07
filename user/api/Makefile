# User API Service Makefile
COMPONENT_NAME := wrale-api
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