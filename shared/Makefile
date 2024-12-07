# Shared Library Makefile
# Common utilities and shared types

COMPONENT_NAME := wrale-shared
COMPONENT_DESCRIPTION := Shared Library Layer
MAKEFILES_DIR := ../Makefiles

include $(MAKEFILES_DIR)/templates/go/service.mk
$(eval $(GO_SERVICE_TEMPLATE))

# Override the default build since this is a library
build: ## Build and verify library
	@go build ./...

test: ## Run tests
	@go test -race ./...

verify: ## Run verification
	@go vet ./...
	@golangci-lint run --timeout=5m

clean: ## Clean build artifacts
	@rm -f coverage.out

all: verify build test