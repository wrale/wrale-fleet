# Metal Daemon Makefile
# Hardware access and control service implementation

.PHONY: all build test verify clean test-hw sim verify-hw

# Build configuration
BINARY_NAME := metald
BUILD_DIR := build

all: verify build test

build: ## Build metal daemon
	@echo "Building metal daemon..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/metald

test: ## Run tests
	@echo "Testing metal package..."
	@go test -race ./...

verify: ## Run verification
	@echo "Verifying metal package..."
	@go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning metal package..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@go clean

# Hardware-specific targets
test-hw: ## Run hardware tests (requires device)
	@echo "Running hardware tests..."
	@go test -v -tags=hardware ./...

sim: ## Run with hardware simulation
	@echo "Running simulated tests..."
	@go test -v -tags=simulation ./...

verify-hw: build ## Verify hardware setup
	@echo "Verifying hardware configuration..."
	@if [ -x "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		sudo $(BUILD_DIR)/$(BINARY_NAME) verify; \
	else \
		echo "Build required before verification"; \
		exit 1; \
	fi