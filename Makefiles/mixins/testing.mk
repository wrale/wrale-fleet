# Testing mixin providing advanced testing capabilities
ifndef TESTING_MK_INCLUDED
TESTING_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk

# Test configuration
TEST_COVERAGE_MIN ?= 80
TEST_TIMEOUT ?= 5m
TEST_PARALLEL ?= 4
TEST_VERBOSE ?= false
TEST_OUTPUT_DIR ?= test-results

# Hardware test configuration
HW_TEST_ENABLED ?= false
HW_TEST_DEVICE ?= 
HW_TEST_TIMEOUT ?= 10m

.PHONY: test-integration test-performance test-hardware test-coverage test-report

# Integration testing
test-integration: ## Run integration tests
	$(call show_test_progress,"Running integration tests")
	@mkdir -p $(TEST_OUTPUT_DIR)
	@if [ "$(TEST_VERBOSE)" = "true" ]; then \
		go test -v -tags=integration -parallel=$(TEST_PARALLEL) \
			-timeout=$(TEST_TIMEOUT) ./... \
			-json > $(TEST_OUTPUT_DIR)/integration.json; \
	else \
		go test -tags=integration -parallel=$(TEST_PARALLEL) \
			-timeout=$(TEST_TIMEOUT) ./...; \
	fi

# Performance testing
test-performance: ## Run performance tests
	$(call show_test_progress,"Running performance tests")
	@mkdir -p $(TEST_OUTPUT_DIR)
	@go test -tags=performance -run=XXX -bench=. -benchmem \
		./... > $(TEST_OUTPUT_DIR)/performance.txt

# Hardware testing
test-hardware: ## Run hardware tests (requires device)
ifeq ($(HW_TEST_ENABLED),true)
	@if [ -z "$(HW_TEST_DEVICE)" ]; then \
		echo "$(RED)Error: HW_TEST_DEVICE must be set$(RESET)"; \
		exit 1; \
	fi
	$(call show_test_progress,"Running hardware tests on $(HW_TEST_DEVICE)")
	@go test -v -tags=hardware -timeout=$(HW_TEST_TIMEOUT) \
		-args -device=$(HW_TEST_DEVICE) ./...
else
	@echo "Hardware testing is disabled (HW_TEST_ENABLED=false)"
endif

# Coverage reporting
test-coverage: ## Generate test coverage report
	$(call show_test_progress,"Generating coverage report")
	@mkdir -p $(TEST_OUTPUT_DIR)
	@go test -coverprofile=$(TEST_OUTPUT_DIR)/coverage.out ./...
	@go tool cover -html=$(TEST_OUTPUT_DIR)/coverage.out \
		-o $(TEST_OUTPUT_DIR)/coverage.html
	@coverage=$$(go tool cover -func=$(TEST_OUTPUT_DIR)/coverage.out | \
		grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $${coverage%.*} -lt $(TEST_COVERAGE_MIN) ]; then \
		echo "$(RED)Coverage $${coverage}% below minimum $(TEST_COVERAGE_MIN)%$(RESET)"; \
		exit 1; \
	else \
		echo "$(GREEN)Coverage: $${coverage}%$(RESET)"; \
	fi

# Test report generation
test-report: test-coverage test-performance ## Generate comprehensive test report
	$(call show_progress,"Generating test report")
	@echo "Test Report" > $(TEST_OUTPUT_DIR)/report.txt
	@echo "===========" >> $(TEST_OUTPUT_DIR)/report.txt
	@echo >> $(TEST_OUTPUT_DIR)/report.txt
	@echo "Coverage Report:" >> $(TEST_OUTPUT_DIR)/report.txt
	@go tool cover -func=$(TEST_OUTPUT_DIR)/coverage.out >> $(TEST_OUTPUT_DIR)/report.txt
	@echo >> $(TEST_OUTPUT_DIR)/report.txt
	@echo "Performance Report:" >> $(TEST_OUTPUT_DIR)/report.txt
	@cat $(TEST_OUTPUT_DIR)/performance.txt >> $(TEST_OUTPUT_DIR)/report.txt
	@echo "$(GREEN)Report generated:$(RESET) $(TEST_OUTPUT_DIR)/report.txt"

endif # TESTING_MK_INCLUDED