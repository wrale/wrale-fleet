define SERVICE_TEMPLATE
include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/golang.mk
include $(MAKEFILES_DIR)/docker.mk
include $(MAKEFILES_DIR)/verify.mk

.PHONY: all build clean test package deploy help

all: clean verify-all build docker-build ## Build everything

build: go-build ## Build the service

clean: go-clean docker-clean ## Clean all artifacts
	rm -rf $(BUILD_DIR) $(DIST_DIR)

test: go-test ## Run tests

package: verify-all docker-build ## Create deployable package
	@echo "Creating distribution package..."
	mkdir -p $(DIST_DIR)
	cp $(BUILD_DIR)/$(COMPONENT_NAME) $(DIST_DIR)/
	cp Dockerfile $(DIST_DIR)/
	tar -czf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz -C $(DIST_DIR) .

deploy: package ## Deploy the service
	@echo "Deploying $(COMPONENT_NAME)..."
	./scripts/deploy.sh $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz

help: ## Show this help
	$(call HELP_FUNCTION)
endef