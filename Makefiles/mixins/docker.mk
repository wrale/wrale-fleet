# Docker/Podman mixin providing container functionality
ifndef DOCKER_MK_INCLUDED
DOCKER_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk

# Auto-detect container engine (prefer Docker if both are available)
DOCKER_CMD := $(shell which docker 2>/dev/null)
PODMAN_CMD := $(shell which podman 2>/dev/null)

ifdef DOCKER_CMD
    CONTAINER_ENGINE := docker
else ifdef PODMAN_CMD
    CONTAINER_ENGINE := podman
endif

# Container configuration
CONTAINER_IMAGE ?= $(DOCKER_REGISTRY)/$(COMPONENT_NAME)
DOCKERFILE ?= Dockerfile
CONTAINER_BUILD_ARGS ?=
CONTAINER_RUN_ARGS ?=
CONTAINER_PUSH_REGISTRY ?= $(DOCKER_REGISTRY)

# Check if we have a container engine available
ifeq ($(CONTAINER_ENGINE),)
    $(warning No container engine found. Please install Docker or Podman.)
    CONTAINER_BUILD_AVAILABLE := false
else
    CONTAINER_BUILD_AVAILABLE := true
endif

.PHONY: docker-info docker-build docker-push docker-run docker-clean

# Display detected container engine
docker-info: ## Show container engine information
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	@echo "Container engine: $(CONTAINER_ENGINE)"
	@echo "Image: $(CONTAINER_IMAGE):$(VERSION)"
	@echo "Registry: $(CONTAINER_PUSH_REGISTRY)"
else
	@echo "No container engine detected (Docker or Podman required)"
endif

# Build container image
docker-build: ## Build container image
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	$(call show_progress,"Building container image $(CONTAINER_IMAGE):$(VERSION)")
	$(CONTAINER_ENGINE) build -t $(CONTAINER_IMAGE):$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
		--build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD) \
		$(CONTAINER_BUILD_ARGS) \
		-f $(DOCKERFILE) .
else
	@echo "Skipping container build - no container engine available"
endif

# Push container image
docker-push: docker-build ## Push container image to registry
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	$(call show_progress,"Pushing image $(CONTAINER_IMAGE):$(VERSION)")
	$(CONTAINER_ENGINE) push $(CONTAINER_IMAGE):$(VERSION)
else
	@echo "Skipping container push - no container engine available"
endif

# Run container
docker-run: docker-build ## Run container locally
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	$(call show_progress,"Running container $(CONTAINER_IMAGE):$(VERSION)")
	$(CONTAINER_ENGINE) run --rm -it \
		$(CONTAINER_RUN_ARGS) \
		$(CONTAINER_IMAGE):$(VERSION)
else
	@echo "Cannot run container - no container engine available"
endif

# Clean container artifacts
docker-clean: ## Clean container artifacts
ifeq ($(CONTAINER_BUILD_AVAILABLE),true)
	$(call show_progress,"Cleaning container artifacts")
	-$(CONTAINER_ENGINE) rmi $(CONTAINER_IMAGE):$(VERSION) 2>/dev/null || true
endif

endif # DOCKER_MK_INCLUDED