# Shared Library Makefile
# Common utilities and shared types

COMPONENT_NAME := wrale-shared
COMPONENT_DESCRIPTION := Shared Library Layer
MAKEFILES_DIR := ../Makefiles

# Simple build setup for library
.PHONY: build test verify

build: ## Build and verify library
	$(call show_progress,"Building shared library...")
	go build ./...

test: ## Run tests
	$(call show_progress,"Testing shared library...")
	go test -v -race ./...

verify: ## Run verification
	$(call show_progress,"Verifying shared library...")
	go vet ./...
	golangci-lint run --timeout=5m

clean: ## Clean build artifacts
	@rm -f coverage.out