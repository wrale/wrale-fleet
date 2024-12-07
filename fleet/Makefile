# Fleet Service Makefile
# Manages the fleet coordination service that handles device orchestration
# and fleet-wide operations.

COMPONENT_NAME := wrale-fleet
COMPONENT_DESCRIPTION := Fleet Service - Manages fleet coordination and orchestration
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/fleetd

# Extend test command to include integration tests
TEST_CMD = $(GOTEST) $(TESTFLAGS) ./... && $(MAKE) test-integration

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Component-specific targets
.PHONY: integration-test verify-fleet coverage-report

integration-test: ## Run fleet integration tests
	@echo "Running $(COMPONENT_NAME) integration tests..."
	$(GOTEST) -v -tags=integration ./integration/...

verify-fleet: verify ## Run fleet-specific verifications
	@echo "Running fleet-specific verifications..."
	./scripts/verify-fleet.sh

coverage-report: ## Generate test coverage report
	$(GOTEST) $(TESTFLAGS) $(COVERFLAGS) ./...
	go tool cover -html=coverage.out