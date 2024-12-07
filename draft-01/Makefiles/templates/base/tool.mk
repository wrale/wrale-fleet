# Base tool template
# Extends component template with tool-specific functionality
ifndef TOOL_MK_INCLUDED
TOOL_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/templates/base/component.mk

# Tool-specific configuration
COMPONENT_TYPE = tool
INSTALL_DIR ?= /usr/local/bin
MAN_DIR ?= /usr/local/share/man/man1
COMPLETION_DIR ?= /etc/bash_completion.d

# Define the tool template
define TOOL_TEMPLATE
$(COMPONENT_TEMPLATE)

.PHONY: install uninstall package release man completion

# Tool lifecycle management
install: build ## Install the tool
	$(call show_progress,"Installing $(COMPONENT_NAME)...")
	@sudo install -m 755 $(BUILD_DIR)/$(COMPONENT_NAME) $(INSTALL_DIR)/
	@if [ -f $(DOC_DIR)/$(COMPONENT_NAME).1 ]; then \
		sudo mkdir -p $(MAN_DIR); \
		sudo install -m 644 $(DOC_DIR)/$(COMPONENT_NAME).1 $(MAN_DIR)/; \
	fi
	@if [ -f $(DOC_DIR)/$(COMPONENT_NAME).completion ]; then \
		sudo install -m 644 $(DOC_DIR)/$(COMPONENT_NAME).completion \
			$(COMPLETION_DIR)/$(COMPONENT_NAME); \
	fi
	@echo "$(GREEN)Installation complete:$(RESET) $(INSTALL_DIR)/$(COMPONENT_NAME)"

uninstall: ## Uninstall the tool
	$(call show_progress,"Uninstalling $(COMPONENT_NAME)...")
	@sudo rm -f $(INSTALL_DIR)/$(COMPONENT_NAME)
	@sudo rm -f $(MAN_DIR)/$(COMPONENT_NAME).1
	@sudo rm -f $(COMPLETION_DIR)/$(COMPONENT_NAME)
	@echo "$(GREEN)Uninstallation complete$(RESET)"

package: build ## Create distributable package
	$(call show_progress,"Creating package for $(COMPONENT_NAME)...")
	@mkdir -p $(DIST_DIR)
	@cp $(BUILD_DIR)/$(COMPONENT_NAME) $(DIST_DIR)/
	@if [ -f $(DOC_DIR)/$(COMPONENT_NAME).1 ]; then \
		cp $(DOC_DIR)/$(COMPONENT_NAME).1 $(DIST_DIR)/; \
	fi
	@if [ -f $(DOC_DIR)/$(COMPONENT_NAME).completion ]; then \
		cp $(DOC_DIR)/$(COMPONENT_NAME).completion $(DIST_DIR)/; \
	fi
	@if [ -f README.md ]; then \
		cp README.md $(DIST_DIR)/; \
	fi
	@if [ -f LICENSE ]; then \
		cp LICENSE $(DIST_DIR)/; \
	fi
	@cd $(DIST_DIR) && tar czf $(COMPONENT_NAME)-$(COMPONENT_VERSION).tar.gz *
	@echo "$(GREEN)Package created:$(RESET) $(DIST_DIR)/$(COMPONENT_NAME)-$(COMPONENT_VERSION).tar.gz"

release: package ## Create and tag a release
	$(call show_progress,"Creating release $(COMPONENT_VERSION)...")
	@if [ -z "$(COMPONENT_VERSION)" ]; then \
		echo "$(RED)Error: Version not set$(RESET)"; \
		exit 1; \
	fi
	@if git rev-parse "v$(COMPONENT_VERSION)" >/dev/null 2>&1; then \
		echo "$(RED)Error: Version $(COMPONENT_VERSION) already exists$(RESET)"; \
		exit 1; \
	fi
	@git tag -a "v$(COMPONENT_VERSION)" -m "Release $(COMPONENT_VERSION)"
	@git push origin "v$(COMPONENT_VERSION)"
	@echo "$(GREEN)Release v$(COMPONENT_VERSION) created and pushed$(RESET)"

# Documentation targets
man: ## Generate man page
	$(call show_progress,"Generating man page...")
	@if [ -f $(DOC_DIR)/$(COMPONENT_NAME).1.md ]; then \
		if command -v pandoc >/dev/null 2>&1; then \
			pandoc $(DOC_DIR)/$(COMPONENT_NAME).1.md \
				-s -t man -o $(DOC_DIR)/$(COMPONENT_NAME).1; \
			echo "$(GREEN)Man page generated:$(RESET) $(DOC_DIR)/$(COMPONENT_NAME).1"; \
		else \
			echo "$(RED)Error: pandoc not found$(RESET)"; \
			exit 1; \
		fi; \
	else \
		echo "$(YELLOW)Warning: No man page source found$(RESET)"; \
	fi

completion: ## Generate shell completion
	$(call show_progress,"Generating completion script...")
	@if [ -x "$(BUILD_DIR)/$(COMPONENT_NAME)" ]; then \
		mkdir -p $(DOC_DIR); \
		$(BUILD_DIR)/$(COMPONENT_NAME) completion bash > $(DOC_DIR)/$(COMPONENT_NAME).completion; \
		echo "$(GREEN)Completion script generated:$(RESET) $(DOC_DIR)/$(COMPONENT_NAME).completion"; \
	else \
		echo "$(RED)Error: Tool must be built first$(RESET)"; \
		exit 1; \
	fi

# Version management
bump-major: ## Bump major version
	@current=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f1); \
	next=$$((current + 1)); \
	echo "$$next.0.0" > VERSION

bump-minor: ## Bump minor version
	@major=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f1); \
	current=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f2); \
	next=$$((current + 1)); \
	echo "$$major.$$next.0" > VERSION

bump-patch: ## Bump patch version
	@major=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f1); \
	minor=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f2); \
	current=$$(echo "$(COMPONENT_VERSION)" | cut -d. -f3); \
	next=$$((current + 1)); \
	echo "$$major.$$minor.$$next" > VERSION

# Tool-specific initialization
init: validate ## Initialize tool
	@$(MAKE) -s COMPONENT_TEMPLATE.init
	$(call ensure_dir,$(DOC_DIR))
	@if [ ! -f $(DOC_DIR)/$(COMPONENT_NAME).1.md ]; then \
		echo "Creating man page template..."; \
		echo "% $(COMPONENT_NAME)(1) Version $(COMPONENT_VERSION) | $(COMPONENT_DESCRIPTION)" \
			> $(DOC_DIR)/$(COMPONENT_NAME).1.md; \
		echo "# NAME\n$(COMPONENT_NAME) - $(COMPONENT_DESCRIPTION)" \
			>> $(DOC_DIR)/$(COMPONENT_NAME).1.md; \
	fi

endef # End of TOOL_TEMPLATE

endif # TOOL_MK_INCLUDED
