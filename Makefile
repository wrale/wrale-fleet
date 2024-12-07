# Root Makefile - Project orchestration
COMPONENTS := fleet fleet/edge metal/core user/api user/ui/wrale-dashboard

.PHONY: all clean test verify help $(COMPONENTS)

all: $(COMPONENTS) ## Build all components

$(COMPONENTS):
	$(MAKE) -C $@

clean test verify:
	@for dir in $(COMPONENTS); do \
		echo "Running $@ in $$dir..."; \
		$(MAKE) -C $$dir $@ || exit 1; \
	done

help: ## Show this help
	@echo "Main targets for all components:"
	@echo "  all     - Build all components"
	@echo "  clean   - Clean all components"
	@echo "  test    - Test all components"
	@echo "  verify  - Verify all components"
	@echo "\nComponent-specific targets:"
	@for dir in $(COMPONENTS); do \
		echo "\n$$dir:"; \
		$(MAKE) -C $$dir help; \
	done