# Go-specific target implementations
ifndef GO_IMPL_MK_INCLUDED
GO_IMPL_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/common.mk
include $(MAKEFILES_DIR)/targets.mk

# Go-specific targets
.PHONY: fmt lint deps test-bench coverage docker-build docker-push

build: # Defined in targets.mk
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

clean: # Defined in targets.mk
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out

test: # Defined in targets.mk
	$(GOTEST) $(TESTFLAGS) ./...

verify: fmt lint test # Defined in targets.mk

fmt: ## Format Go code
	$(GOFMT) ./...

lint: ## Run linter
	$(GOLINT) $(LINTFLAGS)

deps: ## Download and tidy dependencies
	$(GOMOD) download
	$(GOMOD) tidy

test-bench: ## Run benchmarks
	$(GOTEST) $(BENCHFLAGS) ./...

coverage: ## Generate test coverage
	$(GOTEST) $(TESTFLAGS) $(COVERFLAGS) ./...

docker-build: ## Build Docker image
	$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: ## Push Docker image
	$(DOCKER) push $(DOCKER_IMAGE):$(DOCKER_TAG)

endif # GO_IMPL_MK_INCLUDED