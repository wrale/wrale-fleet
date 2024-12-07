# Base service template
# Extends component template with service-specific functionality
ifndef SERVICE_MK_INCLUDED
SERVICE_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/component.mk

# Service-specific configuration
COMPONENT_TYPE = service
SERVICE_PORT ?= 8080
SERVICE_HOST ?= 0.0.0.0
SERVICE_CONFIG_DIR ?= config
SERVICE_LOG_DIR ?= logs
SERVICE_PID_DIR ?= tmp
SERVICE_USER ?= $(shell whoami)

# Health check configuration
HEALTH_CHECK_URL ?= http://$(SERVICE_HOST):$(SERVICE_PORT)/health
HEALTH_CHECK_TIMEOUT ?= 5
HEALTH_CHECK_INTERVAL ?= 10

# Define the service template
define SERVICE_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: run stop restart status logs health-check config

# Service lifecycle management
run: build ## Run the service
	$(call show_progress,"Starting $(COMPONENT_NAME)...")
	$(call ensure_dir,$(SERVICE_LOG_DIR))
	$(call ensure_dir,$(SERVICE_PID_DIR))
	@echo "Starting service on $(SERVICE_HOST):$(SERVICE_PORT)"

stop: ## Stop the service
	$(call show_progress,"Stopping $(COMPONENT_NAME)...")
	@if [ -f $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid ]; then \
		kill $$(cat $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid) 2>/dev/null || true; \
		rm -f $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid; \
	fi

restart: stop run ## Restart the service

# Service monitoring
status: ## Show service status
	@echo "$(BOLD)Service Status:$(RESET)"
	@echo "  Name: $(COMPONENT_NAME)"
	@echo "  Host: $(SERVICE_HOST)"
	@echo "  Port: $(SERVICE_PORT)"
	@if [ -f $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid ]; then \
		pid=$$(cat $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid); \
		if ps -p $$pid >/dev/null; then \
			echo "  Status: $(GREEN)Running (PID: $$pid)$(RESET)"; \
		else \
			echo "  Status: $(RED)Not running (stale PID file)$(RESET)"; \
			rm -f $(SERVICE_PID_DIR)/$(COMPONENT_NAME).pid; \
		fi; \
	else \
		echo "  Status: $(RED)Not running$(RESET)"; \
	fi

health-check: ## Run health check
	$(call show_progress,"Checking $(COMPONENT_NAME) health...")
	@curl -s -o /dev/null -w "%{http_code}" \
		--connect-timeout $(HEALTH_CHECK_TIMEOUT) \
		$(HEALTH_CHECK_URL) | grep -q 200 || \
		(echo "$(RED)Service is unhealthy$(RESET)" && exit 1)
	@echo "$(GREEN)Service is healthy$(RESET)"

logs: ## Show service logs
	$(call show_progress,"Showing logs for $(COMPONENT_NAME)...")
	@if [ -d $(SERVICE_LOG_DIR) ]; then \
		tail -f $(SERVICE_LOG_DIR)/$(COMPONENT_NAME).log; \
	else \
		echo "$(RED)No logs found$(RESET)"; \
	fi

# Configuration management
config-show: ## Show service configuration
	@echo "$(BOLD)Service Configuration:$(RESET)"
	@echo "  Config dir: $(SERVICE_CONFIG_DIR)"
	@echo "  Log dir: $(SERVICE_LOG_DIR)"
	@echo "  PID dir: $(SERVICE_PID_DIR)"
	@echo "  User: $(SERVICE_USER)"
	@if [ -d $(SERVICE_CONFIG_DIR) ]; then \
		echo "$(BOLD)Available configs:$(RESET)"; \
		ls -1 $(SERVICE_CONFIG_DIR)/ 2>/dev/null || echo "  No config files found"; \
	fi

config-validate: ## Validate service configuration
	$(call show_progress,"Validating configuration...")
	@test -d $(SERVICE_CONFIG_DIR) || (echo "$(RED)Config directory not found$(RESET)" && exit 1)
	@echo "$(GREEN)Configuration is valid$(RESET)"

# Override component hooks for services
PREBUILD_HOOK += config-validate
POSTBUILD_HOOK += @echo "$(GREEN)Service build complete$(RESET)"

# Service-specific initialization
init: validate ## Initialize service
	@$(MAKE) -s COMPONENT_TEMPLATE.init
	$(call ensure_dir,$(SERVICE_CONFIG_DIR))
	$(call ensure_dir,$(SERVICE_LOG_DIR))
	$(call ensure_dir,$(SERVICE_PID_DIR))

endef # End of SERVICE_TEMPLATE

endif # SERVICE_MK_INCLUDED