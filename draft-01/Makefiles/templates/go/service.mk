# Go service template
# Extends base service template with Go-specific functionality
ifndef GO_SERVICE_MK_INCLUDED
GO_SERVICE_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/service.mk
include $(MAKEFILES_DIR)/impl/go.mk

# Go-specific service configuration
MAIN_PACKAGE ?= ./cmd/server
HTTP_PORT ?= 8080
GRPC_PORT ?= 9090
METRICS_PORT ?= 2112

# Define the Go service template
define GO_SERVICE_TEMPLATE
$(SERVICE_TEMPLATE)

# Override service port with HTTP port
SERVICE_PORT = $(HTTP_PORT)

.PHONY: go-run go-debug go-profile go-bench-service go-generate go-swagger

# Service implementation
run: go-build ## Run the service
	$(call show_progress,"Starting $(COMPONENT_NAME) service...")
	@$(BUILD_DIR)/$(BINARY_NAME) \
		--http-port=$(HTTP_PORT) \
		--grpc-port=$(GRPC_PORT) \
		--metrics-port=$(METRICS_PORT)

go-debug: go-build ## Run service with delve debugger
	$(call show_progress,"Starting debug session...")
	@if ! command -v dlv &> /dev/null; then \
		$(call log_error,"dlv not found. Install with: go install github.com/go-delve/delve/cmd/dlv@latest"); \
		exit 1; \
	fi
	@dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient \
		exec $(BUILD_DIR)/$(BINARY_NAME) -- \
		--http-port=$(HTTP_PORT) \
		--grpc-port=$(GRPC_PORT) \
		--metrics-port=$(METRICS_PORT)

go-profile: go-build ## Run service with profiling enabled
	$(call show_progress,"Starting with profiling...")
	@GOGC=off $(BUILD_DIR)/$(BINARY_NAME) \
		--http-port=$(HTTP_PORT) \
		--grpc-port=$(GRPC_PORT) \
		--metrics-port=$(METRICS_PORT) \
		--profile

go-bench-service: ## Run service benchmarks
	$(call show_progress,"Running service benchmarks...")
	@go test -bench=. -benchmem ./... \
		-run=^$$ \
		-benchtime=5s \
		-cpu=1,2,4

go-generate: ## Run go generate
	$(call show_progress,"Generating code...")
	@go generate ./...

go-swagger: ## Generate Swagger/OpenAPI docs
	$(call show_progress,"Generating API documentation...")
	@if ! command -v swag &> /dev/null; then \
		$(call log_error,"swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"); \
		exit 1; \
	fi
	@swag init -g $(MAIN_PACKAGE)/main.go -o ./api/swagger

# Docker support
DOCKER_PORTS = -p $(HTTP_PORT):$(HTTP_PORT) \
               -p $(GRPC_PORT):$(GRPC_PORT) \
               -p $(METRICS_PORT):$(METRICS_PORT)

# Override hooks
PREBUILD_HOOK += go-generate
POSTBUILD_HOOK += go-swagger

# Health check command
HEALTH_CHECK_CMD = curl -f http://localhost:$(HTTP_PORT)/health || exit 1

endef # End of GO_SERVICE_TEMPLATE

endif # GO_SERVICE_MK_INCLUDED