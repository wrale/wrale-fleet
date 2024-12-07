# Go agent template
# Extends base service template with agent-specific functionality
ifndef GO_AGENT_MK_INCLUDED
GO_AGENT_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/service.mk
include $(MAKEFILES_DIR)/impl/go.mk

# Agent-specific configuration
MAIN_PACKAGE ?= ./cmd/agent
AGENT_CONFIG_DIR ?= /etc/$(COMPONENT_NAME)
AGENT_DATA_DIR ?= /var/lib/$(COMPONENT_NAME)
AGENT_LOG_DIR ?= /var/log/$(COMPONENT_NAME)
AGENT_RUN_DIR ?= /var/run/$(COMPONENT_NAME)
AGENT_USER ?= $(COMPONENT_NAME)
AGENT_GROUP ?= $(COMPONENT_NAME)

# Device-specific settings
DEVICE_ID ?= $(shell hostname)
FLEET_SERVER ?= localhost:9090
FLEET_TOKEN ?=
METRICS_PORT ?= 2112

# Define the Go agent template
define GO_AGENT_TEMPLATE
$(SERVICE_TEMPLATE)

.PHONY: install-agent uninstall-agent agent-status agent-logs init-device register-device

install-agent: go-build ## Install agent
	$(call show_progress,"Installing $(COMPONENT_NAME) agent...")
	@sudo useradd -r -s /sbin/nologin \
		-d $(AGENT_DATA_DIR) $(AGENT_USER) 2>/dev/null || true
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/sbin/
	@sudo mkdir -p $(AGENT_CONFIG_DIR) $(AGENT_DATA_DIR) \
		$(AGENT_LOG_DIR) $(AGENT_RUN_DIR)
	@sudo chown -R $(AGENT_USER):$(AGENT_GROUP) \
		$(AGENT_CONFIG_DIR) $(AGENT_DATA_DIR) \
		$(AGENT_LOG_DIR) $(AGENT_RUN_DIR)
	@if [ -f systemd/$(COMPONENT_NAME).service ]; then \
		sudo install -m 644 systemd/$(COMPONENT_NAME).service \
			/etc/systemd/system/$(COMPONENT_NAME).service; \
		sudo systemctl daemon-reload; \
		sudo systemctl enable $(COMPONENT_NAME); \
		echo "$(GREEN)Agent installed and enabled$(RESET)"; \
	else \
		$(call log_error,"systemd service file not found"); \
		exit 1; \
	fi

uninstall-agent: ## Uninstall agent
	$(call show_progress,"Uninstalling $(COMPONENT_NAME) agent...")
	@sudo systemctl stop $(COMPONENT_NAME) || true
	@sudo systemctl disable $(COMPONENT_NAME) || true
	@sudo rm -f /etc/systemd/system/$(COMPONENT_NAME).service
	@sudo systemctl daemon-reload
	@sudo rm -f /usr/local/sbin/$(BINARY_NAME)
	@sudo rm -rf $(AGENT_CONFIG_DIR) $(AGENT_RUN_DIR)
	@echo "$(GREEN)Agent uninstalled$(RESET)"

agent-status: ## Show agent status
	@echo "$(BOLD)Agent Status:$(RESET)"
	@systemctl status $(COMPONENT_NAME) || true
	@echo
	@echo "$(BOLD)Device Information:$(RESET)"
	@echo "  Device ID: $(DEVICE_ID)"
	@echo "  Fleet Server: $(FLEET_SERVER)"
	@if [ -f $(AGENT_CONFIG_DIR)/device.json ]; then \
		echo "  Registration: $(GREEN)Registered$(RESET)"; \
	else \
		echo "  Registration: $(RED)Not registered$(RESET)"; \
	fi

agent-logs: ## Show agent logs
	@journalctl -u $(COMPONENT_NAME) -f

init-device: ## Initialize device configuration
	$(call show_progress,"Initializing device configuration...")
	@sudo mkdir -p $(AGENT_CONFIG_DIR)
	@sudo tee $(AGENT_CONFIG_DIR)/config.yaml > /dev/null << EOF
device_id: $(DEVICE_ID)
fleet_server: $(FLEET_SERVER)
metrics_port: $(METRICS_PORT)
log_dir: $(AGENT_LOG_DIR)
data_dir: $(AGENT_DATA_DIR)
EOF
	@sudo chown $(AGENT_USER):$(AGENT_GROUP) $(AGENT_CONFIG_DIR)/config.yaml
	@echo "$(GREEN)Device configuration initialized$(RESET)"

register-device: ## Register device with fleet
	$(call check_var,FLEET_TOKEN)
	$(call show_progress,"Registering device $(DEVICE_ID)...")
	@$(BUILD_DIR)/$(BINARY_NAME) register \
		--device-id=$(DEVICE_ID) \
		--fleet-server=$(FLEET_SERVER) \
		--fleet-token=$(FLEET_TOKEN)

# Override installation targets
install: install-agent init-device

# Override startup to ensure registration
run: go-build
	@if [ ! -f $(AGENT_CONFIG_DIR)/device.json ]; then \
		$(call log_error,"Device not registered. Run 'make register-device FLEET_TOKEN=<token>' first"); \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(BINARY_NAME) run \
		--config=$(AGENT_CONFIG_DIR)/config.yaml

# Generate systemd service
systemd-service: ## Generate systemd service file
	$(call show_progress,"Generating systemd service file...")
	@mkdir -p systemd
	@cat > systemd/$(COMPONENT_NAME).service << EOF
[Unit]
Description=$(COMPONENT_DESCRIPTION)
After=network.target

[Service]
Type=simple
User=$(AGENT_USER)
Group=$(AGENT_GROUP)
ExecStart=/usr/local/sbin/$(BINARY_NAME) run --config=$(AGENT_CONFIG_DIR)/config.yaml
Restart=always
RestartSec=5
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
EOF
	@echo "$(GREEN)Service file generated:$(RESET) systemd/$(COMPONENT_NAME).service"

# Health check for agent
HEALTH_CHECK_CMD = curl -f http://localhost:$(METRICS_PORT)/health || exit 1

endef # End of GO_AGENT_TEMPLATE

endif # GO_AGENT_MK_INCLUDED