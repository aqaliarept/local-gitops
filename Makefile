# Local GitOps CLI Makefile

.PHONY: build install test clean version help tag-patch push-tag

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

# Version management targets
tag-patch: ## Increment patch version (v0.0.1 -> v0.0.2)
	@$(eval CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"))
	@$(eval CURRENT_VERSION := $(shell echo $(CURRENT_TAG) | sed 's/v//'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print $$1"."$$2"."($$3+1)}'))
	@$(eval NEW_TAG := v$(NEW_VERSION))
	@echo "Current tag: $(CURRENT_TAG)"
	@echo "New tag: $(NEW_TAG)"
	@git tag -a $(NEW_TAG) -m "Release $(NEW_TAG): Patch version increment"
	@echo "✅ Created tag $(NEW_TAG)"

push-tag: ## Push the latest tag to remote
	@$(eval LATEST_TAG := $(shell git describe --tags --abbrev=0))
	@echo "Pushing tag $(LATEST_TAG) to remote..."
	@git push origin $(LATEST_TAG)
	@echo "✅ Tag $(LATEST_TAG) pushed to remote"

release-patch: tag-patch push-tag ## Create patch release (tag + push)
	@echo "✅ Patch release completed"
