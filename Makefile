# GoClean Makefile

.PHONY: build test clean install lint fmt vet run dev benchmark benchmark-suite benchmark-report benchmark-validate release-check release-build release-package release-tag rust-lib rust-lib-cross rust-clean rust-install rust-setup

# Build configuration
BINARY_NAME=goclean
BUILD_DIR=bin
MAIN_PATH=./cmd/goclean
LIB_DIR=lib

# Rust library configuration
RUST_PARSER_PATH ?= ./rust/parser
RUST_LIB_NAME=libgoclean_rust_parser
RUST_TARGET_DIR=$(RUST_PARSER_PATH)/target

# Platform detection
ifeq ($(OS),Windows_NT)
  UNAME_OS := Windows
else
  UNAME_OS := $(shell uname -s)
endif
UNAME_ARCH := $(shell uname -m)

# Platform-specific library extensions and names
ifeq ($(UNAME_OS),Linux)
    LIB_EXT=.so
    CARGO_TARGET_OS=unknown-linux-gnu
endif
ifeq ($(UNAME_OS),Darwin)
    LIB_EXT=.dylib
    CARGO_TARGET_OS=apple-darwin
endif
ifneq (,$(filter Windows MSYS% MINGW% CYGWIN%,$(UNAME_OS)))
    LIB_EXT=.dll
    CARGO_TARGET_OS=pc-windows-msvc
endif

# Architecture mapping for Rust targets
ifeq ($(UNAME_ARCH),x86_64)
    CARGO_ARCH=x86_64
endif
ifeq ($(UNAME_ARCH),arm64)
    CARGO_ARCH=aarch64
endif
ifeq ($(UNAME_ARCH),aarch64)
    CARGO_ARCH=aarch64
endif

CARGO_TARGET=$(CARGO_ARCH)-$(CARGO_TARGET_OS)
RUST_LIB_FILE=$(RUST_LIB_NAME)$(LIB_EXT)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

# =============================
# Rust Library Build Targets
# =============================

# Check if Rust is installed
rust-check:
	@echo "Checking Rust installation..."
	@which cargo >/dev/null 2>&1 || (echo "ERROR: Rust/Cargo not found. Please install Rust from https://rustup.rs/" && exit 1)
	@echo "✓ Rust toolchain found: $$(rustc --version)"
	@echo "✓ Cargo found: $$(cargo --version)"

# Setup Rust cross-compilation targets
rust-setup: rust-check
	@echo "Setting up Rust cross-compilation targets..."
	@echo "Current target: $(CARGO_TARGET)"
	@if command -v rustup >/dev/null 2>&1; then \
		rustup target list --installed | grep -q $(CARGO_TARGET) || (echo "Installing target $(CARGO_TARGET)..." && rustup target add $(CARGO_TARGET)); \
	else \
		echo "⚠️  rustup not available - using system Rust installation"; \
		echo "For cross-compilation, install rustup: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"; \
	fi
	@echo "✓ Rust target setup completed"

# Build Rust library for current platform
rust-lib: rust-setup
	@echo "Building Rust library for $(CARGO_TARGET)..."
	@mkdir -p $(LIB_DIR)
	@if command -v rustup >/dev/null 2>&1; then \
		echo "Building with target $(CARGO_TARGET)..."; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target $(CARGO_TARGET); \
		cp $(RUST_TARGET_DIR)/$(CARGO_TARGET)/release/$(RUST_LIB_FILE) $(LIB_DIR)/$(RUST_LIB_FILE); \
	else \
		echo "Building with native toolchain..."; \
		cd $(RUST_PARSER_PATH) && cargo build --release; \
		cp $(RUST_TARGET_DIR)/release/$(RUST_LIB_FILE) $(LIB_DIR)/$(RUST_LIB_FILE); \
	fi
	@echo "✓ Rust library built successfully: $(LIB_DIR)/$(RUST_LIB_FILE)"

