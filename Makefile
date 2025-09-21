# Local GitOps CLI Makefile

.PHONY: build install test clean version help

# Version information
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Default target
help: ## Show this help message
	@echo "Local GitOps CLI - Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the CLI binary
	@echo "Building gitops CLI..."
	go build $(LDFLAGS) -o bin/gitops ./cmd/gitops
	@echo "✅ Binary built: bin/gitops"

install: ## Install the CLI using go install
	@echo "Installing gitops CLI..."
	go install $(LDFLAGS) ./cmd/gitops
	@echo "✅ CLI installed to $(shell go env GOPATH)/bin/gitops"

test: ## Run tests
	@echo "Running tests..."
	go test ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	@echo "✅ Cleaned build artifacts"

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# Development targets
dev-build: ## Build with development version
	@$(MAKE) build VERSION=dev

dev-install: ## Install with development version
	@$(MAKE) install VERSION=dev

# Release targets
release-build: ## Build release version
	@$(MAKE) build VERSION=$(VERSION)

release-install: ## Install release version
	@$(MAKE) install VERSION=$(VERSION)
