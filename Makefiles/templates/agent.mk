# Agent-specific template
include $(MAKEFILES_DIR)/templates/base.mk

# Define the agent template
define AGENT_TEMPLATE
$(BASE_TEMPLATE)

.PHONY: package install-deps

build: ## Build the agent
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

package: verify-all ## Create agent package
	@echo "Creating agent package..."
	mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(DIST_DIR)/
	cp scripts/agent-install.sh $(DIST_DIR)/
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz -C $(DIST_DIR) .

endef