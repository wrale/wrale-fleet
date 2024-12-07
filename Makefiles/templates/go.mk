# Go-specific template
include $(MAKEFILES_DIR)/templates/base.mk

# Define Go-specific build commands
BUILD_CMD = mkdir -p $(BUILD_DIR) && $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
TEST_CMD = $(GOTEST) $(TESTFLAGS) ./...
VERIFY_CMD = $(MAKE) fmt && $(MAKE) lint && $(GOTEST) $(TESTFLAGS) ./...

# Define the Go template
define GO_TEMPLATE
$(BASE_TEMPLATE)

.PHONY: fmt lint test-bench coverage docker-build docker-push

fmt: ## Format Go code
	$(GOFMT) ./...

lint: ## Run linter
	$(GOLINT) $(LINTFLAGS)

test-bench: ## Run benchmarks
	$(GOTEST) $(BENCHFLAGS) ./...

coverage: ## Generate test coverage
	$(GOTEST) $(TESTFLAGS) $(COVERFLAGS) ./...

docker-build: ## Build Docker image
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push Docker image
	$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)

endef