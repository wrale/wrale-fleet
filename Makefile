# Wrale Fleet Management Platform Build System
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
BINARY_NAME=fleetd

# Tools
GOLINT=golangci-lint
GOSEC=gosec

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test coverage lint sec-check vet fmt help

help: ## Display this help message
	@echo "Wrale Fleet Management Platform - Make Targets"
	@echo
	@echo "Usage:"
	@awk 'BEGIN {FS = ":.*##"; printf "  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

fmt: ## Format code using gofmt
	@echo "==> Formatting code"
	@go fmt ./...

vet: fmt ## Run go vet
	@echo "==> Running go vet"
	$(GOVET) ./...

lint: vet ## Run golangci-lint
	@echo "==> Running golangci-lint"
	$(GOLINT) run

sec-check: ## Run security checks
	@echo "==> Running security checks"
	$(GOSEC) ./...

test: ## Run tests
	@echo "==> Running tests"
	$(GOTEST) -v -race ./...

coverage: ## Generate test coverage report
	@echo "==> Generating coverage report"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

build: ## Build the binary
	@echo "==> Building $(BINARY_NAME)"
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/fleetd

install-tools: ## Install required development tools
	@echo "==> Installing development tools"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

all: fmt vet lint sec-check test build ## Run all checks and build

# Development targets
run: build ## Run the application
	@echo "==> Running $(BINARY_NAME)"
	./$(BINARY_NAME)

dev: ## Run the application with hot reload
	@echo "==> Starting development server"
	air -c .air.toml

.PHONY: all build clean test coverage lint sec-check vet fmt help install-tools run dev