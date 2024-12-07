# Container build rules (supports Docker and Podman)
include $(MAKEFILES_DIR)/common.mk

# Detect available container engine (prefer Docker if both are available)
DOCKER_CMD := $(shell which docker 2>/dev/null)
PODMAN_CMD := $(shell which podman 2>/dev/null)

ifdef DOCKER_CMD
    CONTAINER_ENGINE := docker
else ifdef PODMAN_CMD
    CONTAINER_ENGINE := podman
endif

CONTAINER_IMAGE = $(DOCKER_REGISTRY)/$(COMPONENT_NAME)
DOCKERFILE ?= Dockerfile

# Check if we have a container engine available
ifeq ($(CONTAINER_ENGINE),)
    $(warning No container engine found. Please install Docker or Podman.)
    CONTAINER_BUILD_AVAILABLE := false
else
    CONTAINER_BUILD_AVAILABLE := true
endif

.PHONY: docker-build docker-push docker-clean container-engine-info

# Display detected container engine
container-engine-info:
ifdef CONTAINER_ENGINE
	@echo "Using container engine: $(CONTAINER_ENGINE)"
else
	@echo "No container engine detected (Docker or Podman required)"
endif

# Build container image (works with both Docker and Podman)
docker-build: container-engine-info ## Build container image
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	@echo "Building container image $(CONTAINER_IMAGE):$(DOCKER_TAG)..."
	$(CONTAINER_ENGINE) build -t $(CONTAINER_IMAGE):$(DOCKER_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-f $(DOCKERFILE) .
else
	@echo "Skipping container build - no container engine available"
endif

# Push container image (works with both Docker and Podman)
docker-push: container-engine-info ## Push container image
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	@echo "Pushing container image $(CONTAINER_IMAGE):$(DOCKER_TAG)..."
	$(CONTAINER_ENGINE) push $(CONTAINER_IMAGE):$(DOCKER_TAG)
else
	@echo "Skipping container push - no container engine available"
endif

# Clean container artifacts (works with both Docker and Podman)
docker-clean: container-engine-info ## Clean container artifacts
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	@echo "Cleaning container artifacts..."
	-$(CONTAINER_ENGINE) rmi $(CONTAINER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
endif