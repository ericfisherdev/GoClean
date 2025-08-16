# GoClean Makefile

.PHONY: build test clean install lint fmt vet run dev benchmark benchmark-suite benchmark-report benchmark-validate release-check release-build release-package release-tag

# Build configuration
BINARY_NAME=goclean
BUILD_DIR=bin
MAIN_PATH=./cmd/goclean

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

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@echo "Working directory: $(PWD)"
	@echo "Main path: $(MAIN_PATH)"
	@ls -la $(MAIN_PATH) 2>/dev/null || echo "Directory $(MAIN_PATH) not found, listing current dir:" && ls -la .
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	GOCLEAN_TEST_MODE=1 $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

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
	golangci-lint run

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

# Build for multiple platforms
build-all:
	@echo "Building $(VERSION) for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

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

# Create release build
release-build: release-check
	@echo "Creating release build $(VERSION)..."
	@$(MAKE) clean
	@$(MAKE) build-all
	@echo "✓ Release build completed"

# Package release
release-package: release-build
	@echo "Packaging release $(VERSION)..."
	@mkdir -p releases/$(VERSION)
	@cd $(BUILD_DIR) && \
		tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
		tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
		zip -q ../releases/$(VERSION)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "✓ Release packages created in releases/$(VERSION)/"

# Create git tag
release-tag:
	@echo "Creating git tag $(VERSION)..."
	@git tag -a v$(VERSION) -m "Release $(VERSION)"
	@echo "✓ Git tag v$(VERSION) created"

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  benchmark        - Run benchmarks"
	@echo "  benchmark-suite  - Run comprehensive benchmark suite"
	@echo "  benchmark-report - Run benchmarks and generate report"
	@echo "  benchmark-validate - Run performance validation benchmarks"
	@echo "  release-check    - Run pre-release validation checks"
	@echo "  release-build    - Create multi-platform release builds"
	@echo "  release-package  - Package release binaries"
	@echo "  release-tag      - Create git release tag"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  install          - Install the binary"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Vet code"
	@echo "  run              - Build and run the application"
	@echo "  dev              - Run in development mode"
	@echo "  build-all        - Build for multiple platforms"
	@echo "  help             - Show this help"