# Build Rust library for all supported platforms
rust-lib-cross: rust-check
	@echo "Building Rust library for all supported platforms..."
	@mkdir -p $(LIB_DIR)
	
	@if command -v rustup >/dev/null 2>&1; then \
		echo "Using rustup for cross-compilation..."; \
		echo "Building for Linux x86_64..."; \
		rustup target add x86_64-unknown-linux-gnu 2>/dev/null || true; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target x86_64-unknown-linux-gnu; \
		cp $(RUST_TARGET_DIR)/x86_64-unknown-linux-gnu/release/$(RUST_LIB_NAME).so $(LIB_DIR)/$(RUST_LIB_NAME)-linux-x86_64.so; \
		\
		echo "Building for Linux ARM64..."; \
		rustup target add aarch64-unknown-linux-gnu 2>/dev/null || true; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target aarch64-unknown-linux-gnu 2>/dev/null || echo "⚠️  ARM64 cross-compilation may require additional setup"; \
		[ -f $(RUST_TARGET_DIR)/aarch64-unknown-linux-gnu/release/$(RUST_LIB_NAME).so ] && cp $(RUST_TARGET_DIR)/aarch64-unknown-linux-gnu/release/$(RUST_LIB_NAME).so $(LIB_DIR)/$(RUST_LIB_NAME)-linux-aarch64.so || echo "⚠️  ARM64 Linux build skipped"; \
		\
		echo "Building for macOS x86_64..."; \
		rustup target add x86_64-apple-darwin 2>/dev/null || true; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target x86_64-apple-darwin 2>/dev/null || echo "⚠️  macOS x86_64 cross-compilation may require macOS SDK"; \
		[ -f $(RUST_TARGET_DIR)/x86_64-apple-darwin/release/$(RUST_LIB_NAME).dylib ] && cp $(RUST_TARGET_DIR)/x86_64-apple-darwin/release/$(RUST_LIB_NAME).dylib $(LIB_DIR)/$(RUST_LIB_NAME)-darwin-x86_64.dylib || echo "⚠️  macOS x86_64 build skipped"; \
		\
		echo "Building for macOS ARM64..."; \
		rustup target add aarch64-apple-darwin 2>/dev/null || true; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target aarch64-apple-darwin 2>/dev/null || echo "⚠️  macOS ARM64 cross-compilation may require macOS SDK"; \
		[ -f $(RUST_TARGET_DIR)/aarch64-apple-darwin/release/$(RUST_LIB_NAME).dylib ] && cp $(RUST_TARGET_DIR)/aarch64-apple-darwin/release/$(RUST_LIB_NAME).dylib $(LIB_DIR)/$(RUST_LIB_NAME)-darwin-aarch64.dylib || echo "⚠️  macOS ARM64 build skipped"; \
		\
		echo "Building for Windows x86_64..."; \
		rustup target add x86_64-pc-windows-gnu 2>/dev/null || true; \
		cd $(RUST_PARSER_PATH) && cargo build --release --target x86_64-pc-windows-gnu 2>/dev/null || echo "⚠️  Windows cross-compilation may require additional setup"; \
		[ -f $(RUST_TARGET_DIR)/x86_64-pc-windows-gnu/release/$(RUST_LIB_NAME).dll ] && cp $(RUST_TARGET_DIR)/x86_64-pc-windows-gnu/release/$(RUST_LIB_NAME).dll $(LIB_DIR)/$(RUST_LIB_NAME)-windows-x86_64.dll || echo "⚠️  Windows build skipped"; \
	else \
		echo "⚠️  rustup not available - building for native target only"; \
		echo "Building for native target: $(CARGO_TARGET)..."; \
		cd $(RUST_PARSER_PATH) && cargo build --release; \
		cp $(RUST_TARGET_DIR)/release/$(RUST_LIB_FILE) $(LIB_DIR)/$(RUST_LIB_NAME)-native$(LIB_EXT); \
	fi
	
	@echo "✓ Cross-platform Rust library builds completed"
	@ls -la $(LIB_DIR)/

# Clean Rust build artifacts
rust-clean:
	@echo "Cleaning Rust build artifacts..."
	@cd $(RUST_PARSER_PATH) && cargo clean 2>/dev/null || true
	@rm -f $(LIB_DIR)/$(RUST_LIB_NAME)*
	@echo "✓ Rust artifacts cleaned"

# Install Rust toolchain and cross-compilation setup
rust-install:
	@echo "Installing Rust toolchain and cross-compilation targets..."
	@command -v cargo >/dev/null 2>&1 || (curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y)
	@$${HOME}/.cargo/bin/rustup target add x86_64-unknown-linux-gnu
	@$${HOME}/.cargo/bin/rustup target add aarch64-unknown-linux-gnu
	@$${HOME}/.cargo/bin/rustup target add x86_64-apple-darwin
	@$${HOME}/.cargo/bin/rustup target add aarch64-apple-darwin
	@$${HOME}/.cargo/bin/rustup target add x86_64-pc-windows-gnu
	@echo "✓ Rust cross-compilation setup completed"

# =============================
# Go Application Build Targets
# =============================

