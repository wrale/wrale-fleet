# Fleet Service Makefile
COMPONENT_NAME := wrale-fleet
MAKEFILES_DIR := ../Makefiles

include $(MAKEFILES_DIR)/templates/service.mk
$(eval $(SERVICE_TEMPLATE))

# Component-specific targets
.PHONY: integration-test

integration-test: ## Run integration tests
	@echo "Running $(COMPONENT_NAME) integration tests..."
	go test -v -tags=integration ./integration/...