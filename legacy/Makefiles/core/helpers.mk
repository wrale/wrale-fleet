# Core helper functions for the Makefile system
ifndef HELPERS_MK_INCLUDED
HELPERS_MK_INCLUDED := 1

# Terminal colors
ifdef TERM
    BLUE := $(shell tput setaf 4)
    GREEN := $(shell tput setaf 2)
    YELLOW := $(shell tput setaf 3)
    RED := $(shell tput setaf 1)
    RESET := $(shell tput sgr0)
    BOLD := $(shell tput bold)
else
    BLUE := 
    GREEN := 
    YELLOW := 
    RED := 
    RESET := 
    BOLD := 
endif

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
    OPEN_CMD := xdg-open
else ifeq ($(UNAME_S),Darwin)
    PLATFORM := darwin
    OPEN_CMD := open
else
    PLATFORM := unknown
    OPEN_CMD := echo
endif

# Command availability checks 
HAS_DOCKER := $(shell command -v docker 2> /dev/null)
HAS_GO := $(shell command -v go 2> /dev/null)
HAS_NODE := $(shell command -v node 2> /dev/null)
HAS_PYTHON := $(shell command -v python3 2> /dev/null)

# Logging functions
define log_info
    @echo "$(BLUE)$(BOLD)INFO:$(RESET) $(1)"
endef

define log_success
    @echo "$(GREEN)$(BOLD)SUCCESS:$(RESET) $(1)"
endef

define log_warn
    @echo "$(YELLOW)$(BOLD)WARNING:$(RESET) $(1)"
endef

define log_error
    @echo "$(RED)$(BOLD)ERROR:$(RESET) $(1)" >&2
endef

# Variable validation
define check_var
    @if [ -z "$($(1))" ]; then \
        $(call log_error,"Required variable $(1) is not set"); \
        exit 1; \
    fi
endef

# Directory operations
define ensure_dir
    @mkdir -p $(1)
endef

define dir_exists
    @test -d $(1) || ($(call log_error,"Directory $(1) does not exist") && exit 1)
endef

# File operations
define ensure_file
    @touch $(1)
endef

define file_exists
    @test -f $(1) || ($(call log_error,"File $(1) does not exist") && exit 1)
endef

# Command validation
define check_cmd
    @command -v $(1) > /dev/null || ($(call log_error,"Command $(1) not found") && exit 1)
endef

# Help function generation
define HELP_FUNCTION
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
    awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
endef

# Progress indication
define show_progress
    @echo "$(BLUE)$(BOLD)→$(RESET) $(1)"
endef

define show_build_progress
    @echo "$(GREEN)$(BOLD)⚡$(RESET) Building $(1)..."
endef

define show_test_progress
    @echo "$(YELLOW)$(BOLD)§$(RESET) Testing $(1)..."
endef

# Version comparison
define version_gt
    @awk 'BEGIN{exit !( "$(1)" > "$(2)" )}' || exit 1
endef

# Path manipulation
define get_abs_path
    $(shell realpath $(1))
endef

define join_paths
    $(shell echo "$(1)/$(2)" | sed 's#//#/#g')
endef

endif # HELPERS_MK_INCLUDED