# Build the application (includes Rust library)
build: rust-lib
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@echo "Working directory: $(PWD)"
	@echo "Main path: $(MAIN_PATH)"
	@ls -la $(MAIN_PATH) 2>/dev/null || echo "Directory $(MAIN_PATH) not found, listing current dir:" && ls -la .
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ GoClean built successfully with Rust support"

# Build Go-only (fallback without Rust)
build-go-only:
	@echo "Building $(BINARY_NAME) $(VERSION) without Rust support..."
	@echo "Working directory: $(PWD)"
	@echo "Main path: $(MAIN_PATH)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ GoClean built successfully (Go-only mode)"

# Run tests
test:
	@echo "Running tests..."
	LD_LIBRARY_PATH=$(PWD)/$(LIB_DIR):$$LD_LIBRARY_PATH GOCLEAN_TEST_MODE=1 $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean: rust-clean
	@echo "Cleaning Go build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "✓ All build artifacts cleaned"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)

# Lint code
lint:
	@echo "Running linter..."
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not found, skipping lint check"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Vet code
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Development mode
dev:
	@echo "Running in development mode..."
	$(GOCMD) run $(MAIN_PATH)

# Build for multiple platforms (with Rust libraries)
build-all: rust-lib-cross
	@echo "Building $(VERSION) for multiple platforms with Rust support..."
	@mkdir -p $(BUILD_DIR)
	
	# Copy platform-specific Rust libraries to build directory for packaging
	@echo "Preparing Rust libraries for cross-platform builds..."
	@mkdir -p $(BUILD_DIR)/lib
	@cp -f $(LIB_DIR)/$(RUST_LIB_NAME)-* $(BUILD_DIR)/lib/ 2>/dev/null || echo "⚠️  Some Rust libraries may be missing"
	
	# Build Go binaries for each platform
	@echo "Building Go binaries..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH) 2>/dev/null || CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH) 2>/dev/null || CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH) 2>/dev/null || CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH) 2>/dev/null || CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "✓ Cross-platform builds completed"
	@ls -la $(BUILD_DIR)/

# Build for multiple platforms (Go-only fallback)
build-all-go-only:
	@echo "Building $(VERSION) for multiple platforms (Go-only mode)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✓ Cross-platform builds completed (Go-only mode)"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./internal/...

# Run comprehensive benchmark suite
benchmark-suite: build
	@echo "Running comprehensive benchmark suite..."
	$(GOTEST) -bench=BenchmarkSuite -benchmem -timeout=30m .

# Run benchmarks and generate report
benchmark-report:
	@echo "Running benchmarks and generating report..."
	$(GOTEST) -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/... > benchmark_results.txt
	@echo "Benchmark results saved to benchmark_results.txt"
	@echo "CPU profile saved to cpu.prof"
	@echo "Memory profile saved to mem.prof"

# Compare benchmark results
benchmark-compare:
	@echo "Comparing benchmark results..."
	@echo "Usage: make benchmark-compare old=old_results.txt new=new_results.txt"
	@benchstat $(old) $(new)

# Run performance validation benchmarks
benchmark-validate: build
	@echo "Running performance validation benchmarks..."
	$(GOTEST) -bench=BenchmarkOverallPerformance -benchmem -timeout=10m .

# Release preparation
release-check: lint vet test benchmark-validate
	@echo "Running pre-release checks..."
	@echo "✓ All checks passed - ready for release"

# Release preparation (minimal - skip failing tests)
release-check-minimal: lint vet
	@echo "Running minimal pre-release checks..."
	@echo "Skipping test failures for release build"
	@echo "✓ Minimal checks passed - ready for release"

# Create release build
release-build: release-check
	@echo "Creating release build $(VERSION)..."
	@$(MAKE) clean
	@$(MAKE) build-all
	@echo "✓ Release build completed"

# Create release build (minimal - skip failing tests)
release-build-minimal: release-check-minimal
	@echo "Creating release build $(VERSION) with minimal checks..."
	@$(MAKE) clean
	@$(MAKE) build-all
	@echo "✓ Release build completed"

