.PHONY: build test lint clean install run help cross-compile release

# Binary name
BINARY_NAME=md-over-here

# Build directory
BUILD_DIR=.
DIST_DIR=dist
VERSION ?= $(shell [ -f VERSION ] && cat VERSION || echo "dev")
MAIN_PKG ?= $(shell if [ -d ./cmd/$(BINARY_NAME) ]; then echo ./cmd/$(BINARY_NAME); else echo .; fi)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Building main package..."
	$(GOBUILD) -o $(BUILD_DIR)/.build-check $(MAIN_PKG)
	@rm -f $(BUILD_DIR)/.build-check

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

# Run tests with detailed coverage report
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the code
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PKG)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Run the binary (pass arguments via ARGS variable)
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) $(ARGS)

# Run with verbose flag (for testing)
run-verbose: build
	@echo "Running $(BINARY_NAME) with verbose output..."
	./$(BINARY_NAME) -v $(ARGS)

# Development - build and run tests
dev: fmt lint test build

# CI/CD - comprehensive checks
ci: fmt lint test-coverage build

# Cross-compile for release artifacts
cross-compile:
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	@echo "Cross-compilation complete"
	@ls -lh $(DIST_DIR)/

# Release build with tarballs and checksums
release: clean test cross-compile
	@echo "Creating release archives for $(VERSION)..."
	@cd $(DIST_DIR) && \
		tar -czf $(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64 && \
		tar -czf $(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz $(BINARY_NAME)-linux-arm64
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "Release $(VERSION) ready in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/*.tar.gz

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage summary"
	@echo "  coverage       - Generate HTML coverage report"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy dependencies"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Build and run (use ARGS='...' for arguments)"
	@echo "  run-verbose    - Build and run with verbose flag"
	@echo "  dev            - Format, lint, test, and build"
	@echo "  ci             - Run CI checks (fmt, lint, coverage, build)"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make lint"
	@echo "  make run ARGS='https://example.com'"
	@echo "  make dev"

# Default target
.DEFAULT_GOAL := help
