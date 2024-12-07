# Edge Agent Makefile
COMPONENT_NAME := wrale-edge
MAKEFILES_DIR := ../../Makefiles

include $(MAKEFILES_DIR)/templates/agent.mk
$(eval $(AGENT_TEMPLATE))

# Component-specific targets
.PHONY: install-deps

install-deps: ## Install agent dependencies
	@echo "Installing $(COMPONENT_NAME) dependencies..."
	./scripts/install-deps.sh