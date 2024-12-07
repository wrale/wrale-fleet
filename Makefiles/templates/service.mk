# Service-specific template
include $(MAKEFILES_DIR)/templates/go.mk

# Define the service template
define SERVICE_TEMPLATE
$(GO_TEMPLATE)

.PHONY: test-integration run deploy

test-integration: ## Run integration tests
	$(GOTEST) $(TESTFLAGS) $(INTFLAGS) ./integration/...

run: build ## Run the service locally
	./$(BUILD_DIR)/$(BINARY_NAME)

deploy: docker-build docker-push ## Deploy the service
	@echo "Deploying $(COMPONENT_NAME) version $(VERSION)..."

monitoring: ## Check service health and metrics
	@echo "Checking $(COMPONENT_NAME) health..."

endef