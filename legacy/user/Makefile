# User Service Makefile
# Builds the user-facing components (API service)

COMPONENT_NAME := wrale-user
COMPONENT_DESCRIPTION := User Service Layer
MAKEFILES_DIR := ../Makefiles

# Build configuration
MAIN_PACKAGE := ./api/cmd/wrale-api
BINARY_NAME := wrale-api

# API configuration
API_PORT ?= 8080
API_DOCS_DIR ?= api/docs

include $(MAKEFILES_DIR)/templates/go/service.mk
$(eval $(GO_SERVICE_TEMPLATE))

# API Documentation
.PHONY: api-doc api-serve api-validate

api-doc: ## Generate API documentation
	$(call show_progress,"Generating API documentation...")
	@mkdir -p $(API_DOCS_DIR)
	@if ! command -v swag >/dev/null 2>&1; then \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g api/cmd/wrale-api/main.go -o $(API_DOCS_DIR)

api-serve: api-doc ## Serve API documentation
	$(call show_progress,"Starting documentation server...")
	@if ! command -v redoc-cli >/dev/null 2>&1; then \
		npm install -g redoc-cli; \
	fi
	@redoc-cli serve $(API_DOCS_DIR)/swagger.json

api-validate: api-doc ## Validate API specification
	$(call show_progress,"Validating API documentation...")
	@if ! command -v swagger >/dev/null 2>&1; then \
		go install github.com/go-swagger/go-swagger/cmd/swagger@latest; \
	fi
	@swagger validate $(API_DOCS_DIR)/swagger.json