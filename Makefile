.PHONY: build build-release sign build-signed release-sign notarize-release test test-coverage coverage \
	lint lint-ci lint-fix clean install run run-verbose help cross-compile release \
	fmt vet check ci ci-coverage dev dev-check run-dev tidy release-preflight

# Binary name
BINARY_NAME=md-over-here

# Build directory
BUILD_DIR=bin
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
DIST_DIR=dist
VERSION ?= $(shell [ -f VERSION ] && cat VERSION || echo "dev")
MAIN_PKG ?= $(shell if [ -d ./cmd/$(BINARY_NAME) ]; then echo ./cmd/$(BINARY_NAME); else echo .; fi)
SIGN_IDENTITY ?= -
LDFLAGS ?= -ldflags "-s -w"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Lint
GOLANGCI_LINT_BIN ?= golangci-lint
GOLANGCI_LINT_VERSION ?= 2.10.1
LINT_TIMEOUT ?= 5m

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "md-over-here - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PKG)
	@echo "✓ Build complete: $(BINARY_PATH)"

build-release: ## Build optimized release binary for current platform
	@echo "Building release binary..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -trimpath -o $(BINARY_PATH) $(MAIN_PKG)
	@echo "✓ Release binary built: $(BINARY_PATH)"

sign: build ## Sign local binary (macOS only)
	@echo "Signing $(BINARY_PATH) (macOS only)..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		codesign -s "$(SIGN_IDENTITY)" -f "$(BINARY_PATH)"; \
		echo "✓ Signed: $(BINARY_PATH)"; \
	else \
		echo "ℹ️  Skipping sign (non-macOS)"; \
	fi

build-signed: build sign ## Build and sign local binary (macOS)

release-sign: ## Sign macOS release binaries in dist/ (optional; set RELEASE_SIGN=1)
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

notarize-release: ## Notarize release artifacts (optional; set RELEASE_NOTARIZE=1)
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

test: ## Run tests (deps + verify + go test)
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

test-coverage: ## Run tests with coverage summary
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

coverage: ## Generate HTML coverage report
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

lint: ## Lint code (requires golangci-lint)
	@echo "Linting code..."
	@which $(GOLANGCI_LINT_BIN) > /dev/null || (echo "$(GOLANGCI_LINT_BIN) not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@actual_version="$$($(GOLANGCI_LINT_BIN) version | awk '{print $$4}' | sed 's/^v//')"; \
	if [ "$$actual_version" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "golangci-lint version mismatch: required $(GOLANGCI_LINT_VERSION), found $$actual_version"; \
		exit 1; \
	fi
	@echo "Using golangci-lint $(GOLANGCI_LINT_VERSION)"
	$(GOLANGCI_LINT_BIN) run --timeout=$(LINT_TIMEOUT) ./...

lint-ci: ## Run linter in CI parity mode (clears cache first)
	@echo "Running CI-parity linter..."
	$(GOLANGCI_LINT_BIN) cache clean
	@$(MAKE) --no-print-directory lint

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	@which $(GOLANGCI_LINT_BIN) > /dev/null || (echo "$(GOLANGCI_LINT_BIN) not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@actual_version="$$($(GOLANGCI_LINT_BIN) version | awk '{print $$4}' | sed 's/^v//')"; \
	if [ "$$actual_version" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "golangci-lint version mismatch: required $(GOLANGCI_LINT_VERSION), found $$actual_version"; \
		exit 1; \
	fi
	@echo "Using golangci-lint $(GOLANGCI_LINT_VERSION)"
	$(GOLANGCI_LINT_BIN) run --timeout=$(LINT_TIMEOUT) --fix ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not found; skipping import formatting"; \
		echo "Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "✓ Vet complete"

check: ## Run full checks (fmt, vet, lint, test, build)
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

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

install: ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PKG)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

run: build ## Build and run (use ARGS='...' for arguments)
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_PATH) $(ARGS)

run-verbose: build ## Build and run with verbose flag
	@echo "Running $(BINARY_NAME) with verbose output..."
	./$(BINARY_PATH) -v $(ARGS)

dev: fmt lint test build ## Format, lint, test, and build

dev-check: fmt vet test ## Quick development check (fmt, vet, test)
	@echo "✓ Development checks passed"

run-dev: ## Run with hot reload (requires air)
	@which air > /dev/null || (echo "Air not installed. Install: go install github.com/cosmtrek/air@latest" && exit 1)
	@echo "Starting hot reload..."
	air

ci: check ## Run CI checks (full pipeline)
	@echo "✓ CI checks passed"

ci-coverage: fmt lint test-coverage build ## Legacy CI coverage flow (fmt, lint, coverage, build)
	@echo "✓ CI coverage flow complete"

cross-compile: ## Cross-compile for release artifacts
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	@echo "✓ Cross-compilation complete"
	@ls -lh $(DIST_DIR)/

release: clean test cross-compile ## Create release build with tarballs and checksums
	@echo "Creating release archives for $(VERSION)..."
	@cd $(DIST_DIR) && \
		tar -czf $(BINARY_NAME)-darwin-arm64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64 && \
		tar -czf $(BINARY_NAME)-darwin-amd64-$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf $(BINARY_NAME)-linux-amd64-$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf $(BINARY_NAME)-linux-arm64-$(VERSION).tar.gz $(BINARY_NAME)-linux-arm64
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "✓ Release $(VERSION) ready in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/*.tar.gz

release-preflight: ## Run release checks against a local tag (usage: make release-preflight TAG=1.0.0)
	@[ -n "$(TAG)" ] || (echo "TAG is required (example: make release-preflight TAG=1.0.0)" && exit 1)
	@git rev-parse --verify --quiet "refs/tags/$(TAG)" >/dev/null || (echo "Tag not found: $(TAG)" && exit 1)
	@git diff --quiet && git diff --cached --quiet || (echo "Working tree must be clean for release-preflight" && exit 1)
	@tmp_dir="$$(mktemp -d)"; \
	echo "Using temporary worktree: $$tmp_dir"; \
	trap 'git worktree remove --force "$$tmp_dir" >/dev/null 2>&1 || true' EXIT; \
	git worktree add --detach "$$tmp_dir" "refs/tags/$(TAG)" >/dev/null; \
	cd "$$tmp_dir"; \
	$(MAKE) --no-print-directory test; \
	$(MAKE) --no-print-directory lint-ci
