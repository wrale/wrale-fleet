define AGENT_TEMPLATE
include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/golang.mk
include $(MAKEFILES_DIR)/docker.mk
include $(MAKEFILES_DIR)/verify.mk

.PHONY: all build clean test package deploy help

all: clean verify-all build ## Build everything

build: go-build ## Build the agent
	@echo "Additional agent-specific build steps..."

package: verify-all ## Create agent package
	@echo "Creating agent package..."
	mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/$(COMPONENT_NAME) $(DIST_DIR)/
	cp scripts/agent-install.sh $(DIST_DIR)/
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz -C $(DIST_DIR) .

help: ## Show this help
	$(call HELP_FUNCTION)
endef