# Base component template
# Provides core functionality for all components
ifndef COMPONENT_MK_INCLUDED
COMPONENT_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk
include $(MAKEFILES_DIR)/core/vars.mk
include $(MAKEFILES_DIR)/core/targets.mk

# Required variables that components must define
ifndef COMPONENT_NAME
    $(call log_error,"COMPONENT_NAME is not defined")
    $(error COMPONENT_NAME must be set)
endif

# Optional variables with defaults
COMPONENT_DESCRIPTION ?= $(COMPONENT_NAME)
COMPONENT_VERSION ?= $(VERSION)
COMPONENT_TYPE ?= generic

# Directory structure
COMPONENT_DIR ?= .
BUILD_DIR ?= $(COMPONENT_DIR)/build
DIST_DIR ?= $(COMPONENT_DIR)/dist
DOC_DIR ?= $(COMPONENT_DIR)/docs
TEST_DIR ?= $(COMPONENT_DIR)/tests

# Template extension points - can be overridden by components
PREBUILD_HOOK ?= @:
POSTBUILD_HOOK ?= @:
PRETEST_HOOK ?= @:
POSTTEST_HOOK ?= @:
PRECLEAN_HOOK ?= @:
POSTCLEAN_HOOK ?= @:

# Define the base template functionality
define COMPONENT_TEMPLATE

.PHONY: validate init build test clean verify help
.DEFAULT_GOAL := help

# Component validation
validate: ## Validate component configuration
	$(call show_progress,"Validating $(COMPONENT_NAME)...")
	$(call check_var,COMPONENT_NAME)
	$(call check_var,COMPONENT_VERSION)
	$(call ensure_dir,$(BUILD_DIR))
	$(call ensure_dir,$(DIST_DIR))

# Component initialization
init: validate ## Initialize component
	$(call show_progress,"Initializing $(COMPONENT_NAME)...")
	$(call ensure_dir,$(DOC_DIR))
	$(call ensure_dir,$(TEST_DIR))

# Core lifecycle targets
build: validate ## Build the component
	$(call show_build_progress,$(COMPONENT_NAME))
	@$(PREBUILD_HOOK)
	@echo "Building $(COMPONENT_NAME) version $(COMPONENT_VERSION)"
	@$(POSTBUILD_HOOK)

test: validate ## Run component tests
	$(call show_test_progress,$(COMPONENT_NAME))
	@$(PRETEST_HOOK)
	@echo "Testing $(COMPONENT_NAME)"
	@$(POSTTEST_HOOK)

clean: ## Clean component artifacts
	$(call show_progress,"Cleaning $(COMPONENT_NAME)...")
	@$(PRECLEAN_HOOK)
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(POSTCLEAN_HOOK)

verify: validate test ## Verify component

# Help system
help: ## Show this help
	@echo "$(BOLD)$(COMPONENT_NAME) v$(COMPONENT_VERSION)$(RESET)"
	@echo "$(COMPONENT_DESCRIPTION)"
	@echo
	@echo "$(BOLD)Available targets:$(RESET)"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2}'

# Status reporting
status: ## Show component status
	@echo "$(BOLD)Component Status:$(RESET)"
	@echo "  Name: $(COMPONENT_NAME)"
	@echo "  Version: $(COMPONENT_VERSION)"
	@echo "  Type: $(COMPONENT_TYPE)"
	@echo "  Build dir: $(BUILD_DIR)"
	@echo "  Dist dir: $(DIST_DIR)"

# Mixin support
LOADED_MIXINS :=

# Include a mixin - Usage: $(call include_mixin,docker)
define include_mixin
    $(if $(filter $(1),$(LOADED_MIXINS)),,\
        $(eval LOADED_MIXINS += $(1))\
        $(call show_progress,"Loading mixin: $(1)")\
        $(eval include $(MAKEFILES_DIR)/mixins/$(1).mk))
endef

# List loaded mixins
mixins: ## List loaded mixins
	@echo "$(BOLD)Loaded Mixins:$(RESET)"
	@for mixin in $(LOADED_MIXINS); do \
		echo "  - $$mixin"; \
	done

endef # End of COMPONENT_TEMPLATE

endif # COMPONENT_MK_INCLUDED