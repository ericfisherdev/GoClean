# GoClean Rust CLI Usage Examples

This document provides comprehensive command-line examples for using GoClean with Rust projects.

## Table of Contents

1. [Basic Commands](#basic-commands)
2. [Configuration Management](#configuration-management)
3. [Scanning Strategies](#scanning-strategies)
4. [Output Formats](#output-formats)
5. [Mixed Language Projects](#mixed-language-projects)
6. [CI/CD Integration](#cicd-integration)
7. [Advanced Usage](#advanced-usage)
8. [Troubleshooting Commands](#troubleshooting-commands)

## Basic Commands

### Simple Rust Project Scanning

```bash
# Scan current directory for Rust files
goclean scan --file-types .rs

# Scan specific directory
goclean scan --path ./src --file-types .rs

# Scan with console output
goclean scan --path ./src --file-types .rs --console

# Scan with HTML report
goclean scan --path ./src --file-types .rs --html
```

### Quick Start Examples

```bash
# First-time setup for Rust project
goclean config init --rust
goclean scan

# Quick validation scan
goclean scan --file-types .rs --console --format table

# Generate comprehensive report
goclean scan --file-types .rs --html --markdown --console
```

## Configuration Management

### Using Predefined Configurations

```bash
# Use minimal Rust configuration
goclean scan --config configs/rust-minimal.yaml

# Use strict configuration for production code
goclean scan --config configs/rust-strict.yaml

# Use performance-focused configuration for large codebases
goclean scan --config configs/rust-performance-focused.yaml

# Use mixed project configuration
goclean scan --config configs/rust-mixed-project.yaml
```

### Configuration Generation

```bash
# Generate basic Rust configuration
goclean config init --template rust --output my-rust-config.yaml

# Generate strict configuration
goclean config init --template rust-strict --output strict.yaml

# Generate minimal configuration
goclean config init --template rust-minimal --output minimal.yaml

# Validate configuration
goclean config validate --config my-rust-config.yaml
```

### Environment Variable Overrides

```bash
# Override paths via environment
export GOCLEAN_SCAN_PATHS="./src,./benches"
goclean scan --file-types .rs

# Override thresholds
export GOCLEAN_THRESHOLDS_FUNCTION_LINES=20
export GOCLEAN_RUST_ALLOW_UNWRAP=false
goclean scan --config configs/rust-minimal.yaml

# Override output settings
export GOCLEAN_OUTPUT_HTML_THEME=dark
export GOCLEAN_OUTPUT_MARKDOWN_ENABLED=true
goclean scan --file-types .rs
```

## Scanning Strategies

### Selective File Scanning

```bash
# Scan only source files (exclude tests and examples)
goclean scan --path ./src --exclude "tests/" --exclude "examples/" --file-types .rs

# Scan specific Rust files
goclean scan --path ./src/main.rs --path ./src/lib.rs

# Scan with size limits
goclean scan --file-types .rs --max-file-size 1MB

# Concurrent scanning for performance
goclean scan --file-types .rs --concurrent 20
```

### Project Structure Scanning

```bash
# Typical Rust project structure
goclean scan \
  --path ./src \
  --path ./benches \
  --path ./examples \
  --exclude "target/" \
  --exclude "Cargo.lock" \
  --file-types .rs

# Library project focus
goclean scan \
  --path ./src \
  --exclude "target/" \
  --exclude "tests/" \
  --file-types .rs

# Workspace scanning
goclean scan \
  --path ./crates \
  --exclude "target/" \
  --exclude "*/target/" \
  --file-types .rs
```

### Violation-Specific Scanning

```bash
# Focus on ownership issues
goclean scan --file-types .rs --config configs/ownership-focus.yaml

# Error handling analysis only
goclean scan --file-types .rs --enable-rules error_handling

# Performance-focused analysis
goclean scan --file-types .rs --enable-rules performance,naming
```

## Output Formats

### Console Output Variations

```bash
# Table format (default)
goclean scan --file-types .rs --console --format table

# JSON output for automation
goclean scan --file-types .rs --console --format json

# CSV output for analysis
goclean scan --file-types .rs --console --format csv

# Summary format for quick overview
goclean scan --file-types .rs --console --format summary

# No color output for logs
goclean scan --file-types .rs --console --no-color
```

### HTML Report Generation

```bash
# Basic HTML report
goclean scan --file-types .rs --html

# Custom HTML output path
goclean scan --file-types .rs --html --html-output ./reports/rust-analysis.html

# Dark theme HTML report
goclean scan --file-types .rs --html --theme dark

# Auto-refreshing report for development
goclean scan --file-types .rs --html --auto-refresh --refresh-interval 5

# HTML with code snippets
goclean scan --file-types .rs --html --show-snippets --snippet-lines 10
```

### Markdown Output

```bash
# Basic markdown report
goclean scan --file-types .rs --markdown

# AI-friendly markdown format
goclean scan --file-types .rs --markdown --ai-friendly

# Markdown with examples included
goclean scan --file-types .rs --markdown --include-examples

# Custom markdown output path
goclean scan --file-types .rs --markdown --markdown-output ./docs/violations.md
```

### Export Formats

```bash
# Export to JSON
goclean scan --file-types .rs --export-json --json-output results.json

# Export to CSV for spreadsheet analysis
goclean scan --file-types .rs --export-csv --csv-output violations.csv

# Pretty-printed JSON
goclean scan --file-types .rs --export-json --pretty-print

# Multiple export formats
goclean scan --file-types .rs --export-json --export-csv --html --markdown
```

## Mixed Language Projects

### Go + Rust Projects

```bash
# Scan both Go and Rust files
goclean scan --file-types .go,.rs

# Mixed project with custom exclusions
goclean scan \
  --file-types .go,.rs \
  --exclude "vendor/" \
  --exclude "target/" \
  --exclude "*.test.go"

# Use mixed project configuration
goclean scan --config configs/rust-mixed-project.yaml

# Separate reports for each language
goclean scan --file-types .go --html --html-output go-report.html
goclean scan --file-types .rs --html --html-output rust-report.html
```

### Microservices Architecture

```bash
# Scan Go services
goclean scan \
  --path ./services/api \
  --path ./services/auth \
  --file-types .go \
  --html --html-output go-services-report.html

# Scan Rust services
goclean scan \
  --path ./services/processor \
  --path ./services/analytics \
  --file-types .rs \
  --html --html-output rust-services-report.html

# Combined microservices scan
goclean scan \
  --path ./services \
  --file-types .go,.rs \
  --exclude "vendor/" \
  --exclude "target/" \
  --html --markdown
```

## CI/CD Integration

### GitHub Actions Examples

```bash
# Basic CI scan
goclean scan \
  --config configs/rust-minimal.yaml \
  --format json \
  --no-color \
  --quiet > goclean-results.json

# Quality gate with error codes
goclean scan \
  --config configs/rust-strict.yaml \
  --format json \
  --fail-on critical,high

# Performance monitoring
goclean scan \
  --config configs/rust-performance-focused.yaml \
  --benchmark \
  --benchmark-output benchmark-results.json

# Artifact generation
goclean scan \
  --config configs/rust-minimal.yaml \
  --html --html-output ./artifacts/code-quality-report.html \
  --markdown --markdown-output ./artifacts/violations.md \
  --export-json --json-output ./artifacts/scan-results.json
```

### GitLab CI Examples

```bash
# GitLab CI job
goclean scan \
  --config .goclean-ci.yaml \
  --format json \
  --export-json \
  --json-output gl-code-quality-report.json

# Parallel analysis
goclean scan \
  --file-types .rs \
  --concurrent 10 \
  --format summary \
  --quiet

# Cache-friendly scanning
goclean scan \
  --config configs/rust-performance-focused.yaml \
  --cache-dir ./.goclean-cache \
  --format json
```

### Jenkins Pipeline Examples

```bash
# Jenkins stage
goclean scan \
  --config jenkins-rust.yaml \
  --format json \
  --export-json \
  --json-output target/goclean-report.json \
  --html \
  --html-output target/goclean-report.html

# Quality gate
goclean scan \
  --config configs/rust-strict.yaml \
  --format json \
  --fail-on critical \
  --quiet || exit 1

# Trend analysis
goclean scan \
  --config configs/rust-minimal.yaml \
  --format json \
  --export-json \
  --json-output "reports/scan-$(date +%Y%m%d).json"
```

## Advanced Usage

### Performance Optimization

```bash
# High-performance scanning
goclean scan \
  --file-types .rs \
  --concurrent 50 \
  --max-file-size 2MB \
  --exclude "target/" \
  --exclude "benchmarks/" \
  --format summary

# Memory-conscious scanning
goclean scan \
  --file-types .rs \
  --concurrent 5 \
  --max-memory 1GB \
  --streaming \
  --format json

# Profile performance
goclean scan \
  --file-types .rs \
  --profile \
  --profile-output performance.prof \
  --benchmark
```

### Custom Rule Configuration

```bash
# Enable specific Rust rules
goclean scan \
  --file-types .rs \
  --enable-rules ownership,error_handling,naming \
  --console

# Disable specific rules
goclean scan \
  --file-types .rs \
  --disable-rules magic_numbers,todo_comments \
  --console

# Custom thresholds via CLI
goclean scan \
  --file-types .rs \
  --function-lines 20 \
  --complexity 6 \
  --parameters 3 \
  --console
```

### Filtering and Selection

```bash
# Severity filtering
goclean scan --file-types .rs --severity critical,high --console

# File pattern filtering
goclean scan \
  --file-types .rs \
  --include "src/**/*.rs" \
  --exclude "src/generated/**" \
  --console

# Violation type filtering
goclean scan \
  --file-types .rs \
  --include-violations "RUST_FUNCTION_TOO_LONG,RUST_TOO_MANY_PARAMETERS" \
  --console

# Time-based filtering (files modified in last 7 days)
goclean scan \
  --file-types .rs \
  --modified-since "7 days ago" \
  --console
```

### Reporting Customization

```bash
# Custom HTML theme
goclean scan \
  --file-types .rs \
  --html \
  --theme custom \
  --css-file ./custom-styles.css

# Grouped reporting
goclean scan \
  --file-types .rs \
  --markdown \
  --group-by severity,file \
  --sort-by severity

# Statistical reporting
goclean scan \
  --file-types .rs \
  --format json \
  --include-stats \
  --include-metrics \
  --export-json
```

## Troubleshooting Commands

### Debugging Issues

```bash
# Verbose output for debugging
goclean scan \
  --file-types .rs \
  --verbose \
  --debug \
  --console

# Dry run to test configuration
goclean scan \
  --config configs/rust-strict.yaml \
  --dry-run \
  --verbose

# Check file discovery
goclean scan \
  --file-types .rs \
  --list-files \
  --verbose

# Debug specific files
goclean scan \
  --path ./src/problematic_file.rs \
  --debug \
  --verbose \
  --console
```

### Configuration Validation

```bash
# Validate configuration syntax
goclean config validate --config configs/rust-strict.yaml

# Test configuration with sample files
goclean scan \
  --config configs/rust-strict.yaml \
  --path ./testdata/rust \
  --verbose

# Show effective configuration
goclean config show --config configs/rust-strict.yaml

# Compare configurations
goclean config diff \
  --config1 configs/rust-minimal.yaml \
  --config2 configs/rust-strict.yaml
```

### Performance Analysis

```bash
# Memory profiling
goclean scan \
  --file-types .rs \
  --memory-profile \
  --profile-output memory.prof

# CPU profiling
goclean scan \
  --file-types .rs \
  --cpu-profile \
  --profile-output cpu.prof

# Benchmark comparison
goclean scan \
  --file-types .rs \
  --benchmark \
  --compare-baseline baseline-bench.json

# Resource monitoring
goclean scan \
  --file-types .rs \
  --monitor-resources \
  --resource-output resources.json
```

### Error Diagnosis

```bash
# Check clippy availability
goclean doctor --check-clippy

# Verify Rust toolchain
goclean doctor --check-rust

# Test parsing capabilities
goclean test-parse --file ./src/main.rs --verbose

# Validate output generation
goclean scan \
  --file-types .rs \
  --test-output \
  --html \
  --markdown \
  --dry-run
```

## Practical Workflows

### Development Workflow

```bash
# Daily development scan
goclean scan --config configs/rust-minimal.yaml --console --format table

# Pre-commit validation
goclean scan \
  --file-types .rs \
  --modified-since "1 hour ago" \
  --severity critical,high \
  --console

# Code review preparation
goclean scan \
  --file-types .rs \
  --html \
  --theme light \
  --markdown \
  --ai-friendly
```

### Release Workflow

```bash
# Pre-release quality check
goclean scan \
  --config configs/rust-strict.yaml \
  --html \
  --markdown \
  --export-json \
  --fail-on critical

# Release documentation
goclean scan \
  --config configs/rust-minimal.yaml \
  --markdown \
  --include-stats \
  --markdown-output QUALITY_REPORT.md

# Archive quality metrics
goclean scan \
  --config configs/rust-minimal.yaml \
  --export-json \
  --json-output "releases/quality-$(git describe --tags).json"
```

### Team Collaboration

```bash
# Team standards enforcement
goclean scan \
  --config team-standards.yaml \
  --html \
  --html-output shared/team-quality-report.html

# Onboarding scan for new developers
goclean scan \
  --config configs/rust-minimal.yaml \
  --console \
  --format table \
  --educational

# Code quality metrics for management
goclean scan \
  --config configs/rust-minimal.yaml \
  --format json \
  --include-metrics \
  --summary-only
```

This comprehensive CLI reference provides practical examples for every common GoClean Rust usage scenario. For additional help with specific commands, use `goclean help` or consult the main documentation.