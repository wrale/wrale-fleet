# Metal Daemon Makefile
# Manages the metal daemon service that provides hardware access and control.

COMPONENT_NAME := metald
COMPONENT_DESCRIPTION := Metal Daemon - Hardware access and control service
MAKEFILES_DIR := ../Makefiles
MAIN_PACKAGE := ./cmd/metald

# Daemon configuration
DAEMON_USER := root
DAEMON_GROUP := root
DAEMON_CONFIG_DIR := /etc/wrale/metal
DAEMON_DATA_DIR := /var/lib/wrale/metal

include $(MAKEFILES_DIR)/templates/go/daemon.mk
$(eval $(GO_DAEMON_TEMPLATE))

# Component-specific targets
.PHONY: test-hardware verify-hardware systemd

test-hardware: ## Run hardware-specific tests (requires device setup)
	@echo "Running hardware tests..."
	@if [ -n "$(HARDWARE_DEVICE)" ]; then \
		$(GO_ENV) $(GO) test -v -tags=hardware \
			--timeout=10m ./... \
			-args -device=$(HARDWARE_DEVICE); \
	else \
		echo "HARDWARE_DEVICE must be set for hardware testing"; \
		exit 1; \
	fi

verify-hardware: ## Verify hardware configuration
	@echo "Verifying hardware setup..."
	@if [ -x "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) verify --check-only; \
	else \
		echo "Build required before hardware verification"; \
		exit 1; \
	fi

systemd: ## Generate systemd service file
	@echo "Generating systemd service file..."
	@mkdir -p systemd
	@sed "s|{{BINARY}}|/usr/local/sbin/$(BINARY_NAME)|g; \
		s|{{CONFIG}}|$(DAEMON_CONFIG_DIR)/config.yaml|g; \
		s|{{USER}}|$(DAEMON_USER)|g; \
		s|{{GROUP}}|$(DAEMON_GROUP)|g" \
		systemd/metald.service.template > systemd/metald.service