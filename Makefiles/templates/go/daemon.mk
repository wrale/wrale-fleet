# Go daemon template
# Extends base service template with daemon-specific functionality
ifndef GO_DAEMON_MK_INCLUDED
GO_DAEMON_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/service.mk
include $(MAKEFILES_DIR)/impl/go.mk

# Daemon-specific configuration
MAIN_PACKAGE ?= ./cmd/daemon
SYSTEMD_DIR ?= /etc/systemd/system
DAEMON_USER ?= root
DAEMON_GROUP ?= root
DAEMON_CONFIG_DIR ?= /etc/$(COMPONENT_NAME)
DAEMON_DATA_DIR ?= /var/lib/$(COMPONENT_NAME)
DAEMON_LOG_DIR ?= /var/log/$(COMPONENT_NAME)

# Define the Go daemon template
define GO_DAEMON_TEMPLATE
$(SERVICE_TEMPLATE)

.PHONY: install-daemon uninstall-daemon daemon-status daemon-logs systemd-service

install-daemon: go-build ## Install daemon
	$(call show_progress,"Installing $(COMPONENT_NAME) daemon...")
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/sbin/
	@sudo mkdir -p $(DAEMON_CONFIG_DIR) $(DAEMON_DATA_DIR) $(DAEMON_LOG_DIR)
	@sudo chown -R $(DAEMON_USER):$(DAEMON_GROUP) \
		$(DAEMON_CONFIG_DIR) $(DAEMON_DATA_DIR) $(DAEMON_LOG_DIR)
	@if [ -f systemd/$(COMPONENT_NAME).service ]; then \
		sudo install -m 644 systemd/$(COMPONENT_NAME).service \
			$(SYSTEMD_DIR)/$(COMPONENT_NAME).service; \
		sudo systemctl daemon-reload; \
		sudo systemctl enable $(COMPONENT_NAME); \
		echo "$(GREEN)Daemon installed and enabled$(RESET)"; \
	else \
		$(call log_error,"systemd service file not found"); \
		exit 1; \
	fi

uninstall-daemon: ## Uninstall daemon
	$(call show_progress,"Uninstalling $(COMPONENT_NAME) daemon...")
	@sudo systemctl stop $(COMPONENT_NAME) || true
	@sudo systemctl disable $(COMPONENT_NAME) || true
	@sudo rm -f $(SYSTEMD_DIR)/$(COMPONENT_NAME).service
	@sudo systemctl daemon-reload
	@sudo rm -f /usr/local/sbin/$(BINARY_NAME)
	@echo "$(GREEN)Daemon uninstalled$(RESET)"

daemon-status: ## Show daemon status
	@echo "$(BOLD)Daemon Status:$(RESET)"
	@systemctl status $(COMPONENT_NAME) || true

daemon-logs: ## Show daemon logs
	@journalctl -u $(COMPONENT_NAME) -f

systemd-service: ## Generate systemd service file
	$(call show_progress,"Generating systemd service file...")
	@mkdir -p systemd
	@cat > systemd/$(COMPONENT_NAME).service << EOF
[Unit]
Description=$(COMPONENT_DESCRIPTION)
After=network.target

[Service]
Type=simple
User=$(DAEMON_USER)
Group=$(DAEMON_GROUP)
ExecStart=/usr/local/sbin/$(BINARY_NAME)
Restart=always
RestartSec=5
StartLimitInterval=0
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
	@echo "$(GREEN)Service file generated:$(RESET) systemd/$(COMPONENT_NAME).service"

# Daemon operation targets
start: ## Start daemon
	@sudo systemctl start $(COMPONENT_NAME)

stop: ## Stop daemon
	@sudo systemctl stop $(COMPONENT_NAME)

restart: ## Restart daemon
	@sudo systemctl restart $(COMPONENT_NAME)

# Override hooks for daemon install
install: install-daemon

# Status command for daemon
status: daemon-status

endef # End of GO_DAEMON_TEMPLATE

endif # GO_DAEMON_MK_INCLUDED