# Metal Daemon Makefile
# Manages the metal daemon that provides low-level hardware control
# and system coordination.

COMPONENT_NAME := metald
COMPONENT_DESCRIPTION := Metal Daemon - Low-level hardware control and system coordination
MAKEFILES_DIR := ../../Makefiles

# Help text customization
HELP_EXTRA := "Note: Some targets require connected hardware or simulation environment"

include $(MAKEFILES_DIR)/templates/daemon.mk
$(eval $(DAEMON_TEMPLATE))

# Component-specific targets
.PHONY: hardware-test

hardware-test: ## Run hardware-specific tests (requires connected hardware)
	@echo "Running hardware tests for $(COMPONENT_NAME)..."
	$(GOTEST) -v -tags=hardware ./hardware/...