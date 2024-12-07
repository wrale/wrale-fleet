# Shared Library Makefile
# Common utilities and shared types

.PHONY: all build test verify clean

all: verify build test

build: ## Build library
	@echo "Building shared library..."
	@go build ./...

test: ## Run tests
	@echo "Testing shared library..."
	@go test -race ./...

verify: ## Run verification
	@echo "Verifying shared library..."
	@go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning shared library..."
	@rm -f coverage.out
	@go clean
	@find . -type d -name "vendor" -exec rm -rf {} +