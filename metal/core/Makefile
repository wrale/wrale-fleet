# Metal Daemon Makefile
# Manages the metal daemon that provides low-level hardware control
# and system coordination.

COMPONENT_NAME := metald
MAKEFILES_DIR := ../../Makefiles

include $(MAKEFILES_DIR)/templates/daemon.mk
$(eval $(DAEMON_TEMPLATE))

# Standard targets (inherited from daemon.mk)
.PHONY: all clean build test docker-build docker-push help

all: ## Build everything (clean, verify, build)
clean: ## Clean build artifacts
build: ## Build the metal daemon binary
test: ## Run all tests (unit, hardware)
docker-build: ## Build metal daemon Docker image
docker-push: ## Push metal daemon Docker image to registry

# Component-specific targets
.PHONY: hardware-test

hardware-test: ## Run hardware-specific tests (requires connected hardware)
	@echo "Running hardware tests for $(COMPONENT_NAME)..."
	go test -v -tags=hardware ./hardware/...

help: ## Display available targets
	@echo "Metal Daemon - Available targets:"
	@echo
	@echo "Standard targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILES_DIR)/templates/daemon.mk | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Component targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'