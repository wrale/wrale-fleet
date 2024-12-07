# Shared verification rules
include $(MAKEFILES_DIR)/common.mk

.PHONY: verify-security verify-deps verify-package verify-all

verify-security: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

verify-deps: ## Verify dependencies
	@echo "Checking dependencies..."
	go list -u -m -json all | go-mod-outdated -update
	govulncheck ./...

verify-package: ## Verify package contents
	@echo "Verifying package $(COMPONENT_NAME)-$(VERSION)..."
	@test -f $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz || \
		(echo "Package not found" && exit 1)
	@tar -tzf $(DIST_DIR)/$(COMPONENT_NAME)-$(VERSION).tar.gz

verify-all: verify-security verify-deps ## Run all verifications