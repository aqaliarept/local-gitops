# Go CLI Makefile

.PHONY: build install clean test deps

# Build the Go CLI
build:
	@echo "🔨 Building Go CLI..."
	@go build -o bin/gitops ./cmd/gitops
	@echo "✅ Build completed: bin/gitops"

# Install the CLI to GOPATH/bin
install: build
	@echo "📦 Installing CLI..."
	@cp bin/gitops $(GOPATH)/bin/gitops
	@echo "✅ CLI installed to $(GOPATH)/bin/gitops"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@echo "✅ Clean completed"

# Download dependencies
deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...
	@echo "✅ Tests completed"

# Run the CLI with help
help:
	@echo "📋 Available commands:"
	@echo "  build   - Build the Go CLI binary"
	@echo "  install - Install the CLI to GOPATH/bin"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Download and tidy dependencies"
	@echo "  test    - Run tests"
	@echo "  help    - Show this help message"
