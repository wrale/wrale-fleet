# Sync Service Makefile
# State synchronization and conflict resolution implementation

COMPONENT_NAME := wrale-sync
COMPONENT_DESCRIPTION := Sync Service Layer
MAKEFILES_DIR := ../Makefiles

# Build configuration
MAIN_PACKAGE := ./cmd/syncd
BINARY_NAME := syncd

include $(MAKEFILES_DIR)/templates/go/service.mk
$(eval $(GO_SERVICE_TEMPLATE))

# Specialized test targets
.PHONY: test-integration verify-state test-consistency

test-integration: ## Run integration tests
	$(call show_progress,"Running integration tests...")
	$(GO_ENV) $(GO) test -v -tags=integration ./integration/...

test-consistency: ## Run state consistency tests
	$(call show_progress,"Running consistency tests...")
	$(GO_ENV) $(GO) test -v ./manager/... -run "TestConsistency"

verify-state: ## Verify state store integrity
	$(call show_progress,"Verifying state store...")
	@if [ -x "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) verify; \
	else \
		echo "Build required before verification"; \
		exit 1; \
	fi