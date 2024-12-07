# Shared Golang build rules
include $(MAKEFILES_DIR)/common.mk

# Go build settings
GO ?= go
GO_FILES ?= $(shell find . -type f -name '*.go')
GO_PACKAGES ?= $(shell $(GO) list ./...)
LDFLAGS ?= -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.GitCommit=$(GIT_COMMIT)
BUILD_FLAGS ?= -v -ldflags="$(LDFLAGS)"

.PHONY: go-build go-test go-clean go-lint go-verify

go-build: ## Build Go binary
	@echo "Building $(COMPONENT_NAME)..."
	$(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(COMPONENT_NAME) ./cmd/$(COMPONENT_NAME)

go-test: ## Run Go tests
	@echo "Running tests..."
	$(GO) test -v -race -cover $(GO_PACKAGES)
	$(GO) test -v -race -coverprofile=coverage.out $(GO_PACKAGES)
	$(GO) tool cover -html=coverage.out -o coverage.html

go-clean: ## Clean Go build artifacts
	@echo "Cleaning Go artifacts..."
	$(GO) clean -cache -testcache
	rm -f coverage.out coverage.html

go-lint: ## Run Go linters
	@echo "Running Go linters..."
	golangci-lint run
	$(GO) vet $(GO_PACKAGES)
	staticcheck ./...

go-verify: go-lint go-test ## Verify Go code