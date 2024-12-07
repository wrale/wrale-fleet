# Go-specific template
include $(MAKEFILES_DIR)/templates/base.mk

# Define the Go template
define GO_TEMPLATE
$(BASE_TEMPLATE)

.PHONY: build test test-unit test-bench coverage

build: ## Build the binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test: test-unit ## Run tests
	@echo "Tests completed successfully"

test-unit: ## Run unit tests
	$(GOTEST) $(TESTFLAGS) ./...

test-bench: ## Run benchmarks
	$(GOTEST) $(BENCHFLAGS) ./...

coverage: ## Generate test coverage
	$(GOTEST) $(TESTFLAGS) $(COVERFLAGS) ./...
	go tool cover -html=coverage.out

docker-build: ## Build Docker image
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push Docker image
	$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)

endef