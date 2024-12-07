# Go language implementation
ifndef GO_MK_INCLUDED
GO_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk
include $(MAKEFILES_DIR)/core/vars.mk

# Check Go installation
ifeq ($(HAS_GO),)
    $(call log_error,"Go is not installed")
    $(error Go is required)
endif

# Go configuration
GO ?= go
GO_ENV ?= CGO_ENABLED=1
GO_TEST_TIMEOUT ?= 5m
GO_BUILD_FLAGS ?= -trimpath
GO_TEST_FLAGS ?= -race -timeout $(GO_TEST_TIMEOUT)
GO_BENCH_FLAGS ?= -benchmem
GO_COVER_FLAGS ?= -coverprofile=coverage.out -covermode=atomic

# Tool installation - use @ to suppress command echo
define go_install_tool
    @if ! command -v $(1) >/dev/null 2>&1; then \
        $(call show_progress,"Installing $(1)..."); \
        $(GO) install $(2); \
    fi
endef

# Go build rules
go-info: ## Show Go build information
	$(call show_progress,"Go version:")
	@$(GO) version
	@echo "GOPATH: $$GOPATH"
	@echo "GOROOT: $$GOROOT"

go-deps: ## Download Go dependencies
	$(call show_progress,"Downloading dependencies...")
	@$(GO) mod download
	@$(GO) mod tidy

go-tools: ## Install required Go tools
	$(call go_install_tool,golangci-lint,github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	$(call go_install_tool,goimports,golang.org/x/tools/cmd/goimports@latest)
	$(call go_install_tool,govulncheck,golang.org/x/vuln/cmd/govulncheck@latest)
	$(call go_install_tool,mockgen,github.com/golang/mock/mockgen@latest)

go-build: ## Build Go binary
	$(call show_build_progress,$(COMPONENT_NAME))
	@$(GO_ENV) $(GO) build $(GO_BUILD_FLAGS) \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

go-test: ## Run Go tests
	$(call show_test_progress,$(COMPONENT_NAME))
	@$(GO_ENV) $(GO) test $(GO_TEST_FLAGS) ./...

go-test-verbose: ## Run Go tests with verbose output
	@$(GO_ENV) $(GO) test -v $(GO_TEST_FLAGS) ./...

go-test-short: ## Run Go tests in short mode
	@$(GO_ENV) $(GO) test -short $(GO_TEST_FLAGS) ./...

go-bench: ## Run Go benchmarks
	@$(GO_ENV) $(GO) test $(GO_BENCH_FLAGS) -run=^$$ -bench=. ./...

go-cover: ## Generate test coverage
	@$(GO_ENV) $(GO) test $(GO_TEST_FLAGS) $(GO_COVER_FLAGS) ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@$(OPEN_CMD) coverage.html

go-cover-func: ## Show test coverage by function
	@$(GO) tool cover -func=coverage.out

go-lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		$(call show_progress,"Running linter..."); \
		golangci-lint run --timeout=5m; \
	else \
		$(call log_error,"golangci-lint not installed. Run 'make go-tools' first"); \
		exit 1; \
	fi

go-fmt: ## Format Go code
	$(call show_progress,"Formatting code...")
	@$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

go-vet: ## Run go vet
	$(call show_progress,"Running go vet...")
	@$(GO) vet ./...

go-clean: ## Clean Go build cache
	$(call show_progress,"Cleaning Go cache...")
	@$(GO) clean -cache -testcache
	@rm -f coverage.out coverage.html

go-vulncheck: ## Run vulnerability check
	@if command -v govulncheck >/dev/null 2>&1; then \
		$(call show_progress,"Checking for vulnerabilities..."); \
		govulncheck ./...; \
	else \
		$(call log_error,"govulncheck not installed. Run 'make go-tools' first"); \
		exit 1; \
	fi

# Function to generate mocks for interfaces
# Usage: $(call go_generate_mock,<package>,<interface>)
define go_generate_mock
	@if command -v mockgen >/dev/null 2>&1; then \
		$(call show_progress,"Generating mock for $(2)..."); \
		mockgen -package=mocks -destination=internal/mocks/$(shell echo $(2) | tr '[:upper:]' '[:lower:]')_mock.go $(1) $(2); \
	else \
		$(call log_error,"mockgen not installed. Run 'make go-tools' first"); \
		exit 1; \
	fi
endef

go-generate-mocks: ## Generate mocks (override in component Makefile)
	$(call show_progress,"No mocks defined. Override this target to generate mocks.")

# Standard Go targets that components can use
go-all: go-deps go-fmt go-vet go-lint go-test go-build ## Full Go build cycle

.PHONY: go-info go-deps go-tools go-build go-test go-test-verbose go-test-short \
        go-bench go-cover go-cover-func go-lint go-fmt go-vet go-clean \
        go-vulncheck go-generate-mocks go-all

endif # GO_MK_INCLUDED
