# Common target definitions
# Only defines target names and dependencies, no commands

ifndef TARGETS_MK_INCLUDED
TARGETS_MK_INCLUDED := 1

# Basic targets that every component must implement
.PHONY: all build clean test verify

build: ## Build the component

clean: ## Clean build artifacts

test: ## Run component tests

verify: ## Run verifications

# The main target that builds everything
all: clean verify build ## Build everything (clean, verify, build)

endif # TARGETS_MK_INCLUDED