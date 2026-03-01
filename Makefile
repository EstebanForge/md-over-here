.PHONY: build sign build-signed release-sign notarize-release test lint clean install run help cross-compile release fmt vet check ci ci-coverage

# Binary name
BINARY_NAME=md-over-here

# Build directory
BUILD_DIR=bin
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
DIST_DIR=dist
VERSION ?= $(shell [ -f VERSION ] && cat VERSION || echo "dev")
MAIN_PKG ?= $(shell if [ -d ./cmd/$(BINARY_NAME) ]; then echo ./cmd/$(BINARY_NAME); else echo .; fi)
SIGN_IDENTITY ?= -

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
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PKG)
	@echo "Build complete: $(BINARY_PATH)"

sign: build
	@echo "Signing $(BINARY_PATH) (macOS only)..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		codesign -s "$(SIGN_IDENTITY)" -f "$(BINARY_PATH)"; \
		echo "✓ Signed: $(BINARY_PATH)"; \
	else \
		echo "Skipping sign (non-macOS)"; \
	fi

build-signed: build sign

release-sign:
	@if [ "$${RELEASE_SIGN:-0}" != "1" ]; then \
		echo "RELEASE_SIGN!=1; skipping release signing"; \
	elif [ "$$(uname)" != "Darwin" ]; then \
		echo "Release signing requires a macOS runner; skipping"; \
	else \
		for bin in "$(DIST_DIR)/$(BINARY_NAME)-darwin-amd64" "$(DIST_DIR)/$(BINARY_NAME)-darwin-arm64"; do \
			if [ -f "$$bin" ]; then \
				codesign -s "$(SIGN_IDENTITY)" -f "$$bin"; \
				echo "✓ Signed $$bin"; \
			else \
				echo "Missing $$bin (skip)"; \
			fi; \
		done; \
	fi

notarize-release:
	@if [ "$${RELEASE_NOTARIZE:-0}" != "1" ]; then \
		echo "RELEASE_NOTARIZE!=1; skipping notarization"; \
	elif [ "$$(uname)" != "Darwin" ]; then \
		echo "Notarization requires macOS runner"; \
		exit 1; \
	elif [ -x "./scripts/notarize-release.sh" ]; then \
		./scripts/notarize-release.sh; \
	else \
		echo "scripts/notarize-release.sh not found/executable"; \
		exit 1; \
	fi

# Run tests
test:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Running tests..."
	@if [ "$$($(GOCMD) env CGO_ENABLED)" = "1" ]; then \
		$(GOTEST) -v -race ./...; \
	else \
		echo "CGO disabled; running tests without -race"; \
		$(GOTEST) -v ./...; \
	fi
	@echo "Building main package..."
	@mkdir -p $(BUILD_DIR)
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
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found; skipping import formatting"; \
		echo "Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run full checks
check:
	@echo "==> make fmt"
	@$(MAKE) --no-print-directory fmt
	@echo "==> make vet"
	@$(MAKE) --no-print-directory vet
	@echo "==> make lint"
	@$(MAKE) --no-print-directory lint
	@echo "==> make test"
	@$(MAKE) --no-print-directory test
	@echo "==> make build"
	@$(MAKE) --no-print-directory build
	@echo "✓ Full checks complete"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
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
	./$(BINARY_PATH) $(ARGS)

# Run with verbose flag (for testing)
run-verbose: build
	@echo "Running $(BINARY_NAME) with verbose output..."
	./$(BINARY_PATH) -v $(ARGS)

# Development - build and run tests
dev: fmt lint test build

# CI/CD - standardized checks
ci: check
	@echo "CI checks passed"

# Legacy CI/CD coverage-heavy flow
ci-coverage: fmt lint test-coverage build
	@echo "CI coverage flow complete"

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
	@echo "  sign           - Sign local binary (macOS only)"
	@echo "  build-signed   - Build and sign local binary (macOS)"
	@echo "  release-sign   - Sign macOS release binaries (optional)"
	@echo "  notarize-release - Notarize release binaries (optional)"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage summary"
	@echo "  coverage       - Generate HTML coverage report"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  check          - Run full checks (fmt, vet, lint, test, build)"
	@echo "  tidy           - Tidy dependencies"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Build and run (use ARGS='...' for arguments)"
	@echo "  run-verbose    - Build and run with verbose flag"
	@echo "  dev            - Format, lint, test, and build"
	@echo "  ci             - Run CI checks (full pipeline = make check)"
	@echo "  ci-coverage    - Run legacy CI coverage flow (fmt, lint, coverage, build)"
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
