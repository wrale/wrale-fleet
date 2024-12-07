define DAEMON_TEMPLATE
include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/golang.mk
include $(MAKEFILES_DIR)/verify.mk

.PHONY: all build clean test package install help

all: clean verify-all build ## Build everything

build: go-build ## Build the daemon
	@echo "Additional daemon-specific build steps..."

package: verify-all ## Create daemon package
	@echo "Creating daemon package..."
	mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/$(COMPONENT_NAME) $(DIST_DIR)/
	cp scripts/daemon-install.sh $(DIST_DIR)/
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz -C $(DIST_DIR) .

install: package ## Install the daemon
	@echo "Installing daemon..."
	sudo ./scripts/daemon-install.sh $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz

help: ## Show this help
	$(call HELP_FUNCTION)
endef