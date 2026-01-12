# gpasswd Makefile
# Go password manager CLI tool

# Variables
BINARY_NAME=gpasswd
VERSION=0.1.0-dev
BUILD_DIR=build
CMD_PATH=cmd/gpasswd/main.go
GO=go
GOFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"
GOTEST=$(GO) test
GOFMT=$(GO) fmt
GOVET=$(GO) vet
GOMOD=$(GO) mod

# Platform specific
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	PLATFORM=darwin
else ifeq ($(UNAME_S),Linux)
	PLATFORM=linux
else
	PLATFORM=unknown
endif

ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	GOARCH=amd64
else ifeq ($(ARCH),arm64)
	GOARCH=arm64
else
	GOARCH=$(ARCH)
endif

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test coverage fmt vet lint install uninstall run help deps tidy

# Default target
all: clean deps build test

## help: Show this help message
help:
	@echo "$(GREEN)gpasswd Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build          # Build the binary"
	@echo "  make test           # Run tests"
	@echo "  make install        # Install to /usr/local/bin"
	@echo "  make run            # Build and run"

## build: Build the binary for current platform
build:
	@echo "$(GREEN)Building $(BINARY_NAME) $(VERSION) for $(PLATFORM)/$(GOARCH)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-dev: Build without optimization (for debugging)
build-dev:
	@echo "$(GREEN)Building $(BINARY_NAME) (dev mode)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Build for multiple platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64

## build-darwin-amd64: Build for macOS (Intel)
build-darwin-amd64:
	@echo "$(GREEN)Building for macOS (Intel)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "$(GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64$(NC)"

## build-darwin-arm64: Build for macOS (Apple Silicon)
build-darwin-arm64:
	@echo "$(GREEN)Building for macOS (Apple Silicon)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "$(GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64$(NC)"

## build-linux-amd64: Build for Linux (amd64)
build-linux-amd64:
	@echo "$(GREEN)Building for Linux (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "$(GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(NC)"

## run: Build and run the application
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## tidy: Tidy go.mod and go.sum
tidy:
	@echo "$(GREEN)Tidying dependencies...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v -race ./...
	@echo "$(GREEN)✓ Tests complete$(NC)"

## test-short: Run short tests only
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) -short -v ./...
	@echo "$(GREEN)✓ Short tests complete$(NC)"

## coverage: Generate test coverage report
coverage:
	@echo "$(GREEN)Generating coverage report...$(NC)"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"
	@echo "$(YELLOW)Opening coverage report in browser...$(NC)"
	@if [ "$(PLATFORM)" = "darwin" ]; then \
		open coverage.html; \
	elif [ "$(PLATFORM)" = "linux" ]; then \
		xdg-open coverage.html 2>/dev/null || echo "Please open coverage.html manually"; \
	fi

## coverage-summary: Show test coverage summary
coverage-summary:
	@echo "$(GREEN)Generating coverage summary...$(NC)"
	@$(GOTEST) -cover ./... | grep coverage

## fmt: Format Go code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## lint: Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✓ Lint complete$(NC)"; \
	else \
		echo "$(RED)✗ golangci-lint not installed$(NC)"; \
		echo "$(YELLOW)Install with: brew install golangci-lint$(NC)"; \
		exit 1; \
	fi

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## install: Install binary to /usr/local/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed successfully$(NC)"
	@echo "$(YELLOW)Run '$(BINARY_NAME) --help' to get started$(NC)"

## uninstall: Remove binary from /usr/local/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled successfully$(NC)"

## install-deps: Install development dependencies
install-deps:
	@echo "$(GREEN)Installing development dependencies...$(NC)"
	@echo "$(YELLOW)Installing golangci-lint...$(NC)"
	@if [ "$(PLATFORM)" = "darwin" ]; then \
		brew install golangci-lint 2>/dev/null || echo "golangci-lint already installed"; \
	else \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@echo "$(YELLOW)Installing delve (debugger)...$(NC)"
	$(GO) install github.com/go-delve/delve/cmd/dlv@latest
	@echo "$(GREEN)✓ Development dependencies installed$(NC)"

## mod-upgrade: Upgrade all dependencies to latest versions
mod-upgrade:
	@echo "$(GREEN)Upgrading dependencies...$(NC)"
	$(GO) get -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies upgraded$(NC)"

## mod-vendor: Create vendor directory
mod-vendor:
	@echo "$(GREEN)Creating vendor directory...$(NC)"
	$(GOMOD) vendor
	@echo "$(GREEN)✓ Vendor directory created$(NC)"

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

## docker-build: Build Docker image (future)
docker-build:
	@echo "$(YELLOW)Docker support coming soon...$(NC)"

## release: Create a release build
release: clean
	@echo "$(GREEN)Creating release build $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@echo "$(YELLOW)Building for macOS (Intel)...$(NC)"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-darwin-amd64 $(CMD_PATH)
	@echo "$(YELLOW)Building for macOS (Apple Silicon)...$(NC)"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-$(VERSION)-darwin-arm64 $(CMD_PATH)
	@echo "$(YELLOW)Creating archives...$(NC)"
	@cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-amd64
	@cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-$(VERSION)-darwin-arm64
	@cd $(BUILD_DIR)/release && shasum -a 256 *.tar.gz > checksums.txt
	@echo "$(GREEN)✓ Release builds created in $(BUILD_DIR)/release/$(NC)"
	@ls -lh $(BUILD_DIR)/release/

## init: Initialize a test vault (for development)
init: build
	@echo "$(YELLOW)Initializing test vault...$(NC)"
	@echo "$(RED)WARNING: This will create a test vault in ~/.gpasswd-test$(NC)"
	@export HOME=$(shell pwd)/test-home && $(BUILD_DIR)/$(BINARY_NAME) init

## clean-vault: Remove test vault
clean-vault:
	@echo "$(YELLOW)Removing test vault...$(NC)"
	@rm -rf ~/.gpasswd-test
	@rm -rf test-home/.gpasswd
	@echo "$(GREEN)✓ Test vault removed$(NC)"

## watch: Watch for changes and rebuild (requires entr)
watch:
	@echo "$(GREEN)Watching for changes...$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make build; \
	else \
		echo "$(RED)✗ entr not installed$(NC)"; \
		echo "$(YELLOW)Install with: brew install entr$(NC)"; \
		exit 1; \
	fi

## size: Show binary size
size: build
	@echo "$(GREEN)Binary size:$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME) | awk '{print $$5 " " $$9}'
	@echo ""
	@echo "$(GREEN)Detailed size breakdown:$(NC)"
	@size $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || echo "size command not available"

## todo: Show TODO comments in code
todo:
	@echo "$(GREEN)TODO items:$(NC)"
	@grep -rn "TODO" --include="*.go" . || echo "No TODO items found"
	@echo ""
	@echo "$(GREEN)FIXME items:$(NC)"
	@grep -rn "FIXME" --include="*.go" . || echo "No FIXME items found"

## version: Show version
version:
	@echo "$(GREEN)Version: $(VERSION)$(NC)"
	@echo "$(GREEN)Platform: $(PLATFORM)/$(GOARCH)$(NC)"
	@echo "$(GREEN)Go version: $(shell $(GO) version)$(NC)"

## info: Show project information
info: version
	@echo ""
	@echo "$(GREEN)Project Information:$(NC)"
	@echo "  Binary name: $(BINARY_NAME)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Command path: $(CMD_PATH)"
	@echo ""
	@echo "$(GREEN)Dependencies:$(NC)"
	@$(GOMOD) graph | head -10

.DEFAULT_GOAL := help
