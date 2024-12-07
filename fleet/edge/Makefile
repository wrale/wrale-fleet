# Edge Agent Makefile
# Manages the edge agents that run on individual devices and handle
# local device control and monitoring.

COMPONENT_NAME := wrale-edge
MAKEFILES_DIR := ../../Makefiles
MAIN_PACKAGE := ./agent

include $(MAKEFILES_DIR)/templates/agent.mk
$(eval $(AGENT_TEMPLATE))

# Component-specific targets
.PHONY: install-deps 

install-deps: ## Install agent system dependencies
	@echo "Installing $(COMPONENT_NAME) dependencies..."
	./scripts/install-deps.sh

help: ## Show available targets
	@echo "Wrale Edge Agent"
	@echo "Manages edge agents for device control and monitoring"
	@echo
	$(HELP_FUNCTION)