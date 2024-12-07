# Sync Service Makefile
# Manages the sync service that provides state synchronization and conflict resolution.

COMPONENT_NAME := sync
COMPONENT_DESCRIPTION := Sync Service - State synchronization and conflict resolution
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/syncd

# Service configuration
SERVICE_PORT := 9090
SERVICE_CONFIG_DIR := /etc/wrale/sync

include $(MAKEFILES_DIR)/templates/go/service.mk
$(eval $(GO_SERVICE_TEMPLATE))

# Component-specific targets
.PHONY: test-integration verify-state

test-integration: ## Run integration tests
	@echo "Running sync integration tests..."
	$(GO_ENV) $(GO) test -v -tags=integration ./integration/...

verify-state: ## Verify state consistency
	@echo "Verifying state consistency..."
	@if [ -x "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) verify --check-state-only; \
	else \
		echo "Build required before state verification"; \
		exit 1; \
	fi