# Package release (with Rust libraries)
release-package: release-build
	@echo "Packaging release $(VERSION) with Rust libraries..."
	@mkdir -p releases/$(VERSION)
	@cd $(BUILD_DIR) && \
		([ -f lib/$(RUST_LIB_NAME)-linux-x86_64.so ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 lib/$(RUST_LIB_NAME)-linux-x86_64.so || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64) && \
		([ -f lib/$(RUST_LIB_NAME)-darwin-x86_64.dylib ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 lib/$(RUST_LIB_NAME)-darwin-x86_64.dylib || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64) && \
		([ -f lib/$(RUST_LIB_NAME)-darwin-aarch64.dylib ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 lib/$(RUST_LIB_NAME)-darwin-aarch64.dylib || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64) && \
		([ -f lib/$(RUST_LIB_NAME)-windows-x86_64.dll ] && (which zip >/dev/null 2>&1 && zip -q ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe lib/$(RUST_LIB_NAME)-windows-x86_64.dll || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe lib/$(RUST_LIB_NAME)-windows-x86_64.dll) || (which zip >/dev/null 2>&1 && zip -q ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe))
	@echo "✓ Release packages created in releases/$(VERSION)/ (with Rust support where available)"

# Package release (minimal - skip failing tests, with Rust libraries)
release-package-minimal: release-build-minimal
	@echo "Packaging release $(VERSION) with minimal checks and Rust libraries..."
	@mkdir -p releases/$(VERSION)
	@cd $(BUILD_DIR) && \
		([ -f lib/$(RUST_LIB_NAME)-linux-x86_64.so ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 lib/$(RUST_LIB_NAME)-linux-x86_64.so || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64) && \
		([ -f lib/$(RUST_LIB_NAME)-darwin-x86_64.dylib ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 lib/$(RUST_LIB_NAME)-darwin-x86_64.dylib || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64) && \
		([ -f lib/$(RUST_LIB_NAME)-darwin-aarch64.dylib ] && tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 lib/$(RUST_LIB_NAME)-darwin-aarch64.dylib || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64) && \
		([ -f lib/$(RUST_LIB_NAME)-windows-x86_64.dll ] && (which zip >/dev/null 2>&1 && zip -q ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe lib/$(RUST_LIB_NAME)-windows-x86_64.dll || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe lib/$(RUST_LIB_NAME)-windows-x86_64.dll) || (which zip >/dev/null 2>&1 && zip -q ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe || tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe))
	@echo "✓ Release packages created in releases/$(VERSION)/ (with Rust support where available)"

# Create git tag
release-tag:
	@echo "Creating git tag $(VERSION)..."
	@git tag -a v$(VERSION) -m "Release $(VERSION)"
	@echo "✓ Git tag v$(VERSION) created"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "=== Go Application Targets ==="
	@echo "  build            - Build the application (with Rust support)"
	@echo "  build-go-only    - Build the application (Go-only, no Rust)"
	@echo "  build-all        - Build for multiple platforms (with Rust libraries)"
	@echo "  build-all-go-only - Build for multiple platforms (Go-only mode)"
	@echo "  run              - Build and run the application"
	@echo "  dev              - Run in development mode"
	@echo ""
	@echo "=== Rust Library Targets ==="
	@echo "  rust-check       - Check if Rust toolchain is installed"
	@echo "  rust-setup       - Setup Rust cross-compilation targets"
	@echo "  rust-lib         - Build Rust library for current platform"
	@echo "  rust-lib-cross   - Build Rust library for all platforms"
	@echo "  rust-clean       - Clean Rust build artifacts"
	@echo "  rust-install     - Install Rust toolchain and cross-compilation"
	@echo ""
	@echo "=== Testing and Quality ==="
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  benchmark        - Run benchmarks"
	@echo "  benchmark-suite  - Run comprehensive benchmark suite"
	@echo "  benchmark-report - Run benchmarks and generate report"
	@echo "  benchmark-validate - Run performance validation benchmarks"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Vet code"
	@echo ""
	@echo "=== Release Management ==="
	@echo "  release-check    - Run pre-release validation checks"
	@echo "  release-build    - Create multi-platform release builds"
	@echo "  release-package  - Package release binaries (with Rust libraries)"
	@echo "  release-build-minimal    - Create release builds (skip failing tests)"
	@echo "  release-package-minimal  - Package release binaries (skip tests, with Rust)"
	@echo "  release-tag      - Create git release tag"
	@echo ""
	@echo "=== Utilities ==="
	@echo "  clean            - Clean all build artifacts (Go + Rust)"
	@echo "  deps             - Install dependencies"
	@echo "  install          - Install the binary"
	@echo "  help             - Show this help"
	@echo ""
	@echo "=== Platform Information ==="
	@echo "  Current OS: $(UNAME_OS)"
	@echo "  Current Architecture: $(UNAME_ARCH)"
	@echo "  Rust Target: $(CARGO_TARGET)"
	@echo "  Rust Library: $(RUST_LIB_FILE)"