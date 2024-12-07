# Common variables and utilities
SHELL := /bin/bash
MAKEFILES_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD)

# Directory structure
BUILD_DIR ?= build
DIST_DIR ?= dist

# Build settings
MAIN_PACKAGE ?= ./cmd/$(COMPONENT_NAME)  # Default for backward compatibility

# Docker registry settings
DOCKER_REGISTRY ?= wrale
DOCKER_TAG ?= $(VERSION)

# Help function
define HELP_FUNCTION
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
endef