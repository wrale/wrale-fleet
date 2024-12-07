# Common make configurations and targets
# This file is included by all component Makefiles

ifndef COMMON_MK_INCLUDED
COMMON_MK_INCLUDED := 1

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
.PHONY: clean fmt lint deps verify help version

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

version: ## Show version info
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

endif # COMMON_MK_INCLUDED