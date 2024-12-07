# Metal Daemon Makefile
# Hardware access and control service implementation

COMPONENT_NAME := wrale-metal
COMPONENT_DESCRIPTION := Metal Hardware Layer
MAKEFILES_DIR := ../Makefiles

# Build configuration
MAIN_PACKAGE := ./cmd/metald
BINARY_NAME := metald

include $(MAKEFILES_DIR)/templates/go/daemon.mk
$(eval $(GO_DAEMON_TEMPLATE))

# Hardware testing targets
.PHONY: test-hw sim verify-hw

test-hw: ## Run hardware-specific tests (requires real hardware)
	$(call show_progress,"Running hardware tests...")
	$(GO_ENV) $(GO) test -v -tags=hardware \
		-timeout=10m ./hw/... ./gpio/... ./thermal/...

sim: ## Run tests with hardware simulation
	$(call show_progress,"Running simulated tests...")
	$(GO_ENV) $(GO) test -v -tags=simulation ./...

verify-hw: ## Verify hardware configuration
	$(call show_progress,"Verifying hardware setup...")
	@if [ -x "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		sudo $(BUILD_DIR)/$(BINARY_NAME) verify; \
	else \
		echo "Build required before verification"; \
		exit 1; \
	fi