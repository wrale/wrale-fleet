# Go API service template
# Extends service template with API-specific functionality
ifndef GO_API_MK_INCLUDED
GO_API_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/go/service.mk
include $(MAKEFILES_DIR)/impl/go.mk

# API configuration
API_PORT ?= 8080
API_VERSION ?= v1
API_TITLE ?= $(COMPONENT_NAME)
API_DESCRIPTION ?= "REST API for $(COMPONENT_NAME)"
SWAGGER_DIR ?= api/swagger

# Override service port for API
SERVICE_PORT = $(API_PORT)

# Define the Go API template
define GO_API_TEMPLATE
$(GO_SERVICE_TEMPLATE)

.PHONY: api-spec api-validate api-serve api-client api-mocks api-test-integration openapi-lint

# Generate OpenAPI spec
api-spec: ## Generate OpenAPI/Swagger specification
	$(call show_progress,"Generating API specification...")
	@mkdir -p $(SWAGGER_DIR)
	@if ! command -v swag >/dev/null 2>&1; then \
		$(GO) install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g $(MAIN_PACKAGE)/main.go \
		--output $(SWAGGER_DIR) \
		--pd --parseVendor

# Validate OpenAPI spec
api-validate: api-spec ## Validate OpenAPI specification
	$(call show_progress,"Validating API specification...")
	@if ! command -v openapi-validator >/dev/null 2>&1; then \
		npm install -g @openapitools/openapi-generator-cli; \
	fi
	@openapi-generator validate -i $(SWAGGER_DIR)/swagger.json

# Serve API documentation
api-serve: api-spec ## Serve API documentation
	$(call show_progress,"Starting API documentation server...")
	@if ! command -v redoc-cli >/dev/null 2>&1; then \
		npm install -g redoc-cli; \
	fi
	@redoc-cli serve $(SWAGGER_DIR)/swagger.json \
		--options.theme.colors.primary.main=$(shell echo "$(COMPONENT_NAME)" | md5sum | cut -c1-6)

# Generate API client
api-client: api-spec ## Generate API client library
	$(call show_progress,"Generating API client...")
	@mkdir -p pkg/client
	@openapi-generator generate -i $(SWAGGER_DIR)/swagger.json \
		-g go -o pkg/client \
		--package-name client

# Generate API mocks
api-mocks: ## Generate interface mocks
	$(call show_progress,"Generating API mocks...")
	@if ! command -v mockgen >/dev/null 2>&1; then \
		$(GO) install github.com/golang/mock/mockgen@latest; \
	fi
	@for i in $$(find pkg/api -name "*.go" -not -name "*_test.go" -not -name "*mock*.go"); do \
		interfaces=$$(grep -l "type.*interface" $$i); \
		if [ -n "$$interfaces" ]; then \
			$(call show_progress,"Generating mocks for $$i"); \
			mockgen -source=$$i -destination=$${i%%.go}_mock_test.go -package=$$(basename $$(dirname $$i)); \
		fi \
	done

# Run integration tests
api-test-integration: ## Run API integration tests
	$(call show_progress,"Running API integration tests...")
	@mkdir -p $(TEST_RESULTS_DIR)
	@$(GO_ENV) $(GO) test -v -tags=integration \
		-coverprofile=$(TEST_RESULTS_DIR)/integration-coverage.out \
		./pkg/api/... ./internal/... \
		| tee $(TEST_RESULTS_DIR)/integration-test.log

# Lint OpenAPI spec
openapi-lint: api-spec ## Lint OpenAPI specification
	$(call show_progress,"Linting OpenAPI specification...")
	@if ! command -v spectral >/dev/null 2>&1; then \
		npm install -g @stoplight/spectral-cli; \
	fi
	@spectral lint $(SWAGGER_DIR)/swagger.json

# Docker integration
CONTAINER_BUILD_ARGS ?= \
	--build-arg API_VERSION=$(API_VERSION)
CONTAINER_RUN_ARGS ?= \
	-p $(API_PORT):$(API_PORT) \
	-e API_VERSION=$(API_VERSION)

# Health check integration
HEALTH_CHECK_CMD = curl -f http://localhost:$(API_PORT)/health

# Monitoring integration
ifdef MONITORING_ENABLED
go-deps::
	@$(GO) get github.com/prometheus/client_golang/prometheus
endif

# Override hooks for API
PREBUILD_HOOK += api-spec api-validate
POSTBUILD_HOOK += $(call show_progress,"API built successfully. Run 'make api-serve' to view documentation")

# Additional test targets
test:: api-test-integration

endef # End of GO_API_TEMPLATE

endif # GO_API_MK_INCLUDED