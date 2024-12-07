# Metal Daemon Makefile
COMPONENT_NAME := metald
MAKEFILES_DIR := ../../Makefiles

include $(MAKEFILES_DIR)/templates/daemon.mk
$(eval $(DAEMON_TEMPLATE))

# Component-specific targets
.PHONY: hardware-test

hardware-test: ## Run hardware-specific tests
	@echo "Running hardware tests for $(COMPONENT_NAME)..."
	go test -v -tags=hardware ./hardware/...