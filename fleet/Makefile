# Fleet Service Makefile
# Manages the fleet coordination service that handles device orchestration
# and fleet-wide operations.

COMPONENT_NAME := wrale-fleet
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/fleetd

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Component-specific targets
.PHONY: integration-test

integration-test: ## Run fleet integration tests (requires running service)
	@echo "Running $(COMPONENT_NAME) integration tests..."
	go test -v -tags=integration ./integration/...

# Override help to provide component-specific information
help: ## Show available targets with descriptions
	@echo "$(COMPONENT_NAME) - Fleet Service"
	@echo "Manages fleet-wide device coordination and orchestration"
	@echo
	$(HELP_FUNCTION)