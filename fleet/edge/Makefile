# Edge Agent Makefile
# Manages the edge agents that run on individual devices and handle
# local device control and monitoring.

COMPONENT_NAME := wrale-edge
MAKEFILES_DIR := ../../Makefiles
MAIN_PACKAGE := ./agent

include $(MAKEFILES_DIR)/templates/agent.mk
$(eval $(AGENT_TEMPLATE))

# Standard targets (inherited from agent.mk)
.PHONY: all clean build test docker-build docker-push help

all: ## Build everything (clean, verify, build)
clean: ## Clean build artifacts
build: ## Build the edge agent binary
test: ## Run all tests (unit)
docker-build: ## Build edge agent Docker image
docker-push: ## Push edge agent Docker image to registry

# Component-specific targets
.PHONY: install-deps

install-deps: ## Install agent system dependencies
	@echo "Installing $(COMPONENT_NAME) dependencies..."
	./scripts/install-deps.sh

help: ## Display available targets
	@echo "Wrale Edge Agent - Available targets:"
	@echo
	@echo "Standard targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILES_DIR)/templates/agent.mk | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Component targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'