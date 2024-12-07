# Common make configurations and targets
# This file is included by all component Makefiles

# Go parameters
GOCMD ?= go
GOBUILD ?= $(GOCMD) build
GOTEST ?= $(GOCMD) test
GOGET ?= $(GOCMD) get
GOMOD ?= $(GOCMD) mod
GOFMT ?= $(GOCMD) fmt

# Build parameters
BUILD_DIR ?= build
DIST_DIR ?= dist
BINARY_NAME ?= $(COMPONENT_NAME)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Test flags
TESTFLAGS ?= -v -race
COVERFLAGS ?= -coverprofile=coverage.out
BENCHFLAGS ?= -bench=. -benchmem

# Tag-based test categories
SIMFLAGS ?= -tags=simulation
HWFLAGS ?= -tags=hardware
INTFLAGS ?= -tags=integration

# Linting
GOLINT ?= golangci-lint
LINTFLAGS ?= run --timeout=5m

# Docker
DOCKER ?= docker
DOCKER_IMAGE ?= wrale/$(COMPONENT_NAME)
DOCKER_TAG ?= $(VERSION)

# Default build flags
LDFLAGS ?= -X main.version=$(VERSION) \
           -X main.commit=$(COMMIT) \
           -X main.buildTime=$(BUILD_TIME)

# Simulation environment
SIM_DIR ?= /tmp/wrale-sim

# Common targets
.PHONY: clean fmt lint deps verify help

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out

fmt: ## Format Go code
	$(GOFMT) ./...

lint: ## Run linter
	$(GOLINT) $(LINTFLAGS)

deps: ## Download and tidy dependencies
	$(GOMOD) download
	$(GOMOD) tidy

verify: fmt lint test coverage ## Run all verifications

# Helper function to process help targets
define help_single_mk
$(shell grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(1) 2>/dev/null | sort -u | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}')
endef

# Helper function to generate help output
define HELP_FUNCTION
	@echo "$(COMPONENT_NAME) - Available targets:"
	@echo
	@echo "Standard targets:"
	@$(foreach mk,$(MAKEFILES_DIR)/templates/*.mk,$(call help_single_mk,$(mk)))
	@echo
	@echo "Component targets:"
	@$(call help_single_mk,$(MAKEFILE_LIST))
endef

# Version information target
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"