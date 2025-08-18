# GoClean CI/CD Integration for Rust Support

This document provides comprehensive guidance for setting up CI/CD integration to validate GoClean's Rust support functionality.

## Overview

The CI/CD pipeline should validate both Go and Rust functionality, ensuring that:
- Rust parsing works correctly
- Rust violation detection functions properly
- Performance benchmarks meet targets
- Integration with real Rust projects succeeds
- Cross-platform compatibility is maintained

## Recommended CI/CD Workflow Structure

### 1. Basic CI Workflow (`ci.yml`)

```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

permissions:
  contents: read

env:
  GO_VERSION: '1.21'
  RUST_VERSION: 'stable'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        test-suite:
          - go-tests
          - rust-tests
          - integration-tests
          - benchmark-tests
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true
          cache-dependency-path: go.sum
          
      - name: Set up Rust toolchain
        if: matrix.test-suite == 'rust-tests'
        uses: actions-rs/toolchain@v1
        with:
          toolchain: stable
          profile: minimal
          override: true
          components: rustfmt, clippy
```

### 2. Test Suite Components

#### Go Tests
```yaml
- name: Build GoClean binary
  run: |
    mkdir -p ./bin
    go build -o ./bin/goclean ./cmd/goclean
    
- name: Run Go tests
  if: matrix.test-suite == 'go-tests'
  run: |
    go test -v -coverprofile=coverage-go.out ./internal/scanner -run="Test.*Go"
    go test -v ./internal/violations -run="Test.*Go"
    go test -v ./internal/reporters ./internal/config ./internal/models
    go test -v ./cmd/goclean
```

#### Rust Tests
```yaml
- name: Run Rust tests
  if: matrix.test-suite == 'rust-tests'
  run: |
    go test -v -coverprofile=coverage-rust.out ./internal/scanner -run="Test.*Rust"
    go test -v ./internal/violations -run="Test.*Rust"
```

#### Integration Tests
```yaml
- name: Run integration tests
  if: matrix.test-suite == 'integration-tests'
  run: |
    go test -v ./internal/scanner -run="TestIntegration"
    go test -v ./internal/violations -run="TestIntegration"
```

#### Benchmark Tests
```yaml
- name: Install benchstat
  if: matrix.test-suite == 'benchmark-tests'
  run: |
    GO111MODULE=on go install golang.org/x/perf/cmd/benchstat@latest
    echo "$GOPATH/bin" >> $GITHUB_PATH

- name: Run benchmark tests
  if: matrix.test-suite == 'benchmark-tests'
  run: |
    go test -bench=BenchmarkRustVsGoParsingPerformance/Small -benchmem -timeout=5m ./internal/scanner
    go test -bench=BenchmarkParsingMemoryComparison -benchmem -timeout=3m ./internal/scanner
```

## 3. Rust Sample Project Creation

### Creating Test Data

The CI should create comprehensive Rust sample projects for testing:

```bash
# Create sample Rust project structure
mkdir -p testdata/rust-samples/sample-project/src

# Create Cargo.toml
cat > testdata/rust-samples/sample-project/Cargo.toml << 'EOF'
[package]
name = "sample-project"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", features = ["full"] }
EOF

# Create main.rs with sample code
cat > testdata/rust-samples/sample-project/src/main.rs << 'EOF'
use serde::{Deserialize, Serialize};
use tokio;

#[derive(Serialize, Deserialize)]
struct Config {
    name: String,
    version: String,
}

#[tokio::main]
async fn main() {
    let config = Config {
        name: "sample-project".to_string(),
        version: "0.1.0".to_string(),
    };
    
    println!("Hello from {}!", config.name);
}
EOF
```

### Sample Rust Code with Violations

Create test files that contain various types of violations:

```rust
// main.rs - Complex async code with various constructs
use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct User {
    pub id: u64,
    pub name: String,
    pub email: Option<String>,
}

impl User {
    pub fn new(id: u64, name: String) -> Self {
        User { id, name, email: None }
    }

    pub async fn fetch_profile(&self) -> Result<String, Box<dyn std::error::Error>> {
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
        Ok(format!("Profile for user: {}", self.name))
    }
}

// Intentional violations for testing
pub fn function_with_many_parameters(
    param1: i32, param2: String, param3: bool, param4: f64,
    param5: Vec<u8>, param6: HashMap<String, i32>,
    param7: Option<String>, param8: Result<i32, String>,
) -> String {
    format!("Processing {} parameters", 8)
}

pub fn calculate_score(base: i32) -> i32 {
    base * 42 + 1337  // Magic numbers
}
```

## 4. Validation Steps

### Rust Parsing Validation
```bash
# Ensure binary exists and build if needed
if [ ! -f ./bin/goclean ]; then
  mkdir -p ./bin
  go build -o ./bin/goclean ./cmd/goclean
fi

# Test GoClean's Rust parsing
./bin/goclean scan testdata/rust-samples/sample-project --console --format table

# Generate JSON output for analysis
./bin/goclean scan testdata/rust-samples/sample-project --format json > rust-scan-results.json

# Verify Rust-specific violations were detected
if grep -q "RUST_" rust-scan-results.json; then
  echo "✓ Rust-specific violations detected successfully"
else
  echo "⚠ No Rust-specific violations detected"
fi
```

