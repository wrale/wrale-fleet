# Docker build rules
include $(MAKEFILES_DIR)/common.mk

DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(COMPONENT_NAME)
DOCKERFILE ?= Dockerfile

.PHONY: docker-build docker-push docker-clean

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-f $(DOCKERFILE) .

docker-push: ## Push Docker image
	@echo "Pushing Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean: ## Clean Docker artifacts
	@echo "Cleaning Docker artifacts..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true