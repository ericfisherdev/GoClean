# GoClean Rust Configuration Guide

This guide provides comprehensive information about configuring GoClean for Rust projects, including examples, best practices, and integration patterns.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration Examples](#configuration-examples)
3. [Rust-Specific Configuration](#rust-specific-configuration)
4. [Clippy Integration](#clippy-integration)
5. [CLI Usage Examples](#cli-usage-examples)
6. [Mixed Projects (Go + Rust)](#mixed-projects-go--rust)
7. [Performance Optimization](#performance-optimization)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Basic Rust Project Setup

For a simple Rust project, use the minimal configuration:

```bash
# Copy the minimal Rust configuration
cp configs/rust-minimal.yaml goclean.yaml

# Run GoClean on your Rust project
goclean scan
```

### 2. Generate Custom Configuration

```bash
# Generate a basic configuration file
goclean config init goclean-rust.yaml

# Generate in different location  
goclean config init configs/rust-config.yaml
```

### 3. Quick Scan Commands

```bash
# Basic Rust project scan
goclean scan ./src --format json

# Include Rust files specifically
goclean scan --types .rs --console-violations

# Mixed project with both Go and Rust
goclean scan --types .go,.rs --format html
```

## Configuration Examples

### Minimal Configuration (rust-minimal.yaml)

Perfect for getting started with Rust projects:

```yaml
scan:
  paths: ["./src"]
  file_types: [".rs"]
  exclude: ["target/", "Cargo.lock"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  allow_unwrap: false

output:
  console: { enabled: true }
  html: { enabled: true, path: "./reports/rust-report.html" }
```

**Use case**: New Rust projects, quick validation, CI/CD integration

### Strict Configuration (rust-strict.yaml)

High standards for production Rust code:

```yaml
thresholds:
  function_lines: 20
  cyclomatic_complexity: 6
  parameters: 3
  nesting_depth: 2

rust:
  allow_unwrap: false
  allow_expect: false
  max_lifetime_params: 2
  require_exhaustive_match: true

clippy:
  categories: ["correctness", "suspicious", "style", "complexity", "perf", "pedantic"]
```

**Use case**: Production systems, security-critical code, team standards enforcement

### Performance-Focused Configuration (rust-performance-focused.yaml)

Optimized for large codebases:

```yaml
scan:
  concurrent_files: 20
  max_file_size: "1MB"

rust:
  enable_pattern_match_check: false  # For speed
  require_unsafe_comments: false

output:
  html: { enabled: false }  # Faster execution
  console: { format: "summary" }
```

**Use case**: Large monorepos, CI/CD pipelines, automated analysis

### Mixed Project Configuration (rust-mixed-project.yaml)

For projects containing both Go and Rust:

```yaml
scan:
  file_types: [".go", ".rs"]
  exclude: ["vendor/", "target/", "*.test.go"]

rust:
  # Rust-specific settings
  enable_ownership_analysis: true
  enforce_snake_case: true

# Standard Go/Rust thresholds
thresholds:
  function_lines: 25
  cyclomatic_complexity: 8
```

**Use case**: Microservices with mixed languages, gradual migration projects

## Rust-Specific Configuration

### Ownership and Borrowing Analysis

```yaml
rust:
  enable_ownership_analysis: true
  max_lifetime_params: 3
  detect_unnecessary_clones: true
```

**Detects**:
- `RUST_UNNECESSARY_CLONE` - Unnecessary `.clone()` calls
- `RUST_INEFFICIENT_BORROWING` - Suboptimal borrowing patterns
- `RUST_COMPLEX_LIFETIME` - Overly complex lifetime parameters

### Error Handling Analysis

```yaml
rust:
  enable_error_handling_check: true
  allow_unwrap: false
  allow_expect: false
  enforce_result_propagation: true
```

**Detects**:
- `RUST_OVERUSE_UNWRAP` - Usage of `.unwrap()`
- `RUST_MISSING_ERROR_PROPAGATION` - Missing `?` operator usage
- `RUST_INCONSISTENT_ERROR_TYPE` - Inconsistent error types
- `RUST_PANIC_PRONE_CODE` - Code that may panic

### Pattern Matching Analysis

```yaml
rust:
  enable_pattern_match_check: true
  require_exhaustive_match: true
  max_nested_match_depth: 3
```

**Detects**:
- `RUST_NON_EXHAUSTIVE_MATCH` - Non-exhaustive pattern matches
- `RUST_NESTED_PATTERN_MATCHING` - Overly nested match expressions
- `RUST_INEFFICIENT_DESTRUCTURING` - Inefficient destructuring patterns

### Safety Analysis

```yaml
rust:
  allow_unsafe: true
  require_unsafe_comments: true
  detect_transmute_usage: true
```

**Detects**:
- `RUST_UNNECESSARY_UNSAFE` - Unnecessary unsafe blocks
- `RUST_UNSAFE_WITHOUT_COMMENT` - Unsafe code without documentation
- `RUST_TRANSMUTE_ABUSE` - Dangerous transmute usage

### Performance Analysis

```yaml
rust:
  detect_inefficient_string: true
  detect_boxed_primitives: true
  detect_blocking_in_async: true
```

**Detects**:
- `RUST_INEFFICIENT_STRING_CONCAT` - Inefficient string concatenation
- `RUST_UNNECESSARY_ALLOCATION` - Unnecessary heap allocations
- `RUST_BLOCKING_IN_ASYNC` - Blocking calls in async functions

### Naming Conventions

```yaml
rust:
  enforce_snake_case: true      # Functions, variables, modules
  enforce_pascal_case: true     # Types (structs, enums, traits)
  enforce_screaming_snake: true # Constants
```

**Detects**:
- `RUST_INVALID_FUNCTION_NAMING` - Incorrect function naming
- `RUST_INVALID_STRUCT_NAMING` - Incorrect struct naming
- `RUST_INVALID_CONSTANT_NAMING` - Incorrect constant naming

## Clippy Integration

GoClean integrates with rust-clippy to provide comprehensive analysis:

### Basic Clippy Configuration

```yaml
clippy:
  enabled: true
  categories:
    - correctness    # Critical correctness issues
    - suspicious     # Suspicious code patterns
    - style         # Style and idiom violations
    - complexity    # Code complexity issues
    - perf          # Performance improvements
```

### Advanced Clippy Configuration

```yaml
clippy:
  enabled: true
  categories:
    - correctness
    - suspicious
    - style
    - complexity
    - perf
    - pedantic      # Extra strict lints
    - nursery       # Experimental lints
  severity_mapping:
    error: critical
    warn: high
    info: medium
    note: low
  additional_lints:
    - clippy::all
    - clippy::pedantic
    - clippy::cargo
```

### Clippy Severity Mapping

| Clippy Level | GoClean Severity | Description |
|--------------|------------------|-------------|
| `error`      | `critical`       | Must fix before production |
| `warn`       | `high`          | Should fix soon |
| `info`       | `medium`        | Improvement opportunity |
| `note`       | `low`           | Minor suggestion |

## CLI Usage Examples

### Basic Commands

```bash
# Scan current directory for Rust files
goclean scan --types .rs

# Use specific configuration with Rust file types
goclean scan --config configs/rust-strict.yaml --types .rs

# Scan specific directories (positional arguments followed by flags)
goclean scan ./src ./benches --types .rs

# Generate HTML report
goclean scan --format html --output reports/rust-report.html

# Generate Markdown report
goclean scan --format markdown --output reports/violations.md

# Generate JSON report
goclean scan --format json --output data/violations.json
```

### Advanced Commands

```bash
# Mixed language scanning
goclean scan --types .go,.rs --config configs/rust-mixed-project.yaml

# Performance-focused scan with aggressive mode
goclean scan --config configs/rust-performance-focused.yaml --aggressive --types .rs

# Strict analysis with JSON output
goclean scan --config configs/rust-strict.yaml --format json --output results.json

# CI/CD friendly console output
goclean scan --config configs/rust-minimal.yaml --console-violations --types .rs
```

### Filtering and Exclusions

```bash
# Exclude test files and benchmarks
goclean scan --exclude "tests/" --exclude "benches/" --exclude "target/"

# Use aggressive scanning for stricter analysis
goclean scan --aggressive --types .rs

# Use minimal configuration for faster scanning
goclean scan --config configs/rust-minimal.yaml
```

### Report Generation

```bash
# Console output for AI analysis
goclean scan --console-violations --types .rs

# Export data for analysis using JSON format
goclean scan --format json --output data.json

# Enable aggressive scanning mode with Rust files
goclean scan --aggressive --types .rs
```

## Mixed Projects (Go + Rust)

For projects containing both Go and Rust code:

### Configuration Strategy

```yaml
scan:
  paths:
    - ./src          # Rust source
    - ./internal     # Go source
    - ./cmd          # Go commands
    - ./rust-modules # Rust modules
  file_types:
    - .go
    - .rs
  exclude:
    - vendor/        # Go dependencies
    - target/        # Rust build artifacts
    - "*.test.go"    # Go test files

# Balanced thresholds for both languages
thresholds:
  function_lines: 25
  cyclomatic_complexity: 8
  parameters: 4

# Rust-specific configuration
rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  enforce_snake_case: true

# Enable clippy for Rust analysis
clippy:
  enabled: true
  categories: ["correctness", "suspicious", "style", "complexity", "perf"]
```

### Best Practices for Mixed Projects

1. **Separate Configurations**: Consider separate configs for different components
2. **Consistent Standards**: Align complexity thresholds between languages
3. **Documentation**: Ensure both languages follow documentation standards
4. **CI Integration**: Run language-specific tests in parallel

### Example Project Structure

```
project/
├── src/           # Rust source code
│   ├── lib.rs
│   └── modules/
├── internal/      # Go internal packages
│   └── api/
├── cmd/           # Go commands
│   └── server/
├── rust-ffi/      # Rust FFI bindings
└── goclean.yaml   # Mixed project configuration
```

## Performance Optimization

### Large Codebase Strategies

```yaml
scan:
  concurrent_files: 50      # Increase parallelism
  max_file_size: "2MB"      # Skip very large files
  exclude:
    - "target/"             # Skip build artifacts
    - "examples/"           # Skip example code
    - "benchmarks/"         # Skip benchmark code

rust:
  enable_pattern_match_check: false  # Disable expensive checks
  require_unsafe_comments: false     # Skip comment analysis

output:
  html: { enabled: false }           # Disable slow HTML generation
  console: { format: "summary" }     # Use fast summary format
```

### Memory Optimization

```yaml
logging:
  level: "warn"              # Reduce logging overhead

output:
  export:
    json: { pretty_print: false }   # Reduce memory usage
```

### CI/CD Optimization

```yaml
# Fast CI configuration
scan:
  concurrent_files: 20
  exclude: ["target/", "examples/", "benches/"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
  enable_pattern_match_check: false

clippy:
  categories: ["correctness", "suspicious"]  # Only critical lints

output:
  console: { format: "json", color: false }
  html: { enabled: false }

logging:
  level: "error"
  format: "simple"
```

## Best Practices

### Development Workflow

1. **Start Minimal**: Begin with `rust-minimal.yaml` configuration
2. **Gradual Strictness**: Progressively adopt stricter rules
3. **Team Agreement**: Establish team consensus on thresholds
4. **Regular Reviews**: Review and adjust configurations periodically

### Configuration Management

```bash
# Version control your configuration
git add goclean.yaml

# Use environment-specific configurations
goclean scan --config configs/rust-dev.yaml      # Development
goclean scan --config configs/rust-prod.yaml     # Production
goclean scan --config configs/rust-ci.yaml       # CI/CD
```

### Integration Patterns

```yaml
# Pre-commit hook configuration
scan:
  paths: ["."]
  exclude: ["target/"]

rust:
  allow_unwrap: false
  allow_expect: false

output:
  console: { format: "table", show_summary: true }
  
# Fail on critical violations
exit_on: ["critical"]
```

### Documentation Standards

```yaml
rust:
  # Require documentation for public APIs
  check_missing_docs: true
  min_doc_coverage: 0.8  # 80% documentation coverage

documentation:
  enabled: true
  check_missing_docs: true
  check_outdated_comments: true
```

## Troubleshooting

### Common Issues

#### 1. Clippy Not Found

```bash
# Install clippy
rustup component add clippy

# Verify installation
cargo clippy --version
```

#### 2. Performance Issues

```bash
# Use performance-focused configuration
goclean scan --config configs/rust-performance-focused.yaml

# Use minimal configuration if running out of memory
goclean scan --config configs/rust-minimal.yaml
```

#### 3. Too Many Violations

```bash
# Start with lenient thresholds
goclean scan --config configs/rust-minimal.yaml

# Use strict configuration for critical violations
goclean scan --config configs/rust-strict.yaml
```

#### 4. Configuration Validation

```bash
# Validate configuration syntax
goclean config validate

# Check configuration with verbose output
goclean scan --config goclean.yaml --verbose
```

### Debug Commands

```bash
# Debug Rust parsing with verbose output
goclean scan --verbose --rust-opt

# Check file discovery with verbose output
goclean scan --verbose

# Analyze with verbose output
goclean scan --verbose --format json --output analysis.json

# Test specific files
goclean scan ./src/main.rs --verbose
```

### Environment Variables

```bash
# Enable debug logging
export GOCLEAN_LOG_LEVEL=debug

# Override Rust configuration
export GOCLEAN_RUST_ALLOW_UNWRAP=false
export GOCLEAN_RUST_ENABLE_OWNERSHIP_ANALYSIS=true

# Performance tuning
export GOCLEAN_SCAN_CONCURRENT_FILES=30
export GOCLEAN_SCAN_MAX_FILE_SIZE=2MB
```

## Migration Examples

### From Clippy-Only to GoClean

If you're currently using only clippy:

```yaml
# Complement clippy with GoClean analysis
clippy:
  enabled: true
  categories: ["correctness", "suspicious", "perf"]

rust:
  # GoClean-specific analysis
  enable_ownership_analysis: true
  detect_unnecessary_clones: true
  max_lifetime_params: 3
  
# Additional clean code metrics
thresholds:
  function_lines: 30
  cyclomatic_complexity: 10
```

### From Go-Only Project to Mixed

```yaml
# Before (Go only)
scan:
  file_types: [".go"]
  
# After (Go + Rust)
scan:
  file_types: [".go", ".rs"]
  exclude: ["vendor/", "target/"]

rust:
  enable_ownership_analysis: true
  enable_error_handling_check: true
```

This configuration guide provides comprehensive coverage of GoClean's Rust support. For additional help, consult the main documentation or create an issue on the GitHub repository.