### Expected Rust Violations
The CI should check for these violation types:
- `RUST_INVALID_FUNCTION_NAMING`
- `RUST_INVALID_STRUCT_NAMING`
- `RUST_FUNCTION_TOO_LONG`
- `RUST_TOO_MANY_PARAMETERS`
- `RUST_MAGIC_NUMBER`
- `RUST_MISSING_DOCUMENTATION`
- `RUST_COMMENTED_CODE`
- `RUST_TODO_COMMENT`

### Performance Benchmarks
```bash
# Run performance benchmarks
go test -bench=BenchmarkRustVsGoParsingPerformance -benchmem -count=3 ./internal/scanner

# Performance targets to validate:
# - Go parsing: >1000 files/sec
# - Rust parsing: >50 files/sec  
# - Memory usage: <10MB per 100 files
```

## 5. Cross-Platform Testing

### Matrix Strategy
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.21']
```

### Platform-Specific Considerations
- **Linux**: Primary development and testing platform
- **macOS**: Test Rust toolchain compatibility
- **Windows**: Verify binary execution and path handling

## 6. Performance Regression Detection

### Benchmark Comparison
```bash
# Store baseline benchmarks
go test -bench=. -benchmem ./internal/scanner > baseline-benchmarks.txt

# Compare current vs baseline (in PR workflow)
go test -bench=. -benchmem ./internal/scanner > current-benchmarks.txt
benchstat baseline-benchmarks.txt current-benchmarks.txt
```

### Performance Thresholds
- **Startup time**: <500ms
- **Scanning speed**: >1000 files/sec (Go), >50 files/sec (Rust)
- **Memory usage**: <100MB per 10k files
- **Report generation**: <2s for 5k violations

## 7. Documentation Validation

### Required Documentation
```bash
# Check for Rust support documentation
grep -q -i "rust" README.md || echo "Update README.md with Rust support"
grep -r -q -i "rust" docs/ || echo "Add Rust-specific documentation"
```

### Documentation Requirements
- README.md mentions Rust support
- Installation instructions include Rust dependencies
- Configuration examples for Rust projects
- Rust-specific violation documentation

## 8. Cache Strategy

### Go Modules Cache
```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

### Rust Dependencies Cache
```yaml
- name: Cache Rust dependencies
  uses: actions/cache@v4
  with:
    path: |
      ~/.cargo/registry
      ~/.cargo/git
      target
    key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}
```

## 9. Integration with Existing Workflows

### Release Workflow Updates
```yaml
# Add Rust validation to release workflow
- name: Validate Rust support
  run: |
    make test
    go test -bench=BenchmarkRustVsGoParsingPerformance -timeout=10m ./internal/scanner
```

### Pre-commit Hooks
```bash
# Add to .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: rust-validation
        name: Validate Rust support
        entry: make test-rust
        language: system
        pass_filenames: false
```

## 10. Monitoring and Alerts

### Success Criteria
- All tests pass
- Benchmarks meet performance targets
- No regressions in existing functionality
- Rust violations detected correctly

### Failure Scenarios
- Rust parsing fails
- Performance regression >20%
- Memory usage exceeds thresholds
- Cross-platform compatibility issues

## 11. Example Make Targets

Add these targets to the Makefile:

```makefile
# Test Rust support specifically
test-rust:
	@echo "Testing Rust support..."
	GOCLEAN_TEST_MODE=1 $(GOTEST) -v ./internal/scanner -run="Test.*Rust"
	GOCLEAN_TEST_MODE=1 $(GOTEST) -v ./internal/violations -run="Test.*Rust"

# Run Rust benchmarks
benchmark-rust:
	@echo "Running Rust benchmarks..."
	$(GOTEST) -bench=BenchmarkRust -benchmem ./internal/scanner

# Validate Rust integration
validate-rust: test-rust benchmark-rust
	@echo "Creating Rust test project..."
	@mkdir -p testdata/integration/rust
	@echo 'fn main() { println!("Hello, world!"); }' > testdata/integration/rust/hello.rs
	@echo 'pub struct TestStruct { pub field: i32 }' > testdata/integration/rust/struct.rs
	./bin/goclean scan testdata/integration/rust --console
	@echo "✓ Rust integration validation completed"
```

## 12. Troubleshooting

### Common Issues
1. **Rust toolchain not found**: Ensure Rust is installed in CI environment
2. **Performance regression**: Check for memory leaks or inefficient parsing
3. **Cross-platform failures**: Verify path handling and file system differences
4. **Benchmark timeouts**: Adjust timeout values for CI environment

### Debug Commands
```bash
# Debug Rust parsing
./bin/goclean scan path/to/rust/project --verbose --console

# Check memory usage
go test -bench=BenchmarkRustMemoryUsage -memprofile=mem.prof ./internal/scanner
go tool pprof mem.prof

# Analyze performance
go test -bench=BenchmarkRustVsGoParsingPerformance -cpuprofile=cpu.prof ./internal/scanner
go tool pprof cpu.prof
```

## Implementation Notes

This documentation provides the framework for comprehensive CI/CD validation of Rust support. The actual implementation should be adapted based on:

- Specific CI/CD platform (GitHub Actions, GitLab CI, etc.)
- Repository structure and existing workflows
- Performance requirements and constraints
- Team development practices

The key is ensuring that both Go and Rust functionality are thoroughly tested across multiple dimensions: functionality, performance, compatibility, and integration.