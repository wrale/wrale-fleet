# Fleet Service Makefile
# Manages the fleet coordination service that handles device orchestration
# and fleet-wide operations.

COMPONENT_NAME := wrale-fleet
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/fleetd

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Standard targets
.PHONY: all build test clean

all: ## Build everything (clean, verify, build)
	$(MAKE) clean verify build

build: ## Build the fleet service binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test: ## Run all tests (unit, integration)
	$(GOTEST) $(TESTFLAGS) ./...
	$(MAKE) test-integration

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out

# Component-specific targets
.PHONY: integration-test

integration-test: ## Run fleet integration tests
	@echo "Running $(COMPONENT_NAME) integration tests..."
	$(GOTEST) -v -tags=integration ./integration/...

help: ## Show available targets
	@echo "$(COMPONENT_NAME) - Fleet Service"
	@echo "Manages fleet-wide device coordination and orchestration"
	@echo
	$(HELP_FUNCTION)