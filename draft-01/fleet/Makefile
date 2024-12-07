# Fleet Service and Edge Agent Makefile
# Builds both central coordination service and edge agents

COMPONENT_NAME := wrale-fleet
COMPONENT_DESCRIPTION := Fleet Control System
MAKEFILES_DIR := ../Makefiles

# Build configuration
BRAIN_PACKAGE := ./cmd/fleetd
EDGE_PACKAGE := ./cmd/fleet-agent
BRAIN_BINARY := fleetd
EDGE_BINARY := fleet-agent

include $(MAKEFILES_DIR)/templates/go/service.mk
$(eval $(GO_SERVICE_TEMPLATE))

# Main build overrides to handle both binaries
build: build-brain build-edge ## Build both brain and edge components

build-brain: ## Build fleet brain service
	$(call show_progress,"Building brain service...")
	@CGO_ENABLED=1 $(GO) build -o $(BUILD_DIR)/$(BRAIN_BINARY) $(BRAIN_PACKAGE)

build-edge: ## Build edge agent
	$(call show_progress,"Building edge agent...")
	@CGO_ENABLED=1 $(GO) build -o $(BUILD_DIR)/$(EDGE_BINARY) $(EDGE_PACKAGE)

# Installation targets
install: install-brain install-edge ## Install all components

install-brain: build-brain ## Install brain service
	@sudo install -m 755 $(BUILD_DIR)/$(BRAIN_BINARY) /usr/local/sbin/
	@sudo mkdir -p /etc/wrale/fleet
	@sudo mkdir -p /var/lib/wrale/fleet

install-edge: build-edge ## Install edge agent
	@sudo install -m 755 $(BUILD_DIR)/$(EDGE_BINARY) /usr/local/sbin/

# Testing targets
test: test-brain test-edge ## Run all tests

test-brain: ## Test brain service
	$(GO_ENV) $(GO) test -v ./cmd/fleetd/... ./brain/... ./service/...

test-edge: ## Test edge agent
	$(GO_ENV) $(GO) test -v ./cmd/fleet-agent/... ./edge/...

test-integration: ## Run integration tests
	$(GO_ENV) $(GO) test -v -tags=integration ./integration/...

# Cleanup
clean: ## Clean build artifacts
	@rm -f $(BUILD_DIR)/$(BRAIN_BINARY) $(BUILD_DIR)/$(EDGE_BINARY)
	@rm -f coverage.out

# Additional targets
.PHONY: build-brain build-edge install-brain install-edge test-brain test-edge test-integration