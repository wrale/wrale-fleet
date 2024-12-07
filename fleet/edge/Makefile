# Edge Agent Makefile
# Manages the edge agents that run on individual devices and handle
# local device control and monitoring.

COMPONENT_NAME := wrale-edge
COMPONENT_DESCRIPTION := Edge Agent - Controls and monitors individual devices
MAKEFILES_DIR := ../../Makefiles
MAIN_PACKAGE := ./agent

# Custom help text
HELP_EXTRA := "Note: The edge agent requires hardware access. Use DEVICE_ID to specify target."

include $(MAKEFILES_DIR)/templates/agent.mk
$(eval $(AGENT_TEMPLATE))

# Component-specific targets
.PHONY: install-deps 

install-deps: ## Install agent system dependencies
	@echo "Installing $(COMPONENT_NAME) dependencies..."
	./scripts/install-deps